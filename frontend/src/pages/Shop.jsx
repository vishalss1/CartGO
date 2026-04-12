import { useEffect, useMemo, useState } from "react";
import MainLayout from "../layout/MainLayout";
import { getStoredCart, storeCart } from "../utils/auth";
import { authenticatedRequest, apiRequest } from "../utils/api";
import { useAuth } from "../context/AuthContext";
import StatusBanner from "../components/StatusBanner";

export default function ShopPage() {
  const { token, user } = useAuth();
  const [products, setProducts] = useState([]);
  const [orders, setOrders] = useState([]);
  const [cart, setCart] = useState(getStoredCart);
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState("");
  const [deliveryAddress, setDeliveryAddress] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(true);
  const [checkingOut, setCheckingOut] = useState(false);
  const [payingOrderId, setPayingOrderId] = useState("");

  useEffect(() => {
    storeCart(cart);
  }, [cart]);

  useEffect(() => {
    let active = true;

    async function loadData() {
      try {
        const [productList, orderList] = await Promise.all([
          apiRequest("/products?limit=100&offset=0"),
          authenticatedRequest(`/orders/user/${user.id}`, token),
        ]);
        if (!active) {
          return;
        }
        setProducts(Array.isArray(productList) ? productList : []);
        setOrders(Array.isArray(orderList) ? orderList : []);
      } catch (loadError) {
        if (active) {
          setError(loadError.message);
        }
      } finally {
        if (active) {
          setLoading(false);
        }
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
        .map((item) => (item.product_id === productId ? { ...item, quantity } : item))
        .filter((item) => item.quantity > 0),
    );
  }

  async function handleCheckout() {
    setCheckingOut(true);
    setError("");
    setSuccess("");

    try {
      const order = await authenticatedRequest("/orders", token, {
        method: "POST",
        body: JSON.stringify({
          items: cart.map((item) => ({
            product_id: item.product_id,
            quantity: item.quantity,
          })),
          delivery_address: deliveryAddress,
        }),
      });
      const freshOrders = await authenticatedRequest(`/orders/user/${user.id}`, token);
      setOrders(Array.isArray(freshOrders) ? freshOrders : []);
      setCart([]);
      storeCart([]);
      setDeliveryAddress("");
      setSuccess(`Order ${order.id} created with status ${order.status}.`);
    } catch (checkoutError) {
      setError(checkoutError.message);
    } finally {
      setCheckingOut(false);
    }
  }

  async function retryPayment(order) {
    setPayingOrderId(order.id);
    setError("");
    setSuccess("");

    try {
      const payment = await authenticatedRequest("/payments", token, {
        method: "POST",
        body: JSON.stringify({
          order_id: order.id,
          amount: order.total_amount,
        }),
      });
      setSuccess(`Payment ${payment.payment_id} returned ${payment.status}.`);
      const freshOrders = await authenticatedRequest(`/orders/user/${user.id}`, token);
      setOrders(Array.isArray(freshOrders) ? freshOrders : []);
    } catch (paymentError) {
      setError(paymentError.message);
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
              Product catalog is loaded from product-service through the API gateway. Checkout
              submits to order-service, and payment retries call payment-service directly.
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

          {error ? <StatusBanner tone="error">{error}</StatusBanner> : null}
          {success ? <StatusBanner tone="success">{success}</StatusBanner> : null}

          <div className="grid gap-4 md:grid-cols-2">
            {loading ? (
              <StatusBanner>Loading products from product-service</StatusBanner>
            ) : (
              filteredProducts.map((product) => (
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
                  <button
                    type="button"
                    onClick={() => addToCart(product)}
                    className="mt-6 inline-flex bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white"
                  >
                    Add to cart
                  </button>
                </article>
              ))
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
                placeholder="Required by order-service"
              />
              <div className="mt-4 flex items-center justify-between">
                <span className="text-sm text-muted">Total</span>
                <span className="text-xl font-semibold">${cartTotal.toFixed(2)}</span>
              </div>
              <button
                type="button"
                onClick={handleCheckout}
                disabled={checkingOut || cart.length === 0 || !deliveryAddress}
                className="mt-5 inline-flex bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white disabled:cursor-not-allowed disabled:bg-line"
              >
                {checkingOut ? "Submitting order" : "Checkout"}
              </button>
            </div>
          </section>

          <section className="border border-line p-6">
            <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
              Orders
            </h2>
            <div className="mt-6 space-y-4">
              {orders.length === 0 ? (
                <p className="text-sm text-muted">No orders returned for this user.</p>
              ) : (
                orders.map((order) => (
                  <div key={order.id} className="border border-line p-4">
                    <div className="flex items-center justify-between gap-4">
                      <div>
                        <p className="font-semibold uppercase tracking-[0.08em]">{order.status}</p>
                        <p className="text-sm text-muted">{order.id}</p>
                      </div>
                      <span className="text-lg font-semibold">${order.total_amount.toFixed(2)}</span>
                    </div>
                    {order.status === "FAILED" ? (
                      <button
                        type="button"
                        onClick={() => retryPayment(order)}
                        disabled={payingOrderId === order.id}
                        className="mt-4 inline-flex bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white disabled:cursor-not-allowed disabled:bg-line"
                      >
                        {payingOrderId === order.id ? "Processing payment" : "Retry payment"}
                      </button>
                    ) : null}
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
