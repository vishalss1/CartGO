import { useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import StatusBanner from "../components/StatusBanner";
import { useAuth } from "../context/AuthContext";
import { useToast } from "../context/ToastContext";
import { ROLE_TO_ROUTE, VALID_ROLES } from "../utils/constants";

export default function RegisterPage() {
  const { register, isAuthenticated, role, backendStatus } = useAuth();
  const { showSuccess, showError } = useToast();
  const navigate = useNavigate();
  const [form, setForm] = useState({ username: "", email: "", password: "", role: "CUSTOMER" });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  if (isAuthenticated && role) {
    return <Navigate to={ROLE_TO_ROUTE[role]} replace />;
  }

  async function handleSubmit(event) {
    event.preventDefault();
    setSubmitting(true);
    setError("");

    try {
      await register(form);
      showSuccess("Account created successfully. Please sign in.");
      navigate("/login", { replace: true });
    } catch (submitError) {
      setError(submitError.message);
      showError(submitError.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="min-h-screen bg-surface px-5 py-8 text-paper sm:px-8 lg:px-12 xl:px-16">
      <div className="grid min-h-[calc(100vh-4rem)] border border-line lg:grid-cols-[1.1fr_0.9fr]">
        <section className="border-b border-line p-6 sm:p-8 lg:border-b-0 lg:border-r lg:p-12">
          <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
            CartGO
          </p>
          <h1 className="mt-6 text-[4rem] font-black uppercase leading-[0.88] tracking-hero sm:text-[5.4rem] xl:text-[7rem]">
            Create account
          </h1>
          <p className="mt-6 max-w-[24rem] text-base leading-relaxed text-muted">
            Join CartGO to start shopping, managing inventory, or fulfilling deliveries based on
            your assigned role.
          </p>
        </section>

        <section className="p-6 sm:p-8 lg:p-12">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">
                Username
              </label>
              <input
                type="text"
                value={form.username}
                onChange={(event) => setForm((current) => ({ ...current, username: event.target.value }))}
                className="w-full border border-line bg-transparent px-4 py-4 text-base text-paper outline-none transition-colors focus:border-paper"
                placeholder="Choose a username"
                required
                minLength={3}
              />
            </div>

            <div className="space-y-2">
              <label className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">
                Email
              </label>
              <input
                type="email"
                value={form.email}
                onChange={(event) => setForm((current) => ({ ...current, email: event.target.value }))}
                className="w-full border border-line bg-transparent px-4 py-4 text-base text-paper outline-none transition-colors focus:border-paper"
                placeholder="you@example.com"
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">
                Password
              </label>
              <input
                type="password"
                value={form.password}
                onChange={(event) =>
                  setForm((current) => ({ ...current, password: event.target.value }))
                }
                className="w-full border border-line bg-transparent px-4 py-4 text-base text-paper outline-none transition-colors focus:border-paper"
                placeholder="Minimum 6 characters"
                required
                minLength={6}
              />
            </div>

            <div className="space-y-2">
              <label className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">
                Role
              </label>
              <select
                value={form.role}
                onChange={(event) => setForm((current) => ({ ...current, role: event.target.value }))}
                className="w-full border border-line bg-surface px-4 py-4 text-base text-paper outline-none transition-colors focus:border-paper"
              >
                {VALID_ROLES.map((r) => (
                  <option key={r} value={r}>
                    {r.replaceAll("_", " ")}
                  </option>
                ))}
              </select>
            </div>

            {error ? <StatusBanner tone="error">{error}</StatusBanner> : null}

            <button
              type="submit"
              disabled={submitting || !backendStatus.online}
              className="inline-flex items-center justify-center bg-paper px-5 py-4 text-sm font-semibold uppercase tracking-[0.18em] text-surface transition-colors duration-200 hover:bg-white disabled:cursor-not-allowed disabled:border-line disabled:bg-line"
            >
              {submitting ? "Creating account" : "Register"}
            </button>
          </form>

          <div className="mt-10 border-t border-line pt-6">
            <p className="text-sm text-muted">
              Already have an account?{" "}
              <Link to="/login" className="font-semibold text-paper underline transition-opacity hover:opacity-70">
                Sign in
              </Link>
            </p>
          </div>
        </section>
      </div>
    </div>
  );
}
