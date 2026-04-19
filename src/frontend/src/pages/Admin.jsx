import { useEffect, useMemo, useState } from "react";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest, apiRequest } from "../utils/api";
import { VALID_ROLES } from "../utils/constants";
import { useAuth } from "../context/AuthContext";
import { useToast } from "../context/ToastContext";
import { logger } from "../utils/logger";

export default function AdminPage() {
  const { token } = useAuth();
  const { showSuccess, showError } = useToast();
  const [users, setUsers] = useState([]);
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [ordersLoading, setOrdersLoading] = useState(false);
  const [updatingUserId, setUpdatingUserId] = useState("");
  const [orderFilters, setOrderFilters] = useState({
    status: "",
    userId: "",
    page: 1,
    limit: 10,
  });
  const [totalOrders, setTotalOrders] = useState(0);

  async function loadAdminData() {
    setUsers([]);
    setLoading(true);

    try {
      const userList = await authenticatedRequest("/user/admin/users", token);
      const validUsers = Array.isArray(userList) ? userList : [];

      logger.info(`Admin: API sync complete. Users: ${validUsers.length}`);
      setUsers(validUsers);
    } catch (loadError) {
      showError(loadError.message);
      logger.error("Admin Load Error:", loadError);
    } finally {
      setLoading(false);
    }
  }

  async function loadGlobalOrders() {
    setOrdersLoading(true);
    try {
      const { status, userId, page, limit } = orderFilters;
      let url = `/orders?page=${page}&limit=${limit}`;
      if (status) url += `&status=${status}`;
      if (userId) url += `&user_id=${userId}`;

      const response = await authenticatedRequest(url, token);
      // Response format: { data, total, page, limit }
      setOrders(response.data || []);
      setTotalOrders(response.total || 0);
    } catch (err) {
      showError("Failed to fetch global orders");
    } finally {
      setOrdersLoading(false);
    }
  }

  useEffect(() => {
    loadAdminData();
  }, [token]);

  useEffect(() => {
    loadGlobalOrders();
  }, [token, orderFilters]);

  const totalPages = Math.ceil(totalOrders / orderFilters.limit);

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

        <section className="grid gap-4 md:grid-cols-2">
          <article className="border border-line p-5">
            <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">Users</p>
            <p className="mt-4 text-[2.4rem] font-black tracking-hero">{users.length}</p>
          </article>
          <article className="border border-line p-5">
            <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">
              Global Orders
            </p>
            <p className="mt-4 text-[2.4rem] font-black tracking-hero">{totalOrders}</p>
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
                      disabled={updatingUserId === user.id}
                      onChange={(event) => updateRole(user.id, event.target.value)}
                      className="border border-line bg-surface px-4 py-3 text-sm uppercase tracking-[0.16em] outline-none focus:border-paper disabled:cursor-not-allowed disabled:bg-line"
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
              <div className="flex items-end justify-between gap-4">
                <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                  All Orders
                </h2>
                {ordersLoading ? <span className="text-sm text-muted animate-pulse">Syncing</span> : null}
              </div>

              {/* Filters */}
              <div className="mt-6 flex flex-wrap gap-4">
                <select
                  value={orderFilters.status}
                  onChange={(e) => setOrderFilters(curr => ({ ...curr, status: e.target.value, page: 1 }))}
                  className="border border-line bg-surface px-4 py-3 text-xs uppercase tracking-[0.14em]"
                >
                  <option value="">All Statuses</option>
                  <option value="PENDING">Pending</option>
                  <option value="CONFIRMED">Confirmed</option>
                  <option value="FAILED">Failed</option>
                </select>

                <select
                   value={orderFilters.userId}
                   onChange={(e) => setOrderFilters(curr => ({ ...curr, userId: e.target.value, page: 1 }))}
                   className="border border-line bg-surface px-4 py-3 text-xs uppercase tracking-[0.14em] max-w-[200px]"
                >
                  <option value="">By User</option>
                  {users.map(u => (
                    <option key={u.id} value={u.id}>{u.username}</option>
                  ))}
                </select>
              </div>

              <div className="mt-6 space-y-4">
                {orders.length === 0 && !ordersLoading ? (
                  <p className="text-sm text-muted">No orders found.</p>
                ) : (
                  orders.map((order) => (
                    <article key={order.id} className="border border-line p-4 transition-colors hover:border-paper">
                      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                        <div>
                          <p className={`text-[10px] font-bold uppercase tracking-[0.2em] ${
                            order.status === 'CONFIRMED' ? 'text-green-500' : 
                            order.status === 'FAILED' ? 'text-accent' : 'text-muted'
                          }`}>
                            {order.status}
                          </p>
                          <p className="mt-1 text-xs text-muted font-mono">{order.id}</p>
                          <p className="mt-2 text-sm text-paper font-semibold">
                            UserID: <span className="text-muted font-normal">{order.user_id}</span>
                          </p>
                        </div>
                        <div className="text-right">
                          <p className="text-sm font-semibold uppercase tracking-hero">
                            ${order.total_amount.toFixed(2)}
                          </p>
                          <p className="mt-1 text-[10px] text-muted">
                            {new Date(order.created_at).toLocaleDateString()}
                          </p>
                        </div>
                      </div>
                    </article>
                  ))
                )}
              </div>

              {/* Pagination Controls */}
              {totalPages > 1 && (
                <div className="mt-8 flex items-center justify-center gap-4">
                  <button
                    onClick={() => setOrderFilters(curr => ({ ...curr, page: Math.max(1, curr.page - 1) }))}
                    disabled={orderFilters.page === 1}
                    className="border border-line px-4 py-2 text-[10px] font-bold uppercase tracking-widest disabled:opacity-30 disabled:cursor-not-allowed"
                  >
                    Prev
                  </button>
                  <span className="text-xs font-mono">
                    {orderFilters.page} / {totalPages}
                  </span>
                  <button
                    onClick={() => setOrderFilters(curr => ({ ...curr, page: Math.min(totalPages, curr.page + 1) }))}
                    disabled={orderFilters.page === totalPages}
                    className="border border-line px-4 py-2 text-[10px] font-bold uppercase tracking-widest disabled:opacity-30 disabled:cursor-not-allowed"
                  >
                    Next
                  </button>
                </div>
              )}
            </section>
          </div>
        </section>
      </div>
    </MainLayout>
  );
}
