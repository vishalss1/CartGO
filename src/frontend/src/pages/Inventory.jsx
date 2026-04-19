import { useEffect, useMemo, useState } from "react";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest, apiRequest } from "../utils/api";
import { useAuth } from "../context/AuthContext";
import { useToast } from "../context/ToastContext";
import { logger } from "../utils/logger";

export default function InventoryPage() {
  const { token, user } = useAuth();
  const { showSuccess, showError } = useToast();
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [updatingId, setUpdatingId] = useState("");
  const [creating, setCreating] = useState(false);
  const [draftProduct, setDraftProduct] = useState({
    name: "",
    description: "",
    price: "",
    category: "",
  });

  const isWarehouseStaff = user?.role === "WAREHOUSE_STAFF";

  async function loadRows(silent = false) {
    if (!silent) {
      setRows([]); // Clear potential ghost data
      setLoading(true);
    }

    try {
      const productList = await apiRequest("/products?limit=100&offset=0");
      const inventoryList = await Promise.all(
        (productList ?? []).map(async (product) => {
          try {
            const inventory = await apiRequest(`/inventory/${product.id}`);
            return { product, inventory };
          } catch (err) {
            logger.error(`[System Integrity] Could not fetch inventory for ${product.id}:`, err);
            return { product, inventory: null }; // No fallback data allowed
          }
        }),
      );

      // ── Traceability Logging (Dev-only) ──
      logger.info(`Inventory: API returned ${productList?.length || 0} base products. Enriched ${inventoryList.length} rows.`);

      setRows(inventoryList);
    } catch (loadError) {
      if (!silent) showError(loadError.message);
    } finally {
      if (!silent) setLoading(false);
    }
  }

  useEffect(() => {
    loadRows();
    const interval = setInterval(() => loadRows(true), 15000);
    return () => clearInterval(interval);
  }, []);

  const lowStockRows = useMemo(
    () => rows.filter((row) => row.inventory && row.inventory.available_stock <= 5),
    [rows],
  );

  async function adjustStock(productId, adjustment) {
    const originalRows = [...rows];
    setUpdatingId(productId);

    // Optimistic Update
    setRows((current) =>
      current.map((row) => {
        if (row.product.id === productId) {
          const nextAvailable = Math.max(0, row.inventory.available_stock + adjustment);
          const nextTotal = Math.max(0, row.inventory.total_stock + adjustment);
          return {
            ...row,
            inventory: {
              ...row.inventory,
              available_stock: nextAvailable,
              total_stock: nextTotal,
            },
          };
        }
        return row;
      }),
    );

    try {
      await authenticatedRequest(`/inventory/${productId}/adjust`, token, {
        method: "POST",
        body: JSON.stringify({ adjustment }),
      });
      showSuccess("Stock updated successfully.");
    } catch (adjustError) {
      setRows(originalRows); // Rollback
      showError(adjustError.message);
    } finally {
      setUpdatingId("");
    }
  }

  async function createProduct(event) {
    event.preventDefault();
    setCreating(true);

    try {
      await authenticatedRequest("/products", token, {
        method: "POST",
        body: JSON.stringify({
          ...draftProduct,
          price: Number(draftProduct.price),
        }),
      });
      setDraftProduct({ name: "", description: "", price: "", category: "" });
      showSuccess("Product created successfully.");
      await loadRows();
    } catch (createError) {
      showError(createError.message);
    } finally {
      setCreating(false);
    }
  }

  async function deactivateProduct(productId) {
    if (!window.confirm("Are you sure you want to deactivate this product? It will no longer be visible in the shop.")) {
      return;
    }
    setUpdatingId(productId);
    try {
      await authenticatedRequest(`/products/${productId}`, token, {
        method: "DELETE", // Backend implements this as soft-delete
      });
      showSuccess("Product deactivated.");
      await loadRows();
    } catch (removeError) {
      showError(removeError.message);
    } finally {
      setUpdatingId("");
    }
  }

  return (
    <MainLayout>
      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <section className="space-y-6">
            <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
                  Warehouse workspace
                </p>
                <h1 className="mt-4 text-[3.2rem] font-black uppercase leading-[0.9] tracking-hero sm:text-[4.8rem]">
                  Inventory
                </h1>
                <p className="mt-4 max-w-[34rem] text-sm leading-relaxed text-muted">
                  View current stock levels across all products. Adjust available inventory and
                  monitor low-stock alerts in real time.
                </p>
              </div>
              <button
                type="button"
                onClick={loadRows}
                disabled={loading}
                className="inline-flex w-fit border border-line px-4 py-3 text-[10px] font-bold uppercase tracking-[0.2em] transition-colors hover:border-paper disabled:cursor-not-allowed"
              >
                {loading ? "Syncing..." : "Refresh list"}
              </button>
            </div>

          <div className="grid gap-4">
            {loading && rows.length === 0 ? (
              <StatusBanner>Loading inventory</StatusBanner>
            ) : (
              rows.map((row) => (
                <article key={row.product.id} className="border border-line p-5">
                  <div className="grid gap-4 lg:grid-cols-[1fr_auto] lg:items-start">
                    <div>
                      <h2 className="text-[1.8rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                        {row.product.name}
                      </h2>
                      <p className="mt-2 text-sm text-muted">{row.product.category}</p>
                      <p className="mt-3 text-sm leading-relaxed text-muted">
                        {row.product.description}
                      </p>
                    </div>
                    <div className="grid grid-cols-3 gap-4 text-right text-sm uppercase tracking-[0.14em]">
                      <div>
                        <p className="text-muted">Available</p>
                        <p className="mt-2 text-xl font-semibold text-paper">
                          {row.inventory ? row.inventory.available_stock : "???"}
                        </p>
                      </div>
                      <div>
                        <p className="text-muted">Reserved</p>
                        <p className="mt-2 text-xl font-semibold text-paper">
                          {row.inventory ? row.inventory.reserved_stock : "???"}
                        </p>
                      </div>
                      <div>
                        <p className="text-muted">Total</p>
                        <p className="mt-2 text-xl font-semibold text-paper">
                          {row.inventory ? row.inventory.total_stock : "???"}
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className="mt-6 flex flex-wrap items-center justify-between gap-3">
                    <div className="flex gap-3">
                      <button
                        type="button"
                        onClick={() => adjustStock(row.product.id, 1)}
                        disabled={updatingId === row.product.id}
                        className="bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white disabled:cursor-not-allowed disabled:bg-line"
                      >
                        {updatingId === row.product.id && updatingId === row.product.id ? "Updating" : "+1 stock"}
                      </button>
                      <button
                        type="button"
                        onClick={() => adjustStock(row.product.id, -1)}
                        disabled={updatingId === row.product.id || !row.inventory || row.inventory.available_stock <= 0}
                        className="border border-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-paper transition-colors hover:border-white hover:text-white disabled:cursor-not-allowed disabled:border-line disabled:text-muted"
                      >
                        -1 stock
                      </button>
                    </div>

                    {isWarehouseStaff && (
                      <button
                        type="button"
                        onClick={() => deactivateProduct(row.product.id)}
                        disabled={updatingId === row.product.id}
                        className="border border-accent px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-accent transition-colors hover:bg-accent hover:text-surface disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        Deactivate
                      </button>
                    )}
                  </div>
                </article>
              ))
            )}
          </div>

        </section>

        <aside className="space-y-6">

          {isWarehouseStaff && (
            <section className="border border-line p-6">
              <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                Add Product
              </h2>
              <form onSubmit={createProduct} className="mt-6 grid gap-4">
                <input
                  value={draftProduct.name}
                  onChange={(event) =>
                    setDraftProduct((current) => ({ ...current, name: event.target.value }))
                  }
                  placeholder="Name"
                  required
                  className="w-full border border-line bg-transparent px-4 py-4 outline-none focus:border-paper"
                />
                <input
                  value={draftProduct.category}
                  onChange={(event) =>
                    setDraftProduct((current) => ({ ...current, category: event.target.value }))
                  }
                  placeholder="Category"
                  required
                  className="w-full border border-line bg-transparent px-4 py-4 outline-none focus:border-paper"
                />
                <input
                  value={draftProduct.price}
                  onChange={(event) =>
                    setDraftProduct((current) => ({ ...current, price: event.target.value }))
                  }
                  placeholder="Price"
                  type="number"
                  step="0.01"
                  required
                  className="w-full border border-line bg-transparent px-4 py-4 outline-none focus:border-paper"
                />
                <textarea
                  value={draftProduct.description}
                  onChange={(event) =>
                    setDraftProduct((current) => ({
                      ...current,
                      description: event.target.value,
                    }))
                  }
                  placeholder="Description"
                  required
                  className="h-24 w-full border border-line bg-transparent px-4 py-4 outline-none focus:border-paper"
                />
                <button
                  type="submit"
                  disabled={creating}
                  className="bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white disabled:cursor-not-allowed"
                >
                  {creating ? "Creating..." : "Create product"}
                </button>
              </form>
            </section>
          )}
        </aside>
      </div>
    </MainLayout>
  );
}
