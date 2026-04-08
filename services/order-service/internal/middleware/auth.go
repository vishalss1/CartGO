package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/vishalss1/CartGO/pkg/auth"
	"github.com/vishalss1/CartGO/services/order-service/internal/util"
)

type contextKey string

const (
	RoleContextKey   contextKey = "user_role"
	UserIDContextKey contextKey = "user_id"
)

func AuthMiddleware(publicKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				util.ErrorJSONResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				util.ErrorJSONResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Invalid authorization header format")
				return
			}

			claims, err := auth.ValidateToken(parts[1], publicKeys, "cartgo-order-service")
			if err != nil {
				util.ErrorJSONResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), RoleContextKey, claims.Role)
			ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDContextKey).(string)
	return userID
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleContextKey).(string)
			if !ok {
				util.ErrorJSONResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "User role not found in context")
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
				util.ErrorJSONResponse(w, http.StatusForbidden, "FORBIDDEN", "Unauthorized role for this operation")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
