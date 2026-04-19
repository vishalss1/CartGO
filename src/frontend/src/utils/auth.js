const TOKEN_KEY = "cartgo_access_token";
const REFRESH_KEY = "cartgo_refresh_token";
const USER_KEY = "cartgo_user";
const CART_KEY = "cartgo_customer_cart";

export function getStoredSession() {
  try {
    const token = localStorage.getItem(TOKEN_KEY) ?? "";
    const refreshToken = localStorage.getItem(REFRESH_KEY) ?? "";
    const rawUser = localStorage.getItem(USER_KEY);
    return {
      token,
      refreshToken,
      user: rawUser ? JSON.parse(rawUser) : null,
    };
  } catch {
    return { token: "", refreshToken: "", user: null };
  }
}

export function storeSession({ token, refreshToken, user }) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(REFRESH_KEY, refreshToken ?? "");
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function clearStoredSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_KEY);
  localStorage.removeItem(USER_KEY);
}

export function getStoredCart() {
  try {
    const raw = localStorage.getItem(CART_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

export function storeCart(items) {
  localStorage.setItem(CART_KEY, JSON.stringify(items));
}
