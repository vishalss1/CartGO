import { useEffect, useMemo, useState } from "react";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest, apiRequest } from "../utils/api";
import { VALID_ROLES } from "../utils/constants";
import { useAuth } from "../context/AuthContext";
import { useToast } from "../context/ToastContext";

export default function AdminPage() {
  const { token } = useAuth();
  const { showSuccess, showError } = useToast();
  const [users, setUsers] = useState([]);
  const [products, setProducts] = useState([]);
  const [selectedUserId, setSelectedUserId] = useState("");
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [ordersLoading, setOrdersLoading] = useState(false);
  const [draftProduct, setDraftProduct] = useState({
    name: "",
    description: "",
    price: "",
    category: "",
  });

  async function loadAdminData() {
    setLoading(true);

    try {
      const [userList, productList] = await Promise.all([
        authenticatedRequest("/user/admin/users", token),
        apiRequest("/products?limit=100&offset=0"),
      ]);
      setUsers(Array.isArray(userList) ? userList : []);
      setProducts(Array.isArray(productList) ? productList : []);
    } catch (loadError) {
      showError(loadError.message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadAdminData();
  }, [token]);

  const usersByRole = useMemo(() => {
    return VALID_ROLES.map((role) => ({
      role,
      count: users.filter((user) => user.role === role).length,
    }));
  }, [users]);

  async function updateRole(userId, role) {
    try {
      await authenticatedRequest(`/user/admin/users/${userId}/role`, token, {
        method: "PATCH",
        body: JSON.stringify({ role }),
      });
      showSuccess("User role updated.");
      await loadAdminData();
    } catch (updateError) {
      showError(updateError.message);
    }
  }

  async function loadOrdersForUser() {
    if (!selectedUserId) {
      return;
    }
    setOrdersLoading(true);

    try {
      const userOrders = await authenticatedRequest(`/orders/user/${selectedUserId}`, token);
      setOrders(Array.isArray(userOrders) ? userOrders : []);
    } catch (loadError) {
      showError(loadError.message);
    } finally {
      setOrdersLoading(false);
    }
  }

  async function createProduct(event) {
    event.preventDefault();

    try {
      await authenticatedRequest("/products", token, {
        method: "POST",
        body: JSON.stringify({
          ...draftProduct,
          price: Number(draftProduct.price),
        }),
      });
      setDraftProduct({ name: "", description: "", price: "", category: "" });
      showSuccess("Product created.");
      await loadAdminData();
    } catch (createError) {
      showError(createError.message);
    }
  }

  async function removeProduct(productId) {
    try {
      await authenticatedRequest(`/products/${productId}`, token, {
        method: "DELETE",
      });
      showSuccess("Product removed.");
      await loadAdminData();
    } catch (removeError) {
      showError(removeError.message);
    }
  }

  return (
    <MainLayout>
      <div className="space-y-6">
        <section className="border border-line p-6 sm:p-8">
          <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
            Admin workspace
          </p>
          <h1 className="mt-4 text-[3.2rem] font-black uppercase leading-[0.9] tracking-hero sm:text-[4.8rem]">
            Admin
          </h1>
          <p className="mt-4 max-w-[34rem] text-sm leading-relaxed text-muted">
            Manage users, products, and review orders. Assign roles and oversee platform
            operations from a single dashboard.
          </p>
        </section>

        <section className="grid gap-4 md:grid-cols-3">
          <article className="border border-line p-5">
            <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">Users</p>
            <p className="mt-4 text-[2.4rem] font-black tracking-hero">{users.length}</p>
          </article>
          <article className="border border-line p-5">
            <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">
              Products
            </p>
            <p className="mt-4 text-[2.4rem] font-black tracking-hero">{products.length}</p>
          </article>
          <article className="border border-line p-5">
            <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">
              Roles
            </p>
            <p className="mt-4 text-[2.4rem] font-black tracking-hero">{VALID_ROLES.length}</p>
          </article>
        </section>

        <section className="grid gap-6 xl:grid-cols-[1fr_0.9fr]">
          <div className="border border-line p-6">
            <div className="flex items-end justify-between gap-4">
              <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                Users
              </h2>
              {loading ? <span className="text-sm text-muted">Loading</span> : null}
            </div>
            <div className="mt-6 space-y-4">
              {users.map((user) => (
                <div key={user.id} className="border border-line p-4">
                  <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                    <div>
                      <p className="font-semibold uppercase tracking-[0.08em]">{user.username}</p>
                      <p className="mt-1 text-sm text-muted">{user.email}</p>
                    </div>
                    <select
                      value={user.role}
                      onChange={(event) => updateRole(user.id, event.target.value)}
                      className="border border-line bg-surface px-4 py-3 text-sm uppercase tracking-[0.16em] outline-none focus:border-paper"
                    >
                      {VALID_ROLES.map((role) => (
                        <option key={role} value={role}>
                          {role}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="space-y-6">
            <section className="border border-line p-6">
              <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                Products
              </h2>
              <form onSubmit={createProduct} className="mt-6 grid gap-4">
                <input
                  value={draftProduct.name}
                  onChange={(event) =>
                    setDraftProduct((current) => ({ ...current, name: event.target.value }))
                  }
                  placeholder="Name"
                  className="w-full border border-line bg-transparent px-4 py-4 outline-none focus:border-paper"
                />
                <input
                  value={draftProduct.category}
                  onChange={(event) =>
                    setDraftProduct((current) => ({ ...current, category: event.target.value }))
                  }
                  placeholder="Category"
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
                  className="h-24 w-full border border-line bg-transparent px-4 py-4 outline-none focus:border-paper"
                />
                <button
                  type="submit"
                  className="inline-flex w-fit bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white"
                >
                  Create product
                </button>
              </form>
              <div className="mt-6 space-y-3">
                {products.slice(0, 8).map((product) => (
                  <div key={product.id} className="flex items-center justify-between gap-4 border border-line p-4">
                    <div>
                      <p className="font-semibold uppercase tracking-[0.08em]">{product.name}</p>
                      <p className="mt-1 text-sm text-muted">{product.category}</p>
                    </div>
                    <button
                      type="button"
                      onClick={() => removeProduct(product.id)}
                      className="border border-accent px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-accent transition-colors hover:bg-accent hover:text-surface"
                    >
                      Delete
                    </button>
                  </div>
                ))}
              </div>
            </section>

            <section className="border border-line p-6">
              <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                Role totals
              </h2>
              <div className="mt-6 grid gap-4 sm:grid-cols-2">
                {usersByRole.map((entry) => (
                  <div key={entry.role} className="border border-line p-4">
                    <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-muted">
                      {entry.role}
                    </p>
                    <p className="mt-3 text-[2rem] font-black tracking-hero">{entry.count}</p>
                  </div>
                ))}
              </div>
            </section>

            <section className="border border-line p-6">
              <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                Orders
              </h2>
              <div className="mt-6 flex flex-col gap-4">
                <select
                  value={selectedUserId}
                  onChange={(event) => setSelectedUserId(event.target.value)}
                  className="border border-line bg-surface px-4 py-4 text-sm uppercase tracking-[0.16em] outline-none focus:border-paper"
                >
                  <option value="">Select user</option>
                  {users.map((user) => (
                    <option key={user.id} value={user.id}>
                      {user.email}
                    </option>
                  ))}
                </select>
                <button
                  type="button"
                  onClick={loadOrdersForUser}
                  disabled={!selectedUserId || ordersLoading}
                  className="inline-flex w-fit bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white disabled:cursor-not-allowed disabled:bg-line"
                >
                  {ordersLoading ? "Loading orders" : "Load user orders"}
                </button>
              </div>
              <div className="mt-6 space-y-4">
                {orders.map((order) => (
                  <div key={order.id} className="border border-line p-4">
                    <p className="font-semibold uppercase tracking-[0.08em]">{order.status}</p>
                    <p className="mt-2 text-sm text-muted">{order.id}</p>
                    <p className="mt-2 text-sm text-muted">
                      Amount: ${order.total_amount.toFixed(2)}
                    </p>
                  </div>
                ))}
              </div>
            </section>
          </div>
        </section>
      </div>
    </MainLayout>
  );
}
