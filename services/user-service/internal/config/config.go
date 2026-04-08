package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/vishalss1/CartGO/pkg/auth"
)

type Config struct {
	DatabaseURL        string
	Port               string
	JWTPrivateKey      string
	JWTPrivateKeyID    string
	JWTPublicKeys      map[string]string
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
		port = "8081"
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

	jwtPrivateKey := os.Getenv("JWT_PRIVATE_KEY")
	jwtPrivateKeyID := os.Getenv("JWT_PRIVATE_KEY_ID")

	if jwtPrivateKey == "" {
		privateKeyPath := keysDir + "/jwt_private.pem"
		log.Printf("JWT_PRIVATE_KEY not set in env, checking for %s\n", privateKeyPath)
		data, err := os.ReadFile(privateKeyPath)
		if err == nil {
			jwtPrivateKey = string(data)
			if jwtPrivateKeyID == "" {
				jwtPrivateKeyID = "jwt_public" // matched the public file base name natively for standard tests
			}
		}
	}
	if jwtPrivateKeyID == "" {
		jwtPrivateKeyID = "default-kid"
	}

	keysJSON := os.Getenv("JWT_PUBLIC_KEYS")
	publicKeys, err := auth.LoadPublicKeys(keysJSON, keysDir)
	if err != nil {
		log.Printf("Warning: Failed to load public keys: %v\n", err)
	}

	return &Config{
		DatabaseURL:        databaseURL,
		Port:               port,
		JWTPrivateKey:      jwtPrivateKey,
		JWTPrivateKeyID:    jwtPrivateKeyID,
		JWTPublicKeys:      publicKeys,
		AccessTokenExpiry:  accessTokenExpiry,
		RefreshTokenExpiry: refreshTokenExpiry,
	}
}
