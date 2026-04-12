import { useEffect, useState } from "react";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest } from "../utils/api";
import { DELIVERY_STATUSES } from "../utils/constants";
import { useAuth } from "../context/AuthContext";
import { useToast } from "../context/ToastContext";

export default function DeliveryPage() {
  const { token, user } = useAuth();
  const { showSuccess, showError } = useToast();
  const [deliveries, setDeliveries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [updatingId, setUpdatingId] = useState("");

  async function loadDeliveries() {
    setLoading(true);
    try {
      const deliveryList = await authenticatedRequest(`/deliveries/partner/${user.id}`, token);
      setDeliveries(Array.isArray(deliveryList) ? deliveryList : []);
    } catch (loadError) {
      showError(loadError.message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadDeliveries();
  }, [token, user.id]);

  async function updateStatus(deliveryId, status) {
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
      await loadDeliveries();
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
        </section>

        <div className="grid gap-4">
          {loading ? (
            <StatusBanner>Loading deliveries</StatusBanner>
          ) : deliveries.length === 0 ? (
            <StatusBanner>No deliveries assigned</StatusBanner>
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
    </MainLayout>
  );
}
