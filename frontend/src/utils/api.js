import { API_BASE_URL, GATEWAY_HEALTH_URL } from "./constants";

async function parseResponse(response) {
  const contentType = response.headers.get("content-type") ?? "";
  const isJson = contentType.includes("application/json");
  const body = isJson ? await response.json() : await response.text();

  if (!response.ok) {
    const message =
      (isJson && (body.message || body.error)) ||
      body ||
      `Request failed with status ${response.status}`;
    throw new Error(message);
  }

  return body;
}

export async function apiRequest(path, options = {}) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      Accept: "application/json",
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...(options.headers ?? {}),
    },
  });

  return parseResponse(response);
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
  return parseResponse(response);
}

export async function loginRequest(credentials) {
  return apiRequest("/user/login", {
    method: "POST",
    body: JSON.stringify(credentials),
  });
}

export async function meRequest(token) {
  return authenticatedRequest("/user/me", token);
}
