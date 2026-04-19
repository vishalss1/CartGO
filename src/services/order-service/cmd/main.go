package main

import (
	"context"
	"log"
	"fmt"
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
	"github.com/vishalss1/CartGO/pkg/util"
)

func main() {
	log.Println("Starting order-service...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	conn, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer conn.Close()

	// Initialize clients
	inventoryClient := client.NewHttpInventoryClient(cfg.InventoryServiceURL)
	paymentClient := client.NewHttpPaymentClient(cfg.PaymentServiceURL)
	productClient := client.NewHttpProductClient(cfg.ProductServiceURL)
	deliveryClient := client.NewHttpDeliveryClient(cfg.DeliveryServiceURL)

	// Initialize repository
	orderRepo := postgres.NewPostgresOrderRepository(conn)

	orderService := service.NewOrderService(orderRepo, inventoryClient, paymentClient, productClient, deliveryClient)

	// Initialize handler
	orderHandler := handler.NewOrderHandler(orderService)

	// Setup router
	r := chi.NewRouter()
	r.Use(util.CorrelationIDMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"OK","service":"%s","timestamp":"%s"}`, "order-service", time.Now().Format(time.RFC3339))
	})

	// API v1 Routes
	r.Route("/api/v1/orders", func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(cfg.JWTPublicKeys))

		// Global orders list (Admin only)
		r.With(customMiddleware.RoleMiddleware("ADMIN")).Get("/", orderHandler.ListAllOrders)

		// Create order (Customer or Admin)
		r.With(customMiddleware.RoleMiddleware("CUSTOMER", "ADMIN")).Post("/", orderHandler.CreateOrder)

		// Get order details
		r.Get("/{id}", orderHandler.GetOrder)

		// Get orders for a user
		r.Get("/user/{user_id}", orderHandler.GetOrdersByUserID)

		// Confirm order after payment retry
		r.Post("/{id}/confirm-after-payment", orderHandler.ConfirmAfterPayment)
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
		log.Printf("Order-service is running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen and serve failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down order-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Order-service stopped.")
}
