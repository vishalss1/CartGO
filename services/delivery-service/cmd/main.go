package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vishalss1/CartGO/services/delivery-service/db"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/config"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/handler"
	deliveryMiddleware "github.com/vishalss1/CartGO/services/delivery-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/repository/postgres"
	"github.com/vishalss1/CartGO/services/delivery-service/internal/service"
)

func main() {
	log.Println("Starting delivery-service...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	conn, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer conn.Close()

	// Initialize repository
	deliveryRepo := postgres.NewPostgresDeliveryRepository(conn)

	// Initialize service
	deliveryService := service.NewDeliveryService(deliveryRepo)

	// Initialize handler
	deliveryHandler := handler.NewDeliveryHandler(deliveryService)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(deliveryMiddleware.AuthMiddleware(cfg.JWTPublicKeys))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := conn.Ping(); err != nil {
			log.Printf("Health check failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"OK"}`))
	})

	// API v1 Routes
	r.Route("/api/v1/deliveries", func(r chi.Router) {
		r.With(deliveryMiddleware.RoleMiddleware("CUSTOMER", "ADMIN", "SERVICE_ORDER")).Post("/", deliveryHandler.CreateDelivery)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", deliveryHandler.GetDelivery)
			r.With(deliveryMiddleware.RoleMiddleware("DELIVERY_PARTNER", "ADMIN")).
				Patch("/status", deliveryHandler.UpdateDeliveryStatus)
		})

		r.Get("/partner/{partner_id}", deliveryHandler.ListDeliveriesByPartner)
	})

	// Setup server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Delivery-service is running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen and serve failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down delivery-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Delivery-service stopped.")
}
