package main

import (
	"log"

	"github.com/vishalss1/CartGO/services/product-service/db"
	"github.com/vishalss1/CartGO/services/product-service/internal/config"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Initialize database
	database, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}
	defer database.Close()

	log.Printf("Product service starting on port %s", cfg.Port)
}
