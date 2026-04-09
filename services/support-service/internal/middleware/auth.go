package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vishalss1/CartGO/pkg/auth"
	"github.com/vishalss1/CartGO/services/support-service/internal/model"
)

type contextKey string

const (
	RoleContextKey   contextKey = "user_role"
	UserIDContextKey contextKey = "user_id"
)

func errorResponse(w http.ResponseWriter, code int, errCode string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(model.ErrorResponse{
		Error:   errCode,
		Message: message,
	})
}

func AuthMiddleware(publicKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Check for Internal Identity Headers (Trust because Gateway has stripped them from external traffic)
			internalUserID := r.Header.Get(auth.HeaderUserID)
			internalRole := r.Header.Get(auth.HeaderUserRole)

			if internalRole != "" {
				ctx := context.WithValue(r.Context(), RoleContextKey, internalRole)
				ctx = context.WithValue(ctx, UserIDContextKey, internalUserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 2. Fallback to JWT Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errorResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				errorResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Invalid authorization header format")
				return
			}

			claims, err := auth.ValidateToken(parts[1], publicKeys, "cartgo-api")
			if err != nil {
				errorResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Invalid token")
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

func GetRole(ctx context.Context) string {
	role, _ := ctx.Value(RoleContextKey).(string)
	return role
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetRole(r.Context())
			if role == "" {
				errorResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "User role not found in context")
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
				errorResponse(w, http.StatusForbidden, "FORBIDDEN", "Unauthorized role for this operation")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
