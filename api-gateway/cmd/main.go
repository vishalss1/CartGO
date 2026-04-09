package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vishalss1/CartGO/api-gateway/internal/config"
	"github.com/vishalss1/CartGO/api-gateway/internal/proxy"
	"github.com/vishalss1/CartGO/api-gateway/internal/router"
)

func main() {
	log.Println("Starting Hardened API Gateway...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize proxies
	userProxy, err := proxy.NewProxy(cfg.UserServiceURL)
	if err != nil {
		log.Printf("Failed to initialize user-service proxy: %v", err)
	}
	productProxy, err := proxy.NewProxy(cfg.ProductServiceURL)
	if err != nil {
		log.Printf("Failed to initialize product-service proxy: %v", err)
	}
	inventoryProxy, err := proxy.NewProxy(cfg.InventoryServiceURL)
	if err != nil {
		log.Printf("Failed to initialize inventory-service proxy: %v", err)
	}
	orderProxy, err := proxy.NewProxy(cfg.OrderServiceURL)
	if err != nil {
		log.Printf("Failed to initialize order-service proxy: %v", err)
	}
	deliveryProxy, err := proxy.NewProxy(cfg.DeliveryServiceURL)
	if err != nil {
		log.Printf("Failed to initialize delivery-service proxy: %v", err)
	}
	supportProxy, err := proxy.NewProxy(cfg.SupportServiceURL)
	if err != nil {
		log.Printf("Failed to initialize support-service proxy: %v", err)
	}
	paymentProxy, err := proxy.NewProxy(cfg.PaymentServiceURL)
	if err != nil {
		log.Printf("Failed to initialize payment-service proxy: %v", err)
	}

	proxies := &router.ProxyContainer{
		User:      userProxy,
		Product:   productProxy,
		Inventory: inventoryProxy,
		Order:     orderProxy,
		Delivery:  deliveryProxy,
		Support:   supportProxy,
		Payment:   paymentProxy,
	}

	// Setup refactored router
	r := router.NewRouter(cfg, proxies)

	// Setup server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("API Gateway is running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen and serve failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API Gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("API Gateway stopped.")
}
