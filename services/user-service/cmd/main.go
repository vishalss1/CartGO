package main

import (
	"log"

	"github.com/vishalss1/CartGO/services/user-service/db"
	"github.com/vishalss1/CartGO/services/user-service/internal/config"
	"github.com/vishalss1/CartGO/services/user-service/internal/repository"
	"github.com/vishalss1/CartGO/services/user-service/internal/service"
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

	log.Printf("User-service initialized and running on port %s", cfg.Port)

	// In a real application, we would initialize handlers and start an HTTP server here.
	_ = userService // Avoid unused variable warning
}
