import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { ROLE_TO_ROUTE } from "../utils/constants";

export default function ProtectedRoute({ allowedRole }) {
  const { booting, isAuthenticated, role } = useAuth();
  const location = useLocation();

  if (booting) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-surface px-6 text-paper">
        <div className="border border-line px-6 py-5 text-sm font-semibold uppercase tracking-[0.24em]">
          Booting CartGO
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  if (allowedRole && role !== allowedRole) {
    return <Navigate to={ROLE_TO_ROUTE[role]} replace />;
  }

  return <Outlet />;
}
