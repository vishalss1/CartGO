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
	"github.com/vishalss1/CartGO/services/payment-service/db"
	"github.com/vishalss1/CartGO/services/payment-service/internal/config"
	"github.com/vishalss1/CartGO/services/payment-service/internal/handler"
	authmw "github.com/vishalss1/CartGO/services/payment-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/payment-service/internal/repository/postgres"
	"github.com/vishalss1/CartGO/services/payment-service/internal/service"
	"github.com/vishalss1/CartGO/pkg/util"
)

func main() {
	log.Println("Starting payment-service...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	conn, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer conn.Close()

	// Initialize repository
	paymentRepo := postgres.NewPostgresPaymentRepository(conn)

	// Initialize service
	paymentService := service.NewPaymentService(paymentRepo)

	// Initialize handler
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// Setup router
	r := chi.NewRouter()
	r.Use(util.CorrelationIDMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

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

	// setup middleware
	authProvider := authmw.AuthMiddleware(cfg.JWTPublicKeys)

	// API v1 Routes
	r.Route("/api/v1/payments", func(r chi.Router) {
		r.Use(authProvider)
		r.Post("/", paymentHandler.ProcessPayment)
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
		log.Printf("Payment-service is running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen and serve failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down payment-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Payment-service stopped.")
}
