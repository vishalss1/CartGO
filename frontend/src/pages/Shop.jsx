import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import MainLayout from "../layout/MainLayout";
import { getStoredCart, storeCart } from "../utils/auth";
import { authenticatedRequest, apiRequest } from "../utils/api";
import { useAuth } from "../context/AuthContext";
import { useToast } from "../context/ToastContext";
import StatusBanner from "../components/StatusBanner";
import { logger } from "../utils/logger";

function OrderDeliveryStatus({ orderId, token }) {
  const [delivery, setDelivery] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    async function fetchStatus() {
      try {
        const data = await authenticatedRequest(`/deliveries/order/${orderId}`, token);
        if (active) setDelivery(data);
      } catch {
        if (active) setDelivery(null);
      } finally {
        if (active) setLoading(false);
      }
    }
    fetchStatus();
    // Poll delivery status every 15s
    const interval = setInterval(fetchStatus, 15000);
    return () => {
      active = false;
      clearInterval(interval);
    };
  }, [orderId, token]);

  if (loading) return <span className="text-[10px] uppercase tracking-widest text-muted">Tracking...</span>;
  if (!delivery) return <span className="text-[10px] uppercase tracking-widest text-muted">Preparing</span>;

  const statusColors = {
    PENDING: "text-muted",
    PICKED_UP: "text-accent",
    DELIVERED: "text-paper",
    CANCELLED: "text-red-500",
  };

  return (
    <span className={`text-[10px] font-bold uppercase tracking-widest ${statusColors[delivery.status] || "text-muted"}`}>
      {delivery.status.replace("_", " ")}
    </span>
  );
}

