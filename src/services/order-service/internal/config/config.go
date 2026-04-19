package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/vishalss1/CartGO/pkg/auth"
)

type Config struct {
	DatabaseURL         string
	Port                string
	InventoryServiceURL string
	PaymentServiceURL   string
	ProductServiceURL   string
	DeliveryServiceURL  string
	JWTPublicKeys       map[string]string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Running without .env file (Docker/production mode)")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084" // Default port for order-service
	}

	inventoryServiceURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryServiceURL == "" {
		inventoryServiceURL = "http://localhost:8083"
	}

	paymentServiceURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentServiceURL == "" {
		paymentServiceURL = "http://localhost:8085"
	}

	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		productServiceURL = "http://localhost:8082"
	}

	deliveryServiceURL := os.Getenv("DELIVERY_SERVICE_URL")
	if deliveryServiceURL == "" {
		deliveryServiceURL = "http://localhost:8086"
	}

	keysDir := os.Getenv("KEYS_DIR")
	if keysDir == "" {
		keysDir = "keys"
		if _, err := os.Stat(keysDir); os.IsNotExist(err) {
			if _, err := os.Stat("../../keys"); err == nil {
				keysDir = "../../keys"
			}
		}
	}

	keysJSON := os.Getenv("JWT_PUBLIC_KEYS")
	publicKeys, err := auth.LoadPublicKeys(keysJSON, keysDir)
	if err != nil {
		log.Printf("Warning: Failed to load public keys: %v\n", err)
	}

	return &Config{
		DatabaseURL:         databaseURL,
		Port:                port,
		InventoryServiceURL: inventoryServiceURL,
		PaymentServiceURL:   paymentServiceURL,
		ProductServiceURL:   productServiceURL,
		DeliveryServiceURL:  deliveryServiceURL,
		JWTPublicKeys:       publicKeys,
	}
}
