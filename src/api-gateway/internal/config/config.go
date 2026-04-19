package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	UserServiceURL      string
	ProductServiceURL   string
	InventoryServiceURL string
	OrderServiceURL     string
	PaymentServiceURL   string
	DeliveryServiceURL  string
	SupportServiceURL   string
	RateLimit           string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Running without .env file (Docker/production mode)")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		log.Fatal("USER_SERVICE_URL is not set")
	}

	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		log.Fatal("PRODUCT_SERVICE_URL is not set")
	}

	inventoryServiceURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryServiceURL == "" {
		log.Fatal("INVENTORY_SERVICE_URL is not set")
	}

	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		log.Fatal("ORDER_SERVICE_URL is not set")
	}

	paymentServiceURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentServiceURL == "" {
		log.Fatal("PAYMENT_SERVICE_URL is not set")
	}

	deliveryServiceURL := os.Getenv("DELIVERY_SERVICE_URL")
	if deliveryServiceURL == "" {
		log.Fatal("DELIVERY_SERVICE_URL is not set")
	}

	supportServiceURL := os.Getenv("SUPPORT_SERVICE_URL")
	if supportServiceURL == "" {
		log.Fatal("SUPPORT_SERVICE_URL is not set")
	}

	rateLimit := os.Getenv("RATE_LIMIT")
	if rateLimit == "" {
		rateLimit = "100"
	}

	return &Config{
		Port:                port,
		UserServiceURL:      userServiceURL,
		ProductServiceURL:   productServiceURL,
		InventoryServiceURL: inventoryServiceURL,
		OrderServiceURL:     orderServiceURL,
		PaymentServiceURL:   paymentServiceURL,
		DeliveryServiceURL:  deliveryServiceURL,
		SupportServiceURL:   supportServiceURL,
		RateLimit:           rateLimit,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
