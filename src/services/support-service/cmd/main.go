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
	"github.com/vishalss1/CartGO/services/support-service/db"
	"github.com/vishalss1/CartGO/services/support-service/internal/config"
	"github.com/vishalss1/CartGO/services/support-service/internal/handler"
	customMiddleware "github.com/vishalss1/CartGO/services/support-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/support-service/internal/repository"
	"github.com/vishalss1/CartGO/services/support-service/internal/service"
)

func main() {
	log.Println("Starting support-service...")

	cfg := config.LoadConfig()

	// Initialize Database
	conn, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer conn.Close()

	// Initialize Layers
	repo := repository.NewRepository(conn)
	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	// Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"OK","service":"%s","timestamp":"%s"}`, "support-service", time.Now().Format(time.RFC3339))
	})

	r.Route("/api/v1/support", func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(cfg.JWTPublicKeys))

		r.Post("/tickets", h.CreateTicket)
		r.Get("/tickets", h.ListTickets)
		r.Route("/tickets/{id}", func(r chi.Router) {
			r.Get("/", h.GetTicket)
			r.Patch("/status", h.UpdateStatus)
			r.Patch("/assign", h.AssignTicket)
			r.Post("/messages", h.AddMessage)
			r.Get("/messages", h.ListMessages)
		})
	})

	// Server Configuration
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful Shutdown
	go func() {
		log.Printf("Support Serivce is running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen and serve failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down support-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("support-service stopped.")
}
