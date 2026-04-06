package middleware

import (
	"context"
	"net/http"

	"github.com/vishalss1/CartGO/services/delivery-service/internal/util"
)

type contextKey string

const (
	RoleContextKey   contextKey = "user_role"
	UserIDContextKey contextKey = "user_id"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("X-User-Role")
		userID := r.Header.Get("X-User-ID")

		if role == "" {
			util.ErrorJSONResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Missing X-User-Role header")
			return
		}

		ctx := context.WithValue(r.Context(), RoleContextKey, role)
		if userID != "" {
			ctx = context.WithValue(ctx, UserIDContextKey, userID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDContextKey).(string)
	return userID
}

func GetRole(ctx context.Context) string {
	role, _ := ctx.Value(RoleContextKey).(string)
	return role
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetRole(r.Context())
			if role == "" {
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
