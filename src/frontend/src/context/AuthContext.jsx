import { createContext, useContext, useEffect, useMemo, useState } from "react";
import {
  clearStoredSession,
  getStoredSession,
  storeSession,
} from "../utils/auth";
import { fetchBackendHealth, loginRequest, registerRequest, meRequest } from "../utils/api";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [{ token, refreshToken, user }, setSession] = useState(getStoredSession);
  const [booting, setBooting] = useState(true);
  const [backendStatus, setBackendStatus] = useState({
    checked: false,
    online: false,
    message: "Connecting to services",
  });

  useEffect(() => {
    let active = true;

    async function bootstrap() {
      try {
        await fetchBackendHealth();
        if (!active) {
          return;
        }
        setBackendStatus({
          checked: true,
          online: true,
          message: "Services online",
        });
      } catch (error) {
        if (!active) {
          return;
        }
        setBackendStatus({
          checked: true,
          online: false,
          message: "Service temporarily unavailable",
        });
      }

      if (!token) {
        if (active) {
          setBooting(false);
        }
        return;
      }

      try {
        const currentUser = await meRequest(token);
        if (!active) {
          return;
        }
        const nextSession = { token, refreshToken, user: currentUser };
        storeSession(nextSession);
        setSession(nextSession);
      } catch {
        if (!active) {
          return;
        }
        clearStoredSession();
        setSession({ token: "", refreshToken: "", user: null });
      } finally {
        if (active) {
          setBooting(false);
        }
      }
    }

    bootstrap();

    return () => {
      active = false;
    };
  }, [refreshToken, token]);
  
  useEffect(() => {
    const handleRefresh = (e) => {
      setSession(e.detail);
    };
    
    window.addEventListener("cartgo-auth-refresh", handleRefresh);
    return () => window.removeEventListener("cartgo-auth-refresh", handleRefresh);
  }, []);

  const value = useMemo(
    () => ({
      token,
      refreshToken,
      user,
      role: user?.role ?? null,
      isAuthenticated: Boolean(token && user),
      booting,
      backendStatus,
      async login(credentials) {
        const response = await loginRequest(credentials);
        const nextSession = {
          token: response.access_token,
          refreshToken: response.refresh_token,
          user: response.user,
        };
        storeSession(nextSession);
        setSession(nextSession);
        return response.user;
      },
      async register(data) {
        const response = await registerRequest(data);
        return response;
      },
      logout() {
        clearStoredSession();
        setSession({ token: "", refreshToken: "", user: null });
      },
    }),
    [backendStatus, booting, refreshToken, token, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
