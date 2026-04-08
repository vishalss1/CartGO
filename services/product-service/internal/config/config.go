package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/vishalss1/CartGO/pkg/auth"
)

type Config struct {
	DatabaseURL   string
	Port          string
	JWTPublicKeys map[string]string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
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
		DatabaseURL:   databaseURL,
		Port:          port,
		JWTPublicKeys: publicKeys,
	}
}
