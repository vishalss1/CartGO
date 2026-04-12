import { API_BASE_URL, GATEWAY_HEALTH_URL } from "./constants";
import { clearStoredSession } from "./auth";

/**
 * Sanitize backend error messages to prevent exposing internal details.
 * Strips service names, endpoint paths, and technical jargon.
 */
function sanitizeErrorMessage(raw) {
  if (!raw || typeof raw !== "string") {
    return "Something went wrong. Please try again.";
  }

  const lower = raw.toLowerCase();

  // Auth-related errors — keep user-friendly
  if (lower.includes("invalid credentials") || lower.includes("unauthorized")) {
    return "Invalid email or password.";
  }
  if (lower.includes("user already exists") || lower.includes("conflict")) {
    return "An account with this email already exists.";
  }
  if (lower.includes("forbidden")) {
    return "You don't have permission to perform this action.";
  }
  if (lower.includes("not found")) {
    return "The requested resource was not found.";
  }
  if (lower.includes("validation") || lower.includes("required")) {
    return "Please check your input and try again.";
  }

  // Strip anything that looks like an internal path or service name
  if (
    /\/(api|service|internal|v1)\//i.test(raw) ||
    /(service|handler|middleware|repository)/i.test(raw) ||
    /localhost:\d+/i.test(raw) ||
    /failed to fetch/i.test(raw) ||
    /network/i.test(raw)
  ) {
    return "Something went wrong. Please try again.";
  }

  // If message is short and clean enough, return it
  if (raw.length < 120) {
    return raw;
  }

  return "Something went wrong. Please try again.";
}

async function parseResponse(response, path) {
  const contentType = response.headers.get("content-type") ?? "";
  const isJson = contentType.includes("application/json");
  const body = isJson ? await response.json() : await response.text();

  if (!response.ok) {
    // For auth endpoints (login, register), 401 means wrong credentials — show the actual error
    const isAuthEndpoint =
      typeof path === "string" &&
      (path.startsWith("/user/login") || path.startsWith("/user/register"));

    if (response.status === 401 && !isAuthEndpoint) {
      // Session expired — clear and redirect
      clearStoredSession();
      if (window.location.pathname !== "/login" && window.location.pathname !== "/register") {
        window.location.href = "/login";
      }
      throw new Error("Your session has expired. Please sign in again.");
    }

    if (response.status === 403) {
      throw new Error("You don't have permission to perform this action.");
    }

    // Backend uses { success: false, error: "message" }
    const rawMessage =
      (isJson && (body.error || body.message)) ||
      (typeof body === "string" ? body : null) ||
      `Request failed (${response.status})`;

    throw new Error(sanitizeErrorMessage(String(rawMessage)));
  }

  // Handle backend wrapper { success: true, data: ... }
  if (isJson && typeof body === "object" && body !== null && "success" in body) {
    if (body.success) {
      return body.data;
    }
    // If somehow success is false but status was 2xx (rare, but handle it)
    throw new Error(sanitizeErrorMessage(String(body.error || "Request failed")));
  }

  return body;
}

export async function apiRequest(path, options = {}) {
  let response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      headers: {
        Accept: "application/json",
        ...(options.body ? { "Content-Type": "application/json" } : {}),
        ...(options.headers ?? {}),
      },
    });
  } catch {
    throw new Error("Unable to connect. Please check your internet connection.");
  }

  return parseResponse(response, path);
}

export async function authenticatedRequest(path, token, options = {}) {
  return apiRequest(path, {
    ...options,
    headers: {
      Authorization: `Bearer ${token}`,
      ...(options.headers ?? {}),
    },
  });
}

export async function fetchBackendHealth() {
  const response = await fetch(GATEWAY_HEALTH_URL);
  return parseResponse(response, "/health");
}

// ── Auth ──

export async function loginRequest(credentials) {
  return apiRequest("/user/login", {
    method: "POST",
    body: JSON.stringify(credentials),
  });
}

export async function registerRequest(data) {
  return apiRequest("/user/register", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function meRequest(token) {
  return authenticatedRequest("/user/me", token);
}

// ── Support messages ──

export async function getTicketMessages(ticketId, token) {
  return authenticatedRequest(`/support/tickets/${ticketId}/messages`, token);
}

export async function addTicketMessage(ticketId, content, token) {
  return authenticatedRequest(`/support/tickets/${ticketId}/messages`, token, {
    method: "POST",
    body: JSON.stringify({ content }),
  });
}
