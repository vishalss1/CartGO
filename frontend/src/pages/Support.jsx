import { useEffect, useState } from "react";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest, getTicketMessages, addTicketMessage } from "../utils/api";
import { TICKET_STATUSES } from "../utils/constants";
import { useAuth } from "../context/AuthContext";
import { useToast } from "../context/ToastContext";
import { logger } from "../utils/logger";

export default function SupportPage() {
  const { token } = useAuth();
  const { showSuccess, showError } = useToast();
  const [tickets, setTickets] = useState([]);
  const [filters, setFilters] = useState({ orderId: "", status: "" });
  const [loading, setLoading] = useState(true);
  const [expandedTicketId, setExpandedTicketId] = useState("");
  const [updatingId, setUpdatingId] = useState("");
  const [messages, setMessages] = useState([]);
  const [messagesLoading, setMessagesLoading] = useState(false);
  const [replyText, setReplyText] = useState("");
  const [replying, setReplying] = useState(false);

  async function loadTickets(silent = false) {
    if (!silent) {
      setTickets([]);
      setLoading(true);
    }

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
      const validTickets = Array.isArray(ticketList) ? ticketList : [];
      
      logger.info(`SupportAgent: API returned ${validTickets.length} tickets.`);
      setTickets(validTickets);
    } catch (loadError) {
      if (!silent) showError(loadError.message);
      logger.error("SupportAgent Load Error:", loadError);
    } finally {
      if (!silent) setLoading(false);
    }
  }

  useEffect(() => {
    loadTickets();
    const interval = setInterval(() => loadTickets(true), 15000);
    return () => clearInterval(interval);
  }, []);

  async function updateStatus(ticketId, status) {
    if (updatingId === ticketId) {
      return;
    }
    setUpdatingId(ticketId);

    try {
      await authenticatedRequest(`/support/tickets/${ticketId}/status`, token, {
        method: "PATCH",
        body: JSON.stringify({ status }),
      });
      showSuccess(`Ticket updated to ${status.replaceAll("_", " ").toLowerCase()}.`);
      
      // Update local state instead of full re-fetch
      setTickets((current) =>
        current.map((t) => (t.id === ticketId ? { ...t, status } : t)),
      );
    } catch (updateError) {
      showError(updateError.message);
    } finally {
      setUpdatingId("");
    }
  }

  async function toggleMessages(ticketId) {
    if (expandedTicketId === ticketId) {
      setExpandedTicketId("");
      setMessages([]);
      return;
    }

    setExpandedTicketId(ticketId);
    setMessagesLoading(true);

    try {
      const messageList = await getTicketMessages(ticketId, token);
      const validMessages = Array.isArray(messageList) ? messageList : [];
      logger.info(`Support: Loaded ${validMessages.length} messages for ticket ${ticketId}`);
      setMessages(validMessages);
    } catch (error) {
      logger.error("Support Toggle Messages Error:", error);
      setMessages([]);
    } finally {
      setMessagesLoading(false);
    }
  }

  async function handleReply(ticketId) {
    if (!replyText.trim() || replying) {
      return;
    }

    setReplying(true);

    try {
      logger.info(`Support: Sending reply to ticket ${ticketId}`);
      await addTicketMessage(ticketId, replyText.trim(), token);
      setReplyText("");
      showSuccess("Reply sent.");
      
      logger.info(`Support: Refreshing messages for ticket ${ticketId}`);
      // Refresh only messages
      const messageList = await getTicketMessages(ticketId, token);
      const validMessages = Array.isArray(messageList) ? messageList : [];
      logger.info(`Support: Received ${validMessages.length} messages after reply`);
      setMessages(validMessages);
    } catch (replyError) {
      logger.error("Support Reply Error:", replyError);
      showError(replyError.message);
    } finally {
      setReplying(false);
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
            Review and respond to customer tickets. Filter by order or status, update ticket
            progress, and communicate directly with users.
          </p>
        </section>

        <div className="grid gap-4 md:grid-cols-[1fr_220px_auto_auto]">
          <input
            value={filters.orderId}
            onChange={(event) => setFilters((current) => ({ ...current, orderId: event.target.value }))}
            placeholder="Filter by order ID"
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
          <button
            type="button"
            onClick={loadTickets}
            disabled={loading}
            className="inline-flex bg-surface border border-line px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-paper transition-colors hover:border-paper disabled:cursor-not-allowed"
          >
            {loading ? "Syncing" : "Sync tickets"}
          </button>
        </div>

        <div className="grid gap-4">
          {loading ? (
            <StatusBanner>Loading tickets</StatusBanner>
          ) : tickets.length === 0 ? (
            <StatusBanner>No tickets found</StatusBanner>
          ) : (
            tickets.map((ticket) => (
              <article key={ticket.id} className="border border-line">
                <div className="grid gap-4 p-5 lg:grid-cols-[1fr_260px] lg:items-start">
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
                    <button
                      type="button"
                      onClick={() => toggleMessages(ticket.id)}
                      className="mt-4 text-xs font-semibold uppercase tracking-[0.16em] text-paper underline transition-opacity hover:opacity-70"
                    >
                      {expandedTicketId === ticket.id ? "Hide messages" : "View messages"}
                    </button>
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

                {expandedTicketId === ticket.id ? (
                  <div className="border-t border-line p-5">
                    <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-muted">
                      Messages
                    </p>
                    <div className="mt-4 space-y-3">
                      {messagesLoading ? (
                        <p className="text-sm text-muted">Loading messages...</p>
                      ) : messages.length === 0 ? (
                        <p className="text-sm text-muted">No messages yet.</p>
                      ) : (
                        messages.map((msg, index) => (
                          <div key={msg.id ?? index} className="border border-line p-3">
                            <p className="text-xs font-semibold uppercase tracking-[0.14em] text-muted">
                              {msg.sender_id ? `User ${msg.sender_id.slice(0, 8)}` : "System"}
                            </p>
                            <p className="mt-2 text-sm leading-relaxed text-paper">
                              {msg.content}
                            </p>
                          </div>
                        ))
                      )}
                    </div>

                    <div className="mt-4 flex gap-3">
                      <textarea
                        value={replyText}
                        onChange={(event) => setReplyText(event.target.value)}
                        placeholder="Type your reply..."
                        className="h-20 flex-1 border border-line bg-transparent px-4 py-3 text-sm outline-none transition-colors focus:border-paper"
                      />
                      <button
                        type="button"
                        onClick={() => handleReply(ticket.id)}
                        disabled={replying || !replyText.trim()}
                        className="self-end bg-paper px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-surface transition-colors hover:bg-white disabled:cursor-not-allowed disabled:bg-line"
                      >
                        {replying ? "Sending" : "Reply"}
                      </button>
                    </div>
                  </div>
                ) : null}
              </article>
            ))
          )}
        </div>
      </div>
    </MainLayout>
  );
}
