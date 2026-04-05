package main

import (
	"log"

	"github.com/vishalss1/CartGO/services/product-service/db"
	"github.com/vishalss1/CartGO/services/product-service/internal/config"
	"github.com/vishalss1/CartGO/services/product-service/internal/repository"
	"github.com/vishalss1/CartGO/services/product-service/internal/service"
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

	// Initialize repository
	prodRepo := repository.NewPostgresProductRepository(database)

	// Initialize service
	prodService := service.NewProductService(prodRepo)

	log.Printf("Product service initialized with repository and service layers")
	log.Printf("Product service starting on port %s", cfg.Port)

	// (Temporary: Placeholder for router and server)
	_ = prodService
}
