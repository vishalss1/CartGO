import { Navigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { ROLE_TO_ROUTE } from "../utils/constants";

export default function RoleRedirectPage() {
  const { booting, isAuthenticated, role } = useAuth();

  if (booting) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-surface px-6 text-paper">
        <div className="border border-line px-6 py-5 text-sm font-semibold uppercase tracking-[0.24em]">
          Loading role
        </div>
      </div>
    );
  }

  if (!isAuthenticated || !role) {
    return <Navigate to="/login" replace />;
  }

  return <Navigate to={ROLE_TO_ROUTE[role]} replace />;
}
