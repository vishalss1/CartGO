import { useEffect, useState } from "react";
import { useLocation } from "react-router-dom";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest } from "../utils/api";
import { useAuth } from "../context/AuthContext";
import { useToast } from "../context/ToastContext";
import { logger } from "../utils/logger";

export default function SupportPortalPage() {
  const { token, user } = useAuth();
  const { showSuccess, showError } = useToast();
  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  const prefilledOrderId = searchParams.get("order_id") || "";

  const [tickets, setTickets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [subject, setSubject] = useState(prefilledOrderId ? `Issue with Order #${prefilledOrderId.slice(0, 8)}` : "");
  const [message, setMessage] = useState(prefilledOrderId ? `I am having an issue with my order (ID: ${prefilledOrderId}). ` : "");
  const [creating, setCreating] = useState(false);
  const [activeTicketId, setActiveTicketId] = useState("");
  const [reply, setReply] = useState("");
  const [sendingReply, setSendingReply] = useState(false);

  const loadTickets = async (silent = false) => {
    if (!silent) {
      setTickets([]);
      setLoading(true);
      setRefreshing(true);
    }
    try {
      // For CUSTOMER, ListTickets should return only their tickets (backend implementation dependent)
      // Assuming GET /support/tickets filters by token's userId
      const data = await authenticatedRequest("/support/tickets", token);
      const ticketList = Array.isArray(data) ? data : [];
      
      logger.info(`SupportPortal: API returned ${ticketList.length} tickets.`);
      setTickets(ticketList);
    } catch (err) {
      if (!silent) showError(err.message);
      logger.error("SupportPortal Load Error:", err);
    } finally {
      if (!silent) {
        setRefreshing(false);
        setLoading(false);
      }
    }
  };

  useEffect(() => {
    loadTickets();
    const interval = setInterval(() => loadTickets(true), 15000);
    return () => clearInterval(interval);
  }, [token]);

  async function handleCreateTicket(e) {
    e.preventDefault();
    setCreating(true);
    try {
      await authenticatedRequest("/support/tickets", token, {
        method: "POST",
        body: JSON.stringify({ 
          subject, 
          message,
          order_id: prefilledOrderId || undefined
        }),
      });
      showSuccess("Support ticket created successfully.");
      setSubject("");
      setMessage("");
      loadTickets();
    } catch (err) {
      showError(err.message);
    } finally {
      setCreating(false);
    }
  }

  async function handleAddMessage(ticketId) {
    if (!reply.trim()) return;
    setSendingReply(true);
    try {
      await authenticatedRequest(`/support/tickets/${ticketId}/messages`, token, {
        method: "POST",
        body: JSON.stringify({ message: reply }),
      });
      setReply("");
      showSuccess("Message sent.");
      loadTickets(true);
    } catch (err) {
      showError(err.message);
    } finally {
      setSendingReply(false);
    }
  }

  const activeTicket = tickets.find(t => t.id === activeTicketId);

  return (
    <MainLayout>
      <div className="grid gap-8 lg:grid-cols-[400px_1fr]">
        <aside className="space-y-6">
          <div className="border border-line p-6 sm:p-8">
            <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
              Customer Support
            </p>
            <h1 className="mt-4 text-[2.8rem] font-black uppercase leading-[0.9] tracking-hero">
              Portal
            </h1>
            <p className="mt-4 text-sm leading-relaxed text-muted">
              Need help? Create a ticket below or track your existing support requests.
            </p>
          </div>

          <form onSubmit={handleCreateTicket} className="grid gap-4 border border-line p-6 shadow-sm">
            <h2 className="text-sm font-bold uppercase tracking-[0.2em]">New Ticket</h2>
            <div className="space-y-1">
              <label className="text-[10px] font-bold uppercase tracking-widest text-muted">Subject</label>
              <input
                required
                value={subject}
                onChange={(e) => setSubject(e.target.value)}
                placeholder="e.g. Delayed Delivery"
                className="w-full border border-line bg-transparent px-4 py-3 text-sm outline-none focus:border-paper"
              />
            </div>
            <div className="space-y-1">
              <label className="text-[10px] font-bold uppercase tracking-widest text-muted">Description</label>
              <textarea
                required
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                placeholder="Describe your issue in detail..."
                className="h-32 w-full border border-line bg-transparent px-4 py-3 text-sm outline-none focus:border-paper"
              />
            </div>
            <button
              disabled={creating}
              type="submit"
              className="bg-paper py-3 text-[10px] font-bold uppercase tracking-[0.2em] text-surface transition-colors hover:bg-white disabled:bg-line"
            >
              {creating ? "Creating..." : "Create Ticket"}
            </button>
          </form>

          <div className="border border-line p-6">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-bold uppercase tracking-[0.2em]">History</h2>
              <button onClick={() => loadTickets()} disabled={refreshing} className="text-[10px] uppercase underline text-muted hover:text-paper">
                {refreshing ? "Syncing" : "Refresh"}
              </button>
            </div>
            <div className="mt-6 space-y-3">
              {loading ? (
                <p className="text-xs text-muted">Loading...</p>
              ) : tickets.length === 0 ? (
                <p className="text-xs text-muted">No tickets found.</p>
              ) : (
                tickets.map(t => (
                  <button
                    key={t.id}
                    onClick={() => setActiveTicketId(t.id)}
                    className={`w-full border p-4 text-left transition-all ${activeTicketId === t.id ? 'border-paper bg-line' : 'border-line hover:border-paper'}`}
                  >
                    <div className="flex justify-between gap-2">
                       <p className="truncate text-xs font-bold uppercase tracking-tight">{t.subject}</p>
                       <span className="text-[9px] font-black uppercase text-paper">{t.status}</span>
                    </div>
                    <p className="mt-1 text-[9px] text-muted">{t.id.slice(0, 8)}</p>
                  </button>
                ))
              )}
            </div>
          </div>
        </aside>

        <section className="border border-line p-6 sm:p-8">
          {activeTicket ? (
            <div className="flex h-full flex-col">
              <div className="border-b border-line pb-6">
                <div className="flex items-center gap-4">
                  <h2 className="text-[2rem] font-black uppercase tracking-hero">{activeTicket.subject}</h2>
                  <span className="border border-line px-2 py-1 text-[10px] font-bold uppercase tracking-widest">{activeTicket.status}</span>
                </div>
                {activeTicket.order_id && (
                   <p className="mt-2 text-xs text-muted uppercase tracking-widest">Linked to Order: {activeTicket.order_id}</p>
                )}
              </div>
              
              <div className="flex-1 overflow-y-auto py-6 space-y-6">
                 {/* This section would ideally fetch and show messages. 
                     For now we show the origin message. */}
                 <div className="space-y-1">
                    <p className="text-[10px] font-bold uppercase text-muted">Your Request:</p>
                    <p className="text-sm border-l-2 border-line pl-4 py-1 italic text-muted">
                        {activeTicket.message || "Initial message not available in list view."}
                    </p>
                 </div>
                 
                 <div className="space-y-4">
                    <p className="text-[10px] font-bold uppercase text-muted">Conversation</p>
                    <div className="border border-line bg-surface/50 p-4 text-xs italic text-muted">
                       Reply from support will appear here.
                    </div>
                 </div>
              </div>

              <div className="mt-auto border-t border-line pt-6">
                <div className="flex flex-col gap-3">
                   <textarea
                     value={reply}
                     onChange={(e) => setReply(e.target.value)}
                     placeholder="Type your response..."
                     className="h-24 w-full border border-line bg-transparent px-4 py-3 text-sm outline-none focus:border-paper"
                   />
                   <button
                     onClick={() => handleAddMessage(activeTicket.id)}
                     disabled={sendingReply || !reply.trim()}
                     className="self-end bg-paper px-8 py-3 text-[10px] font-bold uppercase tracking-[0.2em] text-surface transition-colors hover:bg-white disabled:bg-line"
                   >
                     {sendingReply ? "Sending..." : "Send Reply"}
                   </button>
                </div>
              </div>
            </div>
          ) : (
            <div className="flex h-full flex-col items-center justify-center text-center">
               <div className="h-16 w-16 border border-line flex items-center justify-center opacity-20">
                  <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                     <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                  </svg>
               </div>
               <p className="mt-4 text-sm font-bold uppercase tracking-[0.2em] text-muted">Select a ticket to view conversation</p>
            </div>
          )}
        </section>
      </div>
    </MainLayout>
  );
}
