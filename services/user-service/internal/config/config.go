package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	Port               string
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  string
	RefreshTokenExpiry string
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

	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	if accessTokenSecret == "" {
		log.Fatal("ACCESS_TOKEN_SECRET is not set")
	}

	refreshTokenSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	if refreshTokenSecret == "" {
		log.Fatal("REFRESH_TOKEN_SECRET is not set")
	}

	accessTokenExpiry := os.Getenv("ACCESS_TOKEN_EXPIRY")
	if accessTokenExpiry == "" {
		accessTokenExpiry = "15m"
	}

	refreshTokenExpiry := os.Getenv("REFRESH_TOKEN_EXPIRY")
	if refreshTokenExpiry == "" {
		refreshTokenExpiry = "7d"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL:        databaseURL,
		Port:               port,
		AccessTokenSecret:  accessTokenSecret,
		RefreshTokenSecret: refreshTokenSecret,
		AccessTokenExpiry:  accessTokenExpiry,
		RefreshTokenExpiry: refreshTokenExpiry,
	}
}
