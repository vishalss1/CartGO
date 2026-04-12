export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api/v1";

export const GATEWAY_HEALTH_URL =
  import.meta.env.VITE_GATEWAY_HEALTH_URL ?? "http://localhost:8080/health";

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
