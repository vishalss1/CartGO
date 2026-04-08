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
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Port:                getEnv("PORT", "8080"),
		UserServiceURL:      getEnv("USER_SERVICE_URL", ""),
		ProductServiceURL:   getEnv("PRODUCT_SERVICE_URL", ""),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", ""),
		OrderServiceURL:     getEnv("ORDER_SERVICE_URL", ""),
		PaymentServiceURL:   getEnv("PAYMENT_SERVICE_URL", ""),
		DeliveryServiceURL:  getEnv("DELIVERY_SERVICE_URL", ""),
		SupportServiceURL:   getEnv("SUPPORT_SERVICE_URL", ""),
		RateLimit:           getEnv("RATE_LIMIT", "100"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
