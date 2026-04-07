package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vishalss1/CartGO/api-gateway/internal/config"
	gwMiddleware "github.com/vishalss1/CartGO/api-gateway/internal/middleware"
	"github.com/vishalss1/CartGO/api-gateway/internal/proxy"
	"golang.org/x/time/rate"
)

type ProxyContainer struct {
	User      *proxy.Proxy
	Product   *proxy.Proxy
	Inventory *proxy.Proxy
	Order     *proxy.Proxy
	Delivery  *proxy.Proxy
	Support   *proxy.Proxy
}

func NewRouter(cfg *config.Config, proxies *ProxyContainer) *chi.Mux {
	r := chi.NewRouter()

	// 1. Logging & Recovery
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 2. Rate Limiting
	// Default: 10 requests per second with a burst of 20
	limiter := gwMiddleware.NewIPRateLimiter(rate.Limit(10), 20)
	r.Use(gwMiddleware.RateLimitMiddleware(limiter))

	// 3. Timeout Control (Hardened)
	r.Use(middleware.Timeout(15 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"OK", "service":"api-gateway"}`))
	})

	// Service Routes
	r.Group(func(r chi.Router) {
		// 4. JWT Validation & Role Injection
		r.Use(gwMiddleware.AuthMiddleware(cfg.JWTSecret))

		// 5. Routing & Proxying
		r.Mount("/api/v1/users", proxies.User.Handler())
		r.Mount("/api/v1/products", proxies.Product.Handler())
		r.Mount("/api/v1/inventory", proxies.Inventory.Handler())
		r.Mount("/api/v1/orders", proxies.Order.Handler())
		r.Mount("/api/v1/deliveries", proxies.Delivery.Handler())
		r.Mount("/api/v1/support", proxies.Support.Handler())
	})

	return r
}
