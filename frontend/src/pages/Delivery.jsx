import { useEffect, useState } from "react";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest } from "../utils/api";
import { DELIVERY_STATUSES } from "../utils/constants";
import { useAuth } from "../context/AuthContext";

export default function DeliveryPage() {
  const { token, user } = useAuth();
  const [deliveries, setDeliveries] = useState([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(true);
  const [updatingId, setUpdatingId] = useState("");

  async function loadDeliveries() {
    setLoading(true);
    setError("");
    try {
      const deliveryList = await authenticatedRequest(`/deliveries/partner/${user.id}`, token);
      setDeliveries(Array.isArray(deliveryList) ? deliveryList : []);
    } catch (loadError) {
      setError(loadError.message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadDeliveries();
  }, [token, user.id]);

  async function updateStatus(deliveryId, status) {
    setUpdatingId(deliveryId);
    setError("");
    setSuccess("");

    try {
      await authenticatedRequest(`/deliveries/${deliveryId}/status`, token, {
        method: "PATCH",
        body: JSON.stringify({
          status,
          delivery_person_id: user.id,
        }),
      });
      setSuccess(`Delivery ${deliveryId} updated to ${status}.`);
      await loadDeliveries();
    } catch (updateError) {
      setError(updateError.message);
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
            Assigned deliveries are loaded from delivery-service using your authenticated user ID.
            Status updates post back to the same service through the gateway.
          </p>
        </section>

        {error ? <StatusBanner tone="error">{error}</StatusBanner> : null}
        {success ? <StatusBanner tone="success">{success}</StatusBanner> : null}

        <div className="grid gap-4">
          {loading ? (
            <StatusBanner>Loading assigned deliveries</StatusBanner>
          ) : deliveries.length === 0 ? (
            <StatusBanner>No deliveries assigned to this partner ID.</StatusBanner>
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
