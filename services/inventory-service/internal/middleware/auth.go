package middleware

import (
	"context"
	"net/http"

	"github.com/vishalss1/CartGO/services/inventory-service/internal/handler"
)

type contextKey string

const RoleContextKey contextKey = "user_role"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// IMPORTANT: This header MUST be injected by a trusted upstream (API Gateway).
		// Public access to this service should be restricted.
		role := r.Header.Get("X-User-Role")
		if role == "" {
			handler.ErrorJSONResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Missing X-User-Role header")
			return
		}

		ctx := context.WithValue(r.Context(), RoleContextKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleContextKey).(string)
			if !ok {
				handler.ErrorJSONResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "User role not found in context")
				return
			}

			isAllowed := false
			for _, allowed := range allowedRoles {
				if role == allowed {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				handler.ErrorJSONResponse(w, http.StatusForbidden, "FORBIDDEN", "Unauthorized role for this operation")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
