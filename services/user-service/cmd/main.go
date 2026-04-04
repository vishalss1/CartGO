package main

import (
	"log"

	"github.com/vishalss1/CartGO/services/user-service/db"
	"github.com/vishalss1/CartGO/services/user-service/internal/config"
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

	log.Printf("User-service is running on port %s", cfg.Port)

	// In a real application, we would start an HTTP server here.
	// For now, we'll just log that it's running.
}
