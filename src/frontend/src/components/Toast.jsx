import { useEffect, useState } from "react";

const TONE_STYLES = {
  success: "border-paper text-paper",
  error: "border-accent text-accent",
  info: "border-line text-muted",
};

function ToastItem({ toast, onDismiss }) {
  const [exiting, setExiting] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      setExiting(true);
      setTimeout(() => onDismiss(toast.id), 300);
    }, 4000);
    return () => clearTimeout(timer);
  }, [onDismiss, toast.id]);

  return (
    <div
      role="alert"
      className={`border bg-surface px-5 py-4 text-xs font-semibold uppercase tracking-[0.18em] transition-all duration-300 ${TONE_STYLES[toast.tone] ?? TONE_STYLES.info} ${exiting ? "translate-x-full opacity-0" : "translate-x-0 opacity-100"}`}
      style={{ pointerEvents: "auto" }}
    >
      <div className="flex items-center justify-between gap-4">
        <span>{toast.message}</span>
        <button
          type="button"
          onClick={() => {
            setExiting(true);
            setTimeout(() => onDismiss(toast.id), 300);
          }}
          className="shrink-0 text-[10px] opacity-60 transition-opacity hover:opacity-100"
          aria-label="Dismiss"
        >
          ✕
        </button>
      </div>
    </div>
  );
}

export default function ToastContainer({ toasts, onDismiss }) {
  if (toasts.length === 0) {
    return null;
  }

  return (
    <div
      aria-live="polite"
      className="pointer-events-none fixed bottom-6 right-6 z-50 flex w-full max-w-sm flex-col gap-3"
    >
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} onDismiss={onDismiss} />
      ))}
    </div>
  );
}
