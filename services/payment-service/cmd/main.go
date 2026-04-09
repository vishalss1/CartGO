package main

import (
	"context"
	"log/slog"
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
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting payment-service...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	conn, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Initialize repository
	paymentRepo := postgres.NewPostgresPaymentRepository(conn)

	// Initialize service
	paymentService := service.NewPaymentService(paymentRepo, logger)

	// Initialize handler
	paymentHandler := handler.NewPaymentHandler(paymentService, logger)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := conn.Ping(); err != nil {
			logger.Error("Health check failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"OK"}`))
	})

	// setup middleware for payment routes
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
		logger.Info("Payment-service is running", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Listen and serve failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down payment-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Payment-service stopped.")
}
