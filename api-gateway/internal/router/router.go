package router

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vishalss1/CartGO/api-gateway/internal/config"
	gwMiddleware "github.com/vishalss1/CartGO/api-gateway/internal/middleware"
	"github.com/vishalss1/CartGO/api-gateway/internal/proxy"
	"github.com/vishalss1/CartGO/pkg/util"
	"golang.org/x/time/rate"
)

type ProxyContainer struct {
	User      *proxy.Proxy
	Product   *proxy.Proxy
	Inventory *proxy.Proxy
	Order     *proxy.Proxy
	Delivery  *proxy.Proxy
	Support   *proxy.Proxy
	Payment   *proxy.Proxy
}

func NewRouter(cfg *config.Config, proxies *ProxyContainer) *chi.Mux {
	r := chi.NewRouter()

	// 1. Correlation ID for tracing
	r.Use(util.CorrelationIDMiddleware)

	// 2. Logging & Recovery
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 2. Header Sanitization (Move to top for security)
	// Prevents external users from spoofing internal identity headers
	r.Use(gwMiddleware.SanitizeHeadersMiddleware)

	// 3. Rate Limiting
	// Default: 10 requests per second with a burst of 20
	// Now safe because spoofed X-User-ID headers have been stripped
	limiter := gwMiddleware.NewIPRateLimiter(rate.Limit(10), 20)
	r.Use(gwMiddleware.RateLimitMiddleware(limiter))

	// 4. Timeout Control (Hardened)
	r.Use(middleware.Timeout(15 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"OK","service":"%s","timestamp":"%s"}`, "api-gateway", time.Now().Format(time.RFC3339))
	})

	// Service Routes
	r.Group(func(r chi.Router) {
		// 5. Routing & Proxying
		// Using Handle with wildcard to ensure the full path (including prefix)
		// is forwarded to services that already expect it.
		r.Handle("/api/v1/user*", proxies.User.Handler())
		r.Handle("/api/v1/products*", proxies.Product.Handler())
		r.Handle("/api/v1/inventory*", proxies.Inventory.Handler())
		r.Handle("/api/v1/orders*", proxies.Order.Handler())
		r.Handle("/api/v1/deliveries*", proxies.Delivery.Handler())
		r.Handle("/api/v1/support*", proxies.Support.Handler())
		r.Handle("/api/v1/payments*", proxies.Payment.Handler())
	})

	return r
}
