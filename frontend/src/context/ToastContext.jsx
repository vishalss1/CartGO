import { createContext, useCallback, useContext, useMemo, useState } from "react";
import ToastContainer from "../components/Toast";

const ToastContext = createContext(null);

let nextId = 0;

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);

  const dismiss = useCallback((id) => {
    setToasts((current) => current.filter((t) => t.id !== id));
  }, []);

  const push = useCallback((message, tone) => {
    const id = ++nextId;
    setToasts((current) => [...current, { id, message, tone }]);
  }, []);

  const value = useMemo(
    () => ({
      showSuccess: (msg) => push(msg, "success"),
      showError: (msg) => push(msg, "error"),
      showInfo: (msg) => push(msg, "info"),
    }),
    [push],
  );

  return (
    <ToastContext.Provider value={value}>
      {children}
      <ToastContainer toasts={toasts} onDismiss={dismiss} />
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within ToastProvider");
  }
  return context;
}
