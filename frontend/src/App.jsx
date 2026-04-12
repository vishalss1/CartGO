import { Route, Routes } from "react-router-dom";
import ProtectedRoute from "./components/ProtectedRoute";
import AdminPage from "./pages/Admin";
import DeliveryPage from "./pages/Delivery";
import InventoryPage from "./pages/Inventory";
import LoginPage from "./pages/Login";
import RegisterPage from "./pages/Register";
import RoleRedirectPage from "./pages/RoleRedirect";
import ShopPage from "./pages/Shop";
import SupportPage from "./pages/Support";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<RoleRedirectPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route element={<ProtectedRoute allowedRole="CUSTOMER" />}>
        <Route path="/shop" element={<ShopPage />} />
      </Route>
      <Route element={<ProtectedRoute allowedRole="WAREHOUSE_STAFF" />}>
        <Route path="/inventory" element={<InventoryPage />} />
      </Route>
      <Route element={<ProtectedRoute allowedRole="DELIVERY_PARTNER" />}>
        <Route path="/delivery" element={<DeliveryPage />} />
      </Route>
      <Route element={<ProtectedRoute allowedRole="ADMIN" />}>
        <Route path="/admin" element={<AdminPage />} />
      </Route>
      <Route element={<ProtectedRoute allowedRole="SUPPORT_AGENT" />}>
        <Route path="/support" element={<SupportPage />} />
      </Route>
      <Route path="*" element={<RoleRedirectPage />} />
    </Routes>
  );
}
