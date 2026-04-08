package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/vishalss1/CartGO/pkg/auth"
	"github.com/vishalss1/CartGO/services/product-service/internal/handler"
)

type contextKey string

const RoleContextKey contextKey = "user_role"

func AuthMiddleware(publicKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handler.ErrorJSONResponse(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				handler.ErrorJSONResponse(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			claims, err := auth.ValidateToken(parts[1], publicKeys, "cartgo-product-service")
			if err != nil {
				handler.ErrorJSONResponse(w, http.StatusUnauthorized, "Invalid token")
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
				handler.ErrorJSONResponse(w, http.StatusUnauthorized, "User role not found in context")
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
				handler.ErrorJSONResponse(w, http.StatusForbidden, "Unauthorized role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
