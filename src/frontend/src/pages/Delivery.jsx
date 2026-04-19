import { useEffect, useState } from "react";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest } from "../utils/api";
import { DELIVERY_STATUSES } from "../utils/constants";
import { useAuth } from "../context/AuthContext";
import { useToast } from "../context/ToastContext";
import { logger } from "../utils/logger";

export default function DeliveryPage() {
  const { token, user } = useAuth();
  const { showSuccess, showError } = useToast();
  const [deliveries, setDeliveries] = useState([]);
  const [availableDeliveries, setAvailableDeliveries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [updatingId, setUpdatingId] = useState("");

  async function loadDeliveries(silent = false) {
    if (!silent) {
      setDeliveries([]);
      setLoading(true);
    }
    try {
      const [myList, availableList] = await Promise.all([
        authenticatedRequest(`/deliveries/partner/${user.id}`, token),
        authenticatedRequest(`/deliveries/available`, token),
      ]);
      
      const validMine = Array.isArray(myList) ? myList : [];
      const validAvailable = Array.isArray(availableList) ? availableList : [];
      
      logger.info(`Delivery: Loaded ${validMine.length} assigned and ${validAvailable.length} available tasks.`);
      setDeliveries(validMine);
      setAvailableDeliveries(validAvailable);
    } catch (loadError) {
      if (!silent) showError(loadError.message);
      logger.error("Delivery Load Error:", loadError);
    } finally {
      if (!silent) setLoading(false);
    }
  }

  useEffect(() => {
    loadDeliveries();
    const interval = setInterval(() => loadDeliveries(true), 15000);
    return () => clearInterval(interval);
  }, [token, user.id]);

  async function updateStatus(deliveryId, status) {
    if (updatingId === deliveryId) {
      return;
    }
    setUpdatingId(deliveryId);

    try {
      await authenticatedRequest(`/deliveries/${deliveryId}/status`, token, {
        method: "PATCH",
        body: JSON.stringify({
          status,
          delivery_person_id: user.id,
        }),
      });
      showSuccess(`Delivery updated to ${status.replaceAll("_", " ").toLowerCase()}.`);
      
      // Update local state immediately for better UX
      setAvailableDeliveries((current) => {
        const found = current.find((d) => d.id === deliveryId);
        if (found) {
          // If it was available, move it to my deliveries
          setDeliveries((prev) => [...prev, { ...found, status, delivery_person_id: user.id }]);
          return current.filter((d) => d.id !== deliveryId);
        }
        return current;
      });

      setDeliveries((current) =>
        current.map((d) => (d.id === deliveryId ? { ...d, status } : d)),
      );
    } catch (updateError) {
      showError(updateError.message);
    } finally {
      setUpdatingId("");
    }
  }

  return (
    <MainLayout>
      <div className="space-y-6">
        <section className="border border-line p-6 sm:p-8">
          <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
                Delivery partner workspace
              </p>
              <h1 className="mt-4 text-[3.2rem] font-black uppercase leading-[0.9] tracking-hero sm:text-[4.8rem]">
                Delivery
              </h1>
              <p className="mt-4 max-w-[34rem] text-sm leading-relaxed text-muted">
                View your assigned deliveries and update their status as you pick up and complete
                each order.
              </p>
            </div>
            <button
              type="button"
              onClick={loadDeliveries}
              disabled={loading}
              className="inline-flex w-fit border border-line px-4 py-3 text-[10px] font-bold uppercase tracking-[0.2em] transition-colors hover:border-paper disabled:cursor-not-allowed"
            >
              {loading ? "Syncing..." : "Sync tasks"}
            </button>
          </div>
        </section>

        <div className="space-y-12">
          {availableDeliveries.length > 0 && (
            <div className="space-y-4">
              <h2 className="text-[11px] font-semibold uppercase tracking-[0.28em] text-accent">
                Available tasks (Unassigned)
              </h2>
              <div className="grid gap-4">
                {availableDeliveries.map((delivery) => (
                  <article key={delivery.id} className="border border-line bg-paper/5 p-5">
                    <div className="grid gap-4 lg:grid-cols-[1fr_260px] lg:items-start">
                      <div>
                        <h2 className="text-[1.8rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                          {delivery.status}
                        </h2>
                        <p className="mt-2 text-sm text-muted">{delivery.id}</p>
                        <p className="mt-3 text-sm leading-relaxed text-muted">
                          Order: {delivery.order_id}
                        </p>
                        <p className="mt-2 text-sm leading-relaxed text-muted">
                          Address: {delivery.delivery_address}
                        </p>
                      </div>
                      <button
                        onClick={() => updateStatus(delivery.id, "PICKED_UP")}
                        disabled={updatingId === delivery.id}
                        className="bg-paper px-4 py-4 text-xs font-bold uppercase tracking-[0.16em] text-surface transition-colors hover:bg-white disabled:cursor-not-allowed"
                      >
                        {updatingId === delivery.id ? "Assigning..." : "Claim & Pick up"}
                      </button>
                    </div>
                  </article>
                ))}
              </div>
            </div>
          )}

          <div className="space-y-4">
            <h2 className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">
              Your active tasks
            </h2>
            <div className="grid gap-4">
              {loading ? (
                <StatusBanner>Loading deliveries</StatusBanner>
              ) : deliveries.length === 0 && availableDeliveries.length === 0 ? (
                <StatusBanner>No deliveries assigned</StatusBanner>
              ) : deliveries.length === 0 ? (
                <StatusBanner>You have no active tasks. Claim one above.</StatusBanner>
              ) : (
                deliveries.map((delivery) => (
                  <article key={delivery.id} className="border border-line p-5">
                    <div className="grid gap-4 lg:grid-cols-[1fr_260px] lg:items-start">
                      <div>
                        <h2 className="text-[1.8rem] font-extrabold uppercase leading-[0.92] tracking-hero">
                          {delivery.status}
                        </h2>
                        <p className="mt-2 text-sm text-muted">{delivery.id}</p>
                        <p className="mt-3 text-sm leading-relaxed text-muted">
                          Order: {delivery.order_id}
                        </p>
                        <p className="mt-2 text-sm leading-relaxed text-muted">
                          Address: {delivery.delivery_address}
                        </p>
                      </div>
                      <select
                        value={delivery.status}
                        onChange={(event) => updateStatus(delivery.id, event.target.value)}
                        disabled={updatingId === delivery.id}
                        className="border border-line bg-surface px-4 py-4 text-sm uppercase tracking-[0.16em] outline-none focus:border-paper"
                      >
                        {DELIVERY_STATUSES.map((status) => (
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
        </div>
      </div>
    </MainLayout>
  );
}
