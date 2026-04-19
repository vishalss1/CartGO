/**
 * Centralized API Configuration
 *
 * All API endpoints and service URLs are defined here.
 * Values are read from environment variables with safe fallbacks.
 */

/** Base URL for all API requests (proxied via Vite in development) */
export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "/api/v1";

/** Gateway health-check endpoint */
export const GATEWAY_HEALTH_URL =
  import.meta.env.VITE_GATEWAY_HEALTH_URL || "/health";

/**
 * WebSocket URL for real-time features.
 * Falls back to a safe default that uses the current host.
 */
export const WS_URL =
  import.meta.env.VITE_WS_URL ||
  `ws://${window.location.hostname}:8080`;
