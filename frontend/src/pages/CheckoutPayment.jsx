import { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import MainLayout from "../layout/MainLayout";
import StatusBanner from "../components/StatusBanner";
import { authenticatedRequest } from "../utils/api";
import { useAuth } from "../context/AuthContext";
import { useToast } from "../context/ToastContext";
import { storeCart } from "../utils/auth";

export default function CheckoutPaymentPage() {
  const { token, user } = useAuth();
  const { showSuccess, showError } = useToast();
  const navigate = useNavigate();
  const { state } = useLocation();
  const { cart, deliveryAddress } = state || { cart: [], deliveryAddress: "" };

  const [processing, setProcessing] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState("");
  const [card, setCard] = useState({ number: "", expiry: "", cvc: "" });

  useEffect(() => {
    if (!cart.length || !deliveryAddress) {
      navigate("/shop");
    }
  }, [cart, deliveryAddress, navigate]);

  const cartTotal = cart.reduce((sum, item) => sum + item.price * item.quantity, 0);

  async function handlePayment(event) {
    event.preventDefault();
    setProcessing(true);
    setError("");

    // Realistic simulation delay
    await new Promise((resolve) => setTimeout(resolve, 2500));

    try {
      await authenticatedRequest("/orders", token, {
        method: "POST",
        body: JSON.stringify({
          items: cart.map((item) => ({
            product_id: item.product_id,
            quantity: item.quantity,
          })),
          delivery_address: deliveryAddress,
        }),
      });

      setSuccess(true);
      showSuccess("Payment processed and order placed!");
      storeCart([]); // Clear cart storage
      
      // Auto-redirect to shop after success
      setTimeout(() => navigate("/shop"), 3000);
    } catch (checkoutError) {
      setError(checkoutError.message);
      showError(checkoutError.message);
    } finally {
      setProcessing(false);
    }
  }

  if (success) {
    return (
      <MainLayout>
        <div className="flex min-h-[60vh] flex-col items-center justify-center space-y-8 text-center">
          <div className="flex h-24 w-24 items-center justify-center rounded-full bg-paper text-surface">
            <svg
              className="h-12 w-12"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={3}
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>
          <div className="space-y-4">
            <h1 className="text-[4rem] font-black uppercase leading-[0.9] tracking-hero sm:text-[6rem]">
              Confirmed
            </h1>
            <p className="mx-auto max-w-md text-base leading-relaxed text-muted">
              Thank you for your order. Your payment was successful and our warehouse team is
              already preparing your items.
            </p>
          </div>
          <button
            onClick={() => navigate("/shop")}
            className="bg-paper px-8 py-4 text-xs font-bold uppercase tracking-[0.2em] text-surface transition-colors hover:bg-white"
          >
            Back to Shop
          </button>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="grid gap-8 lg:grid-cols-[1fr_450px]">
        <section className="space-y-8">
          <div className="border border-line p-6 sm:p-8">
            <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
              Secure Checkout
            </p>
            <h1 className="mt-4 text-[3.2rem] font-black uppercase leading-[0.9] tracking-hero sm:text-[4.8rem]">
              Payment
            </h1>
            <p className="mt-4 max-w-[34rem] text-sm leading-relaxed text-muted">
              Please enter your payment details below. Your transaction is processed through our
              encrypted gateway.
            </p>
          </div>

          <form onSubmit={handlePayment} className="grid gap-6 border border-line p-6 sm:p-8">
            <div className="space-y-2">
              <label className="text-[11px] font-semibold uppercase tracking-[0.24em] text-muted">
                Card Number
              </label>
              <input
                required
                value={card.number}
                onChange={(e) => setCard({ ...card, number: e.target.value })}
                placeholder="0000 0000 0000 0000"
                className="w-full border border-line bg-transparent px-5 py-4 text-lg outline-none transition-colors focus:border-paper"
              />
            </div>
            <div className="grid grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-[11px] font-semibold uppercase tracking-[0.24em] text-muted">
                  Expiry
                </label>
                <input
                  required
                  value={card.expiry}
                  onChange={(e) => setCard({ ...card, expiry: e.target.value })}
                  placeholder="MM/YY"
                  className="w-full border border-line bg-transparent px-5 py-4 text-lg outline-none transition-colors focus:border-paper"
                />
              </div>
              <div className="space-y-2">
                <label className="text-[11px] font-semibold uppercase tracking-[0.24em] text-muted">
                  CVC
                </label>
                <input
                  required
                  value={card.cvc}
                  onChange={(e) => setCard({ ...card, cvc: e.target.value })}
                  placeholder="000"
                  className="w-full border border-line bg-transparent px-5 py-4 text-lg outline-none transition-colors focus:border-paper"
                />
              </div>
            </div>

            {error ? (
              <div className="mt-2">
                <StatusBanner tone="error">{error}</StatusBanner>
              </div>
            ) : null}

            <button
              type="submit"
              disabled={processing}
              className="mt-4 inline-flex items-center justify-center bg-paper px-8 py-5 text-sm font-black uppercase tracking-[0.2em] text-surface transition-colors hover:bg-white disabled:cursor-not-allowed disabled:bg-line"
            >
              {processing ? (
                <>
                  <svg
                    className="-ml-1 mr-3 h-5 w-5 animate-spin text-surface"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    ></circle>
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  Processing
                </>
              ) : (
                `Pay $${cartTotal.toFixed(2)}`
              )}
            </button>
          </form>
        </section>

        <aside className="space-y-6">
          <div className="border border-line p-6">
            <h2 className="text-[2rem] font-extrabold uppercase leading-[0.92] tracking-hero">
              Summary
            </h2>
            <div className="mt-6 space-y-4">
              {cart.map((item) => (
                <div key={item.product_id} className="flex justify-between border-b border-line pb-4 last:border-0">
                  <div>
                    <p className="font-semibold uppercase tracking-[0.08em]">{item.name}</p>
                    <p className="text-xs text-muted">Qty: {item.quantity}</p>
                  </div>
                  <span className="font-semibold">${(item.price * item.quantity).toFixed(2)}</span>
                </div>
              ))}
            </div>
            <div className="mt-8 border-t border-line pt-6">
              <div className="flex items-center justify-between">
                <span className="text-sm uppercase tracking-[0.2em] text-muted">Total</span>
                <span className="text-3xl font-black">${cartTotal.toFixed(2)}</span>
              </div>
            </div>
          </div>

          <div className="border border-line p-6">
            <h2 className="text-sm font-bold uppercase tracking-[0.2em] text-muted">
              Delivery to
            </h2>
            <p className="mt-3 text-sm leading-relaxed text-paper italic">
              &quot;{deliveryAddress}&quot;
            </p>
          </div>
        </aside>
      </div>
    </MainLayout>
  );
}
