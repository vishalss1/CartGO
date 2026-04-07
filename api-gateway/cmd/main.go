package main

import (
	"context"
	"log/slog"
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
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting Hardened API Gateway...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize proxies
	userProxy, _ := proxy.NewProxy(cfg.UserServiceURL)
	productProxy, _ := proxy.NewProxy(cfg.ProductServiceURL)
	inventoryProxy, _ := proxy.NewProxy(cfg.InventoryServiceURL)
	orderProxy, _ := proxy.NewProxy(cfg.OrderServiceURL)
	deliveryProxy, _ := proxy.NewProxy(cfg.DeliveryServiceURL)
	supportProxy, _ := proxy.NewProxy(cfg.SupportServiceURL)

	proxies := &router.ProxyContainer{
		User:      userProxy,
		Product:   productProxy,
		Inventory: inventoryProxy,
		Order:     orderProxy,
		Delivery:  deliveryProxy,
		Support:   supportProxy,
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
		logger.Info("API Gateway is running", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Listen and serve failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API Gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("API Gateway stopped.")
}
