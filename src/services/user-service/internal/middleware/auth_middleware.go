package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/user-service/internal/auth"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func AuthMiddleware(publicKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				util.WriteError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				util.WriteError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			claims, err := auth.ValidateToken(parts[1], publicKeys)
			if err != nil {
				util.WriteError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				util.WriteError(w, http.StatusUnauthorized, "invalid token subject")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleMiddleware(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(RoleKey).(string)
			if !ok {
				util.WriteError(w, http.StatusForbidden, "user role not found")
				return
			}

			authorized := false
			for _, role := range requiredRoles {
				if userRole == role {
					authorized = true
					break
				}
			}

			if !authorized {
				util.WriteError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
