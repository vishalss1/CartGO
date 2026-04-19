// Re-export from centralized config (single source of truth)
export { API_BASE_URL, GATEWAY_HEALTH_URL, WS_URL } from "../config/api";


export const ROLE_TO_ROUTE = {
  CUSTOMER: "/shop",
  WAREHOUSE_STAFF: "/inventory",
  DELIVERY_PARTNER: "/delivery",
  ADMIN: "/admin",
  SUPPORT_AGENT: "/support",
};

export const VALID_ROLES = Object.keys(ROLE_TO_ROUTE);

export const TICKET_STATUSES = [
  "OPEN",
  "IN_PROGRESS",
  "RESOLVED",
  "CLOSED",
  "REOPENED",
];

export const DELIVERY_STATUSES = [
  "PENDING",
  "PICKED_UP",
  "DELIVERED",
  "CANCELLED",
];
