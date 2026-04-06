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
	"github.com/vishalss1/CartGO/services/order-service/db"
	"github.com/vishalss1/CartGO/services/order-service/internal/client"
	"github.com/vishalss1/CartGO/services/order-service/internal/config"
	"github.com/vishalss1/CartGO/services/order-service/internal/handler"
	customMiddleware "github.com/vishalss1/CartGO/services/order-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/order-service/internal/repository/postgres"
	"github.com/vishalss1/CartGO/services/order-service/internal/service"
	"github.com/vishalss1/CartGO/services/order-service/internal/util"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting order-service...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	conn, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Initialize clients
	inventoryClient := client.NewHttpInventoryClient(cfg.InventoryServiceURL)
	paymentClient := client.NewHttpPaymentClient(cfg.PaymentServiceURL)

	// Initialize repository
	orderRepo := postgres.NewPostgresOrderRepository(conn)

	// Initialize service
	orderService := service.NewOrderService(orderRepo, inventoryClient, paymentClient, logger)

	// Initialize handler
	orderHandler := handler.NewOrderHandler(orderService, logger)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := conn.Ping(); err != nil {
			logger.Error("Health check failed", "error", err)
			util.ErrorJSONResponse(w, http.StatusInternalServerError, "DB_ERROR", "Database is down")
			return
		}
		util.JSONResponse(w, http.StatusOK, map[string]string{"status": "OK"})
	})

	// API v1 Routes
	r.Route("/api/v1/orders", func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware)

		// Create order (Customer or Admin)
		r.With(customMiddleware.RoleMiddleware("CUSTOMER", "ADMIN")).Post("/", orderHandler.CreateOrder)

		// Get order details (Open to all authenticated users for now)
		r.Get("/{id}", orderHandler.GetOrder)

		// Get orders for a user
		r.Get("/user/{user_id}", orderHandler.GetOrdersByUserID)
	})

	// Setup server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Order-service is running", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Listen and serve failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down order-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Order-service stopped.")
}
