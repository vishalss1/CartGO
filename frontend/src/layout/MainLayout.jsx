import Navbar from "../components/Navbar";
import StatusBanner from "../components/StatusBanner";
import { useAuth } from "../context/AuthContext";

export default function MainLayout({ children }) {
  const { backendStatus } = useAuth();

  return (
    <div id="top" className="min-h-screen bg-surface font-sans text-paper antialiased">
      <Navbar />
      <main className="px-5 py-6 sm:px-8 lg:px-12 xl:px-16">
        {!backendStatus.online ? (
          <div className="mb-6">
            <StatusBanner tone="error">
              Gateway check failed: {backendStatus.message}
            </StatusBanner>
          </div>
        ) : null}
        {children}
      </main>
    </div>
  );
}