export default function ShopPage() {
  const { token, user } = useAuth();
  const { showSuccess, showError } = useToast();
  const navigate = useNavigate();
  const [products, setProducts] = useState([]);
  const [orders, setOrders] = useState([]);
  const [cart, setCart] = useState(getStoredCart);
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState("");
  const [deliveryAddress, setDeliveryAddress] = useState("");
  const [loading, setLoading] = useState(true);
  const [refreshingOrders, setRefreshingOrders] = useState(false);
  const [payingOrderId, setPayingOrderId] = useState("");

  const loadOrders = async (silent = false) => {
    if (!silent) setRefreshingOrders(true);
    try {
      const freshOrders = await authenticatedRequest(`/orders/user/${user.id}`, token);
      setOrders(Array.isArray(freshOrders) ? freshOrders : []);
    } catch (err) {
      if (!silent) showError(err.message);
    } finally {
      if (!silent) setRefreshingOrders(false);
    }
  };

  useEffect(() => {
    storeCart(cart);
  }, [cart]);

  // Hybrid Sync: Polling every 15s
  useEffect(() => {
    const interval = setInterval(() => loadOrders(true), 15000);
    return () => clearInterval(interval);
  }, [token, user.id]);

  useEffect(() => {
    let active = true;

    async function loadData() {
      // 1. Reset state (Guarantee clean start, avoid ghosting)
      setProducts([]);
      setLoading(true);

      try {
        const [productList, orderList] = await Promise.all([
          apiRequest("/products?limit=100&offset=0"),
          authenticatedRequest(`/orders/user/${user.id}`, token),
        ]);

        if (!active) return;

        // ── Option B: Manual Merge with Inventory ──
        const enriched = await Promise.all(
          (productList ?? []).map(async (p) => {
            // 1. Strict Validation Guard (Log & Skip Invalid)
            if (!p.id || !p.name || typeof p.price !== "number") {
              logger.error("[System Integrity] Malformed product record filtered:", p);
              return null;
            }

            // 2. Anti-Regression Guard
            const lowerName = p.name.toLowerCase();
            if (lowerName.includes("test") || lowerName.includes("demo") || lowerName.includes("sample")) {
              logger.warn(`[Anti-Regression] DEMO DATA DETECTED IN CATALOG: "${p.name}" (ID: ${p.id}). Please purge the database.`);
            }

            try {
              const inv = await apiRequest(`/inventory/${p.id}`);
              return { ...p, inventory: inv };
            } catch {
              logger.error(`[System Integrity] Missing inventory for product ${p.id}`);
              return { ...p, inventory: null }; // Mark as invalid state
            }
          })
        );

        const finalProducts = enriched.filter(Boolean);
        
        // ── Traceability Logging (Dev-only) ──
        logger.info(`Shop: API returned ${productList?.length || 0} products. Rendered ${finalProducts.length} valid items.`);

        setProducts(finalProducts);
        setOrders(Array.isArray(orderList) ? orderList : []);
      } catch (loadError) {
        if (active) showError(loadError.message);
      } finally {
        if (active) setLoading(false);
      }
    }

    loadData();
    return () => {
      active = false;
    };
  }, [token, user.id]);

  const categories = useMemo(
    () => [...new Set(products.map((product) => product.category).filter(Boolean))],
    [products],
  );

  const filteredProducts = useMemo(() => {
    return products.filter((product) => {
      const matchesSearch =
        !search ||
        `${product.name} ${product.description}`.toLowerCase().includes(search.toLowerCase());
      const matchesCategory = !category || product.category === category;
      return matchesSearch && matchesCategory;
    });
  }, [category, products, search]);

  function addToCart(product) {
    setCart((current) => {
      const existing = current.find((item) => item.product_id === product.id);
      if (existing) {
        return current.map((item) =>
          item.product_id === product.id ? { ...item, quantity: item.quantity + 1 } : item,
        );
      }
      return [
        ...current,
        {
          product_id: product.id,
          name: product.name,
          price: product.price,
          quantity: 1,
        },
      ];
    });
  }

  function updateQuantity(productId, quantity) {
    setCart((current) =>
      current
        .map((item) => (item.product_id === productId ? { ...item, quantity: Math.max(0, quantity) } : item))
        .filter((item) => item.quantity > 0),
    );
  }

  async function handleCheckout() {
    if (cart.length === 0 || !deliveryAddress) {
      return;
    }
    // Redirect to dedicated payment page
    navigate("/checkout/payment", {
      state: { cart, deliveryAddress }
    });
  }

  async function retryPayment(order) {
    setPayingOrderId(order.id);

    try {
      // 1. Process Payment
      await authenticatedRequest("/payments", token, {
        method: "POST",
        body: JSON.stringify({
          order_id: order.id,
          amount: order.total_amount,
        }),
      });

      showSuccess("Payment successful. Finalizing order...");

      // 2. Synchronize Order Fulfillment
      await authenticatedRequest(`/orders/${order.id}/confirm-after-payment`, token, {
        method: "POST"
      });

      showSuccess("Order confirmed and fulfillment triggered.");
      await loadOrders();
    } catch (paymentError) {
      showError(paymentError.message);
    } finally {
      setPayingOrderId("");
    }
  }

  const cartTotal = cart.reduce((sum, item) => sum + item.price * item.quantity, 0);

  return (
    <MainLayout>
      <div className="grid gap-6 xl:grid-cols-[1.25fr_0.75fr]">
        <section className="space-y-6">
          <div className="border border-line p-6 sm:p-8">
            <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
              Customer workspace
            </p>
            <h1 className="mt-4 text-[3.2rem] font-black uppercase leading-[0.9] tracking-hero sm:text-[4.8rem]">
              Shop
            </h1>
            <p className="mt-4 max-w-[34rem] text-sm leading-relaxed text-muted">
              Browse the catalog, add items to your cart, and place orders. Track your order
              history and payment status below.
            </p>
          </div>

          <div className="grid gap-4 md:grid-cols-[1fr_220px]">
            <input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search products"
              className="border border-line bg-transparent px-4 py-4 text-base outline-none transition-colors focus:border-paper"
            />
            <select
              value={category}
              onChange={(event) => setCategory(event.target.value)}
              className="border border-line bg-surface px-4 py-4 text-base outline-none transition-colors focus:border-paper"
            >
              <option value="">All categories</option>
              {categories.map((item) => (
                <option key={item} value={item}>
                  {item}
                </option>
              ))}
            </select>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            {loading ? (
              <StatusBanner>Loading products</StatusBanner>
            ) : filteredProducts.length === 0 ? (
              <div className="md:col-span-2 border border-line border-dashed p-12 text-center">
                 <p className="text-sm font-black uppercase tracking-[0.2em] text-muted">No products available in the catalog</p>
                 <p className="mt-2 text-[10px] text-muted uppercase tracking-widest">Add products via the Admin Dashboard to see them here.</p>
              </div>
            ) : (
              filteredProducts.map((product) => {
                const isOutOfStock = product.inventory && product.inventory.available_stock <= 0;
                const isInvalid = !product.inventory;

                return (
                  <article key={product.id} className="border border-line p-5">
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <h2 className="text-[1.8rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                          {product.name}
                        </h2>
                        <p className="mt-2 text-[11px] font-semibold uppercase tracking-[0.26em] text-muted">
                          {product.category}
                        </p>
                      </div>
                      <span className="text-lg font-semibold text-paper">
                        ${product.price.toFixed(2)}
                      </span>
                    </div>
                    <p className="mt-5 text-sm leading-relaxed text-muted">{product.description}</p>
                    
                    <div className="mt-6 flex items-center justify-between">
                      {isInvalid ? (
                        <span className="text-[10px] font-black uppercase tracking-widest text-red-500">System Error: No Inventory</span>
                      ) : isOutOfStock ? (
                        <span className="text-[10px] font-black uppercase tracking-widest text-muted">Out of Stock</span>
                      ) : (
                        <button
                          type="button"
                          onClick={() => addToCart(product)}
                          className="inline-flex bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white"
                        >
                          Add to cart
                        </button>
                      )}
                    </div>
                  </article>
                );
              })
            )}
          </div>
        </section>

        <aside className="space-y-6">
          <section className="border border-line p-6">
            <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
              Cart
            </h2>
            <div className="mt-6 space-y-4">
              {cart.length === 0 ? (
                <p className="text-sm text-muted">No items added yet.</p>
              ) : (
                cart.map((item) => (
                  <div key={item.product_id} className="border border-line p-4">
                    <div className="flex items-center justify-between gap-4">
                      <div>
                        <p className="font-semibold uppercase tracking-[0.08em]">{item.name}</p>
                        <p className="text-sm text-muted">${item.price.toFixed(2)} each</p>
                      </div>
                      <input
                        type="number"
                        min="0"
                        value={item.quantity}
                        onChange={(event) =>
                          updateQuantity(item.product_id, Number(event.target.value))
                        }
                        className="w-20 border border-line bg-transparent px-3 py-2 text-right outline-none focus:border-paper"
                      />
                    </div>
                  </div>
                ))
              )}
            </div>
            <div className="mt-6 border-t border-line pt-5">
              <p className="text-sm uppercase tracking-[0.18em] text-muted">Delivery address</p>
              <textarea
                value={deliveryAddress}
                onChange={(event) => setDeliveryAddress(event.target.value)}
                className="mt-3 h-28 w-full border border-line bg-transparent px-4 py-4 text-base outline-none transition-colors focus:border-paper"
                placeholder="Enter your delivery address"
              />
              <div className="mt-4 flex items-center justify-between">
                <span className="text-sm text-muted">Total</span>
                <span className="text-xl font-semibold">${cartTotal.toFixed(2)}</span>
              </div>
              <button
                type="button"
                onClick={handleCheckout}
                disabled={cart.length === 0 || !deliveryAddress}
                className="mt-5 inline-flex w-full justify-center bg-paper px-4 py-4 text-xs font-black uppercase tracking-[0.2em] text-surface transition-colors hover:bg-white disabled:cursor-not-allowed disabled:bg-line"
              >
                Go to Checkout
              </button>
            </div>
          </section>

          <section className="border border-line p-6">
            <div className="flex items-end justify-between gap-4">
              <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                Orders
              </h2>
              <button
                type="button"
                onClick={() => loadOrders()}
                disabled={refreshingOrders}
                className="text-[10px] font-bold uppercase tracking-[0.2em] text-muted underline transition-colors hover:text-paper disabled:cursor-not-allowed"
              >
                {refreshingOrders ? "Syncing" : "Refresh"}
              </button>
            </div>
            <div className="mt-6 space-y-4">
              {orders.length === 0 ? (
                <p className="text-sm text-muted">No orders yet.</p>
              ) : (
                orders.map((order) => (
                  <div key={order.id} className="border border-line p-4 transition-all hover:border-paper">
                    <div className="flex items-center justify-between gap-4">
                      <div className="space-y-1">
                        <div className="flex items-center gap-3">
                          <p onClick={() => navigate(`/order/${order.id}`)} className="cursor-pointer font-semibold uppercase tracking-[0.08em] hover:text-paper hover:underline">{order.status}</p>
                          <OrderDeliveryStatus orderId={order.id} token={token} />
                        </div>
                        <p onClick={() => navigate(`/order/${order.id}`)} className="cursor-pointer text-[10px] text-muted hover:text-paper">{order.id}</p>
                      </div>
                      <span className="text-lg font-semibold">${order.total_amount.toFixed(2)}</span>
                    </div>
                    
                    <div className="mt-4 flex flex-wrap gap-3">
                      <button
                        type="button"
                        onClick={() => navigate(`/order/${order.id}`)}
                        className="border border-paper border-opacity-30 bg-paper bg-opacity-5 px-4 py-2 text-[10px] font-bold uppercase tracking-[0.16em] text-paper transition-all hover:bg-opacity-20"
                      >
                        View System Trace
                      </button>
                      {order.status === "FAILED" && (
                        <button
                          type="button"
                          onClick={() => retryPayment(order)}
                          disabled={payingOrderId === order.id}
                          className="bg-paper px-4 py-2 text-[10px] font-bold uppercase tracking-[0.16em] text-surface transition-colors hover:bg-white disabled:cursor-not-allowed"
                        >
                          {payingOrderId === order.id ? "Processing" : "Retry payment"}
                        </button>
                      )}
                      <button
                        type="button"
                        onClick={() => navigate(`/support?order_id=${order.id}`)}
                        className="border border-line px-4 py-2 text-[10px] font-bold uppercase tracking-[0.16em] text-muted transition-colors hover:border-paper hover:text-paper"
                      >
                        Report Issue
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>
          </section>
        </aside>
      </div>
    </MainLayout>
  );
}
