import { useEffect, useState } from "react";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest } from "../utils/api";
import { TICKET_STATUSES } from "../utils/constants";
import { useAuth } from "../context/AuthContext";

export default function SupportPage() {
  const { token } = useAuth();
  const [tickets, setTickets] = useState([]);
  const [filters, setFilters] = useState({ orderId: "", status: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(true);

  async function loadTickets() {
    setLoading(true);
    setError("");

    try {
      const params = new URLSearchParams();
      if (filters.orderId) {
        params.set("order_id", filters.orderId);
      }
      if (filters.status) {
        params.set("status", filters.status);
      }
      const query = params.toString() ? `?${params.toString()}` : "";
      const ticketList = await authenticatedRequest(`/support/tickets${query}`, token);
      setTickets(Array.isArray(ticketList) ? ticketList : []);
    } catch (loadError) {
      setError(loadError.message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadTickets();
  }, []);

  async function updateStatus(ticketId, status) {
    setError("");
    setSuccess("");

    try {
      await authenticatedRequest(`/support/tickets/${ticketId}/status`, token, {
        method: "PATCH",
        body: JSON.stringify({ status }),
      });
      setSuccess(`Ticket ${ticketId} updated to ${status}.`);
      await loadTickets();
    } catch (updateError) {
      setError(updateError.message);
    }
  }

  return (
    <MainLayout>
      <div className="space-y-6">
        <section className="border border-line p-6 sm:p-8">
          <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
            Support workspace
          </p>
          <h1 className="mt-4 text-[3.2rem] font-black uppercase leading-[0.9] tracking-hero sm:text-[4.8rem]">
            Support
          </h1>
          <p className="mt-4 max-w-[34rem] text-sm leading-relaxed text-muted">
            Support tickets are filtered by order ID or ticket status and updated directly through
            support-service.
          </p>
        </section>

        <div className="grid gap-4 md:grid-cols-[1fr_220px_auto]">
          <input
            value={filters.orderId}
            onChange={(event) => setFilters((current) => ({ ...current, orderId: event.target.value }))}
            placeholder="Filter by order_id"
            className="border border-line bg-transparent px-4 py-4 outline-none focus:border-paper"
          />
          <select
            value={filters.status}
            onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}
            className="border border-line bg-surface px-4 py-4 outline-none focus:border-paper"
          >
            <option value="">All statuses</option>
            {TICKET_STATUSES.map((status) => (
              <option key={status} value={status}>
                {status}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={loadTickets}
            className="inline-flex bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white"
          >
            Apply filters
          </button>
        </div>

        {error ? <StatusBanner tone="error">{error}</StatusBanner> : null}
        {success ? <StatusBanner tone="success">{success}</StatusBanner> : null}

        <div className="grid gap-4">
          {loading ? (
            <StatusBanner>Loading support tickets</StatusBanner>
          ) : tickets.length === 0 ? (
            <StatusBanner>No tickets returned for the current filters.</StatusBanner>
          ) : (
            tickets.map((ticket) => (
              <article key={ticket.id} className="border border-line p-5">
                <div className="grid gap-4 lg:grid-cols-[1fr_260px] lg:items-start">
                  <div>
                    <h2 className="text-[1.8rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                      {ticket.subject}
                    </h2>
                    <p className="mt-2 text-sm text-muted">Ticket: {ticket.id}</p>
                    <p className="mt-2 text-sm text-muted">Order: {ticket.order_id ?? "None"}</p>
                    <p className="mt-2 text-sm text-muted">Priority: {ticket.priority}</p>
                    <p className="mt-2 text-sm text-muted">
                      Assigned agent: {ticket.assigned_agent_id ?? "Unassigned"}
                    </p>
                  </div>
                  <select
                    value={ticket.status}
                    onChange={(event) => updateStatus(ticket.id, event.target.value)}
                    className="border border-line bg-surface px-4 py-4 text-sm uppercase tracking-[0.16em] outline-none focus:border-paper"
                  >
                    {TICKET_STATUSES.map((status) => (
                      <option key={status} value={status}>
                        {status}
                      </option>
                    ))}
                  </select>
                </div>
              </article>
            ))
          )}
        </div>
      </div>
    </MainLayout>
  );
}
