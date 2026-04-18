package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/auth"
)

type contextKey string

const (
	RoleContextKey   contextKey = "user_role"
	UserIDContextKey contextKey = "user_id"
)

func AuthMiddleware(publicKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Check for Internal Identity Headers (Trust because Gateway has stripped them from external traffic)
			internalUserIDStr := r.Header.Get(auth.HeaderUserID)
			internalRole := r.Header.Get(auth.HeaderUserRole)

			if internalRole != "" {
				userID, err := uuid.Parse(internalUserIDStr)
				if err != nil {
					http.Error(w, "Unauthorized: Invalid internal user id", http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), RoleContextKey, internalRole)
				ctx = context.WithValue(ctx, UserIDContextKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 2. Fallback to JWT Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized: Missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Unauthorized: Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			claims, err := auth.ValidateToken(parts[1], publicKeys, "cartgo-delivery-service")
			if err != nil {
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				http.Error(w, "Unauthorized: Invalid token subject", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), RoleContextKey, claims.Role)
			ctx = context.WithValue(ctx, UserIDContextKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) uuid.UUID {
	userID, _ := ctx.Value(UserIDContextKey).(uuid.UUID)
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
				http.Error(w, "Unauthorized: User role not found in context", http.StatusUnauthorized)
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
				http.Error(w, "Forbidden: Unauthorized role for this operation", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
