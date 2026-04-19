package middleware

import (
	"net/http"

	"github.com/vishalss1/CartGO/pkg/auth"
)

// SanitizeHeadersMiddleware aggressively strips internal headers to prevent spoofing
func SanitizeHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip internal identity headers
		r.Header.Del(auth.HeaderUserID)
		r.Header.Del(auth.HeaderUserRole)

		// Strip potential spoofing forwards if gateway handles SSL termination exclusively
		// In a real cloud setup, LB handles X-Forwarded-For, but good to ensure
		// downstream services only trust what the Gateway allows.
		// We leave Authorization intact for backend RS256 validation.

		next.ServeHTTP(w, r)
	})
}
