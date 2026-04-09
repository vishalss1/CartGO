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
	"github.com/vishalss1/CartGO/services/inventory-service/db"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/config"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/handler"
	customMiddleware "github.com/vishalss1/CartGO/services/inventory-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/repository"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/service"
)

func main() {
	log.Println("Starting inventory-service...")

	// Load config
	cfg := config.LoadConfig()

	// Initialize database
	database, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}
	defer database.Close()

	// Initialize repository
	inventoryRepo := repository.NewPostgresInventoryRepository(database)

	// Initialize service
	inventoryService := service.NewInventoryService(inventoryRepo)

	// Initialize handler
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Ping(); err != nil {
			log.Printf("Health check failed: DB down: %v", err)
			handler.ErrorJSONResponse(w, http.StatusInternalServerError, "DB_ERROR", "DB connectivity lost")
			return
		}
		handler.JSONResponse(w, http.StatusOK, map[string]string{"status": "OK"})
	})

	// API v1 Routes
	r.Route("/api/v1/inventory", func(r chi.Router) {
		r.Get("/{product_id}", inventoryHandler.GetInventory)

		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.AuthMiddleware(cfg.JWTPublicKeys))

			// Warehouse Staff / Admin mutations
			r.Group(func(r chi.Router) {
				r.Use(customMiddleware.RoleMiddleware("WAREHOUSE_STAFF", "ADMIN"))
				r.Post("/{product_id}/adjust", inventoryHandler.AdjustStock)
			})

			// Service Order / Admin mutations
			r.Group(func(r chi.Router) {
				r.Use(customMiddleware.RoleMiddleware("SERVICE_ORDER", "ADMIN"))
				r.Post("/{product_id}/reserve", inventoryHandler.ReserveStock)
				r.Post("/release", inventoryHandler.ReleaseStock)
				r.Post("/commit", inventoryHandler.CommitStock)
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
		log.Printf("Inventory-service is running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen and serve failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down inventory-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Inventory-service stopped.")
}
