import { useParams, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { useAuth } from "../context/AuthContext";
import { useSystemOrder } from "../hooks/useSystemOrder";
import MainLayout from "../layout/MainLayout";
import OrderLifecycleStepper from "../components/OrderLifecycleStepper";
import SystemLogViewer from "../components/SystemLogViewer";
import StatusBanner from "../components/StatusBanner";
import { SYSTEM_STATES } from "../utils/systemModel";
import { apiRequest, authenticatedRequest } from "../utils/api";

export default function OrderDetailsPage() {
  const { orderId } = useParams();
  const { token } = useAuth();
  const navigate = useNavigate();
  const system = useSystemOrder(orderId, token);
  const [productNames, setProductNames] = useState({});
  const [retrying, setRetrying] = useState(false);

  const { order, payment, delivery, derivedState, logs, inconsistency } = system;

  async function handleRetryPayment() {
    setRetrying(true);
    try {
      await authenticatedRequest("/payments", token, {
        method: "POST",
        body: JSON.stringify({
          order_id: order.id,
          amount: order.total_amount,
        }),
      });
      window.location.reload();
    } catch (err) {
      console.error(err);
      alert(err.message || "Failed to retry payment");
    } finally {
      setRetrying(false);
    }
  }

  useEffect(() => {
    if (!order || !order.items) return;

    async function fetchNames() {
      const names = { ...productNames };
      await Promise.all(
        order.items.map(async (item) => {
          if (names[item.product_id]) return;
          try {
            const p = await apiRequest(`/products/${item.product_id}`);
            names[item.product_id] = p.name;
          } catch {
            names[item.product_id] = "Product no longer in catalog";
          }
        })
      );
      setProductNames(names);
    }

    fetchNames();
  }, [order]);

  if (system.loading) {
    return (
      <MainLayout>
        <div className="flex min-h-[60vh] items-center justify-center">
          <p className="text-xs font-black uppercase tracking-[0.4em] animate-pulse">Initializing System State...</p>
        </div>
      </MainLayout>
    );
  }


  if (!order) {
    return (
      <MainLayout>
        <div className="space-y-6">
           <StatusBanner tone="error">System Error: Order Record Not Found</StatusBanner>
           <button onClick={() => navigate("/shop")} className="text-xs underline uppercase tracking-widest text-muted">Return to Shop</button>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="space-y-12">
        {/* Header Section */}
        <section className="border border-line p-8 sm:p-12">
          <div className="flex flex-col justify-between gap-8 md:flex-row md:items-end">
            <div className="space-y-4">
              <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
                Order Dashboard
              </p>
              <h1 className="text-[3.2rem] font-black uppercase leading-[0.9] tracking-hero sm:text-[4.8rem]">
                Trace {order.id.slice(0, 8)}
              </h1>
              <div className="flex flex-wrap gap-4">
                 <span className="bg-paper px-3 py-1 text-[10px] font-bold uppercase tracking-widest text-surface">
                   ID: {order.id}
                 </span>
                 <span className="border border-line px-3 py-1 text-[10px] font-bold uppercase tracking-widest text-muted">
                   {new Date(order.created_at).toLocaleString()}
                 </span>
              </div>
            </div>
            
            <div className="flex flex-col items-end gap-2">
               <p className="text-[10px] font-bold uppercase tracking-widest text-muted">System Status</p>
               <span className={`text-2xl font-black uppercase tracking-hero ${derivedState === SYSTEM_STATES.INCONSISTENT ? 'text-red-500' : 'text-paper'}`}>
                 {derivedState.replaceAll("_", " ")}
               </span>
            </div>
          </div>

          <OrderLifecycleStepper currentState={derivedState} inconsistency={inconsistency} />
        </section>

        <div className="grid gap-12 lg:grid-cols-[1fr_400px]">
          {/* Detailed Composition */}
          <section className="space-y-8">
            <div className="border border-line p-8">
               <h2 className="text-[1.8rem] font-black uppercase tracking-hero">Composition</h2>
               <div className="mt-8 space-y-6">
                  {order.items.map((item) => (
                    <div key={item.id} className="flex items-center justify-between border-b border-line pb-6 last:border-0 last:pb-0">
                      <div>
                        <p className="text-xs font-black uppercase tracking-widest">
                          {productNames[item.product_id] || `Product ${item.product_id.slice(0, 8)}`}
                        </p>
                        <p className="mt-1 text-sm text-muted italic">Quantity: {item.quantity}</p>
                      </div>
                      <div className="text-right">
                        <p className="text-lg font-bold">${(item.price_per_unit * item.quantity).toFixed(2)}</p>
                        <p className="text-[10px] text-muted uppercase tracking-widest">${item.price_per_unit.toFixed(2)} / unit</p>
                      </div>
                    </div>
                  ))}
               </div>
               <div className="mt-12 flex items-center justify-between border-t-2 border-paper pt-8">
                  <span className="text-sm font-black uppercase tracking-[0.32em]">Total Value</span>
                  <span className="text-[2.4rem] font-black">${order.total_amount.toFixed(2)}</span>
               </div>
            </div>

            <div className="grid gap-8 md:grid-cols-2">
               <div className="border border-line p-6">
                  <h3 className="text-[10px] font-black uppercase tracking-widest text-muted">Payment Detail</h3>
                  <div className="mt-4 space-y-2">
                     <p className="text-sm font-bold uppercase tracking-normal">Status: {payment?.status || "NOT_INITIATED"}</p>
                     <p className="text-[10px] text-muted">Correlation: {payment?.id || "N/A"}</p>
                  </div>
                  {derivedState === SYSTEM_STATES.PAYMENT_FAILED && (
                     <button 
                       onClick={handleRetryPayment}
                       disabled={retrying}
                       className="mt-6 w-full bg-red-500 py-3 text-[10px] font-bold uppercase tracking-widest text-white transition-colors hover:bg-red-600 disabled:opacity-50"
                     >
                        {retrying ? "Retrying..." : "Initiate Payment Retry"}
                     </button>
                  )}
               </div>
               <div className="border border-line p-6">
                  <h3 className="text-[10px] font-black uppercase tracking-widest text-muted">Delivery Detail</h3>
                  <div className="mt-4 space-y-2">
                     <p className="text-sm font-bold uppercase tracking-normal">Status: {delivery?.status || "PREPARING"}</p>
                     <p className="text-[10px] text-muted leading-relaxed">Destination: {delivery?.delivery_address || order.delivery_address}</p>
                  </div>
               </div>
            </div>
          </section>

          {/* System Audit Log Sidebar */}
          <aside className="border border-line p-8">
             <SystemLogViewer logs={logs} />
             
             <div className="mt-12 border-t border-line pt-8">
                <h4 className="text-[10px] font-black uppercase tracking-widest text-muted">Support Context</h4>
                <p className="mt-4 text-[10px] leading-relaxed text-muted">
                   Need assistance with this system state? Our agents have full access to these audit logs for rapid troubleshooting.
                </p>
                <button 
                  onClick={() => navigate(`/support?order_id=${order.id}`)}
                  className="mt-6 border border-line px-6 py-3 text-[9px] font-bold uppercase tracking-widest text-paper transition-all hover:bg-paper hover:text-surface"
                >
                  Create Linked Ticket
                </button>
             </div>
          </aside>
        </div>
      </div>
    </MainLayout>
  );
}
