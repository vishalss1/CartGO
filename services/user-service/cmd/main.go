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
	"github.com/vishalss1/CartGO/services/user-service/db"
	"github.com/vishalss1/CartGO/services/user-service/internal/config"
	"github.com/vishalss1/CartGO/services/user-service/internal/handler"
	customMiddleware "github.com/vishalss1/CartGO/services/user-service/internal/middleware"
	"github.com/vishalss1/CartGO/services/user-service/internal/repository"
	"github.com/vishalss1/CartGO/services/user-service/internal/service"
	"github.com/vishalss1/CartGO/pkg/util"
)

func main() {
	log.Println("Starting user-service...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	conn, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer conn.Close()

	// Initialize repositories
	userRepo := repository.NewPostgresUserRepository(conn)
	tokenRepo := repository.NewPostgresRefreshTokenRepository(conn)

	// Initialize service
	userService := service.NewUserService(userRepo, tokenRepo, cfg)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)

	// Setup router
	r := chi.NewRouter()
	r.Use(util.CorrelationIDMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, map[string]string{"status": "OK"})
	})

	r.Get("/api/v1/user/health", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, map[string]string{"status": "OK"})
	})

	r.Route("/api/v1/user", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)
		r.Post("/refresh", userHandler.Refresh)
		r.Post("/logout", userHandler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.AuthMiddleware(cfg.JWTPublicKeys))
			r.Get("/me", userHandler.Me)
			r.Patch("/me", userHandler.UpdateMe)

			r.Group(func(r chi.Router) {
				r.Use(customMiddleware.RoleMiddleware("ADMIN"))
				r.Get("/admin/users", userHandler.AdminListUsers)
				r.Patch("/admin/users/{id}/role", userHandler.AdminUpdateRole)
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
		log.Printf("User-service is running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen and serve failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down user-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("User-service stopped.")
}
