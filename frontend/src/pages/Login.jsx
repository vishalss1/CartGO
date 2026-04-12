import { useState } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import StatusBanner from "../components/StatusBanner";
import { useAuth } from "../context/AuthContext";
import { ROLE_TO_ROUTE, VALID_ROLES } from "../utils/constants";

export default function LoginPage() {
  const { login, isAuthenticated, role, backendStatus } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [form, setForm] = useState({ email: "", password: "" });
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
      const user = await login(form);
      if (!VALID_ROLES.includes(user.role)) {
        throw new Error(`Unsupported role returned by backend: ${user.role}`);
      }
      const destination = location.state?.from?.pathname;
      navigate(destination && destination !== "/login" ? destination : ROLE_TO_ROUTE[user.role], {
        replace: true,
      });
    } catch (submitError) {
      setError(submitError.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="min-h-screen bg-surface px-5 py-8 text-paper sm:px-8 lg:px-12 xl:px-16">
      <div className="grid min-h-[calc(100vh-4rem)] border border-line lg:grid-cols-[1.1fr_0.9fr]">
        <section className="border-b border-line p-6 sm:p-8 lg:border-b-0 lg:border-r lg:p-12">
          <p className="text-[11px] font-semibold uppercase tracking-[0.32em] text-muted">
            CartGO access
          </p>
          <h1 className="mt-6 text-[4rem] font-black uppercase leading-[0.88] tracking-hero sm:text-[5.4rem] xl:text-[7rem]">
            Sign in
          </h1>
          <p className="mt-6 max-w-[24rem] text-base leading-relaxed text-muted">
            Authenticate against the API gateway, load your role from the JWT-backed user service,
            and route into the correct operational workspace.
          </p>
          <div className="mt-10 space-y-3">
            <StatusBanner tone={backendStatus.online ? "success" : "error"}>
              {backendStatus.online ? "Gateway reachable on localhost:8080" : backendStatus.message}
            </StatusBanner>
            <StatusBanner>
              Verified login endpoint: POST /api/v1/user/login
            </StatusBanner>
          </div>
        </section>

        <section className="p-6 sm:p-8 lg:p-12">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">
                Email
              </label>
              <input
                type="email"
                value={form.email}
                onChange={(event) => setForm((current) => ({ ...current, email: event.target.value }))}
                className="w-full border border-line bg-transparent px-4 py-4 text-base text-paper outline-none transition-colors focus:border-paper"
                placeholder="admin@cartgo.com"
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
                placeholder="Enter password"
                required
              />
            </div>

            {error ? <StatusBanner tone="error">{error}</StatusBanner> : null}

            <button
              type="submit"
              disabled={submitting || !backendStatus.online}
              className="inline-flex items-center justify-center bg-paper px-5 py-4 text-sm font-semibold uppercase tracking-[0.18em] text-surface transition-colors duration-200 hover:bg-white disabled:cursor-not-allowed disabled:border-line disabled:bg-line"
            >
              {submitting ? "Signing in" : "Login"}
            </button>
          </form>

          <div className="mt-10 border-t border-line pt-6">
            <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-muted">
              Seeded admin
            </p>
            <p className="mt-2 text-sm text-paper">admin@cartgo.com / admin123</p>
          </div>
        </section>
      </div>
    </div>
  );
}
