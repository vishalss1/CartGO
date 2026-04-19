package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/vishalss1/CartGO/pkg/auth"
	"github.com/vishalss1/CartGO/pkg/util"
)

type contextKey string

const RoleContextKey contextKey = "user_role"

func AuthMiddleware(publicKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Check for Internal Identity Headers (Trust because Gateway has stripped them from external traffic)
			internalUserID := r.Header.Get(auth.HeaderUserID)
			internalRole := r.Header.Get(auth.HeaderUserRole)

			if internalRole != "" {
				ctx := context.WithValue(r.Context(), RoleContextKey, internalRole)
				if internalUserID != "" {
				    // Optionally propagate UserID if needed for the operation
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 2. Fallback to JWT Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				util.WriteError(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				util.WriteError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			claims, err := auth.ValidateToken(parts[1], publicKeys, "cartgo-inventory-service")
			if err != nil {
				util.WriteError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), RoleContextKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleContextKey).(string)
			if !ok {
				util.WriteError(w, http.StatusUnauthorized, "User role not found in context")
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
				util.WriteError(w, http.StatusForbidden, "Unauthorized role for this operation")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
