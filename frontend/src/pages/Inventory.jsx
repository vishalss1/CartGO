import { useEffect, useMemo, useState } from "react";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest, apiRequest } from "../utils/api";
import { useAuth } from "../context/AuthContext";

export default function InventoryPage() {
  const { token } = useAuth();
  const [rows, setRows] = useState([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(true);

  async function loadRows() {
    setLoading(true);
    setError("");

    try {
      const productList = await apiRequest("/products?limit=100&offset=0");
      const inventoryList = await Promise.all(
        (productList ?? []).map(async (product) => {
          try {
            const inventory = await apiRequest(`/inventory/${product.id}`);
            return { product, inventory };
          } catch {
            return {
              product,
              inventory: {
                available_stock: 0,
                total_stock: 0,
                reserved_stock: 0,
              },
            };
          }
        }),
      );
      setRows(inventoryList);
    } catch (loadError) {
      setError(loadError.message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadRows();
  }, []);

  const lowStockRows = useMemo(
    () => rows.filter((row) => row.inventory.available_stock <= 5),
    [rows],
  );

  async function adjustStock(productId, adjustment) {
    setError("");
    setSuccess("");

    try {
      await authenticatedRequest(`/inventory/${productId}/adjust`, token, {
        method: "POST",
        body: JSON.stringify({ adjustment }),
      });
      setSuccess("Inventory updated from inventory-service.");
      await loadRows();
    } catch (adjustError) {
      setError(adjustError.message);
    }
  }

  return (
    <MainLayout>
      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <section className="space-y-6">
          <div className="border border-line p-6 sm:p-8">
            <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
              Warehouse workspace
            </p>
            <h1 className="mt-4 text-[3.2rem] font-black uppercase leading-[0.9] tracking-hero sm:text-[4.8rem]">
              Inventory
            </h1>
            <p className="mt-4 max-w-[34rem] text-sm leading-relaxed text-muted">
              Product data comes from product-service. Stock adjustments hit inventory-service.
              Product creation and deletion are enforced by backend RBAC as ADMIN-only mutations,
              so this warehouse route focuses on live stock operations.
            </p>
          </div>

          {error ? <StatusBanner tone="error">{error}</StatusBanner> : null}
          {success ? <StatusBanner tone="success">{success}</StatusBanner> : null}

          <div className="grid gap-4">
            {loading ? (
              <StatusBanner>Loading product and inventory data</StatusBanner>
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
                          {row.inventory.available_stock}
                        </p>
                      </div>
                      <div>
                        <p className="text-muted">Reserved</p>
                        <p className="mt-2 text-xl font-semibold text-paper">
                          {row.inventory.reserved_stock}
                        </p>
                      </div>
                      <div>
                        <p className="text-muted">Total</p>
                        <p className="mt-2 text-xl font-semibold text-paper">
                          {row.inventory.total_stock}
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className="mt-6 flex flex-wrap gap-3">
                    <button
                      type="button"
                      onClick={() => adjustStock(row.product.id, 1)}
                      className="bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white"
                    >
                      +1 stock
                    </button>
                    <button
                      type="button"
                      onClick={() => adjustStock(row.product.id, -1)}
                      className="border border-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-paper transition-colors hover:border-white hover:text-white"
                    >
                      -1 stock
                    </button>
                  </div>
                </article>
              ))
            )}
          </div>
        </section>

        <aside className="space-y-6">
          <section className="border border-line p-6">
            <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
              Alerts
            </h2>
            <div className="mt-6 space-y-4">
              {lowStockRows.length === 0 ? (
                <p className="text-sm text-muted">No low stock rows detected.</p>
              ) : (
                lowStockRows.map((row) => (
                  <div key={row.product.id} className="border border-line p-4">
                    <p className="font-semibold uppercase tracking-[0.08em]">{row.product.name}</p>
                    <p className="mt-2 text-sm text-muted">
                      Available stock: {row.inventory.available_stock}
                    </p>
                  </div>
                ))
              )}
            </div>
          </section>

          <section className="border border-line p-6">
            <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
              RBAC
            </h2>
            <p className="mt-6 text-sm leading-relaxed text-muted">
              Verified from service middleware: `WAREHOUSE_STAFF` can adjust stock at
              `/inventory/:product_id/adjust`, while product creation and deletion remain
              `ADMIN`-only on product-service.
            </p>
          </section>
        </aside>
      </div>
    </MainLayout>
  );
}
