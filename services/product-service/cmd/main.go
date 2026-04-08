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
	"github.com/vishalss1/CartGO/services/product-service/db"
	"github.com/vishalss1/CartGO/services/product-service/internal/config"
	"github.com/vishalss1/CartGO/services/product-service/internal/handler"
	customMiddleware "github.com/vishalss1/CartGO/services/product-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/product-service/internal/repository"
	"github.com/vishalss1/CartGO/services/product-service/internal/service"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting product-service...")

	// Load config
	cfg := config.LoadConfig()

	// Initialize database
	database, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Could not initialize database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize repository
	prodRepo := repository.NewPostgresProductRepository(database)

	// Initialize service
	prodService := service.NewProductService(prodRepo)

	// Initialize handler
	prodHandler := handler.NewProductHandler(prodService, logger)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check (depth: verifies DB)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Ping(); err != nil {
			logger.Error("Health check failed: DB down", "error", err)
			handler.ErrorJSONResponse(w, http.StatusInternalServerError, "DB connectivity lost")
			return
		}
		handler.JSONResponse(w, http.StatusOK, map[string]string{"status": "OK"})
	})

	// API v1 Routes
	r.Route("/api/v1/products", func(r chi.Router) {
		// Public routes
		r.Get("/", prodHandler.ListProducts)
		r.Get("/{id}", prodHandler.GetProduct)

		// Protected routes (X-User-Role required -> now Authorization header)
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.AuthMiddleware(cfg.JWTPublicKeys))
			
			// Admin only mutations
			r.Group(func(r chi.Router) {
				r.Use(customMiddleware.RoleMiddleware("ADMIN"))
				r.Post("/", prodHandler.CreateProduct)
				r.Patch("/{id}", prodHandler.UpdateProduct)
				r.Delete("/{id}", prodHandler.DeleteProduct)
			})
		})
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
		logger.Info("Product-service is running", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Listen and serve failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down product-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Product-service stopped.")
}
