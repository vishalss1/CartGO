package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Helper to generate a new key pair for tests
func generateTestKeyPair() (privatePEM, publicPEM string) {
	reader := rand.Reader
	bitSize := 2048

	// Generate key
	key, _ := rsa.GenerateKey(reader, bitSize)

	// Encode private key
	privateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	privatePEM = string(pem.EncodeToMemory(privateBlock))

	// Encode public key
	pubASN1, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}
	publicPEM = string(pem.EncodeToMemory(pubBlock))

	return privatePEM, publicPEM
}

func createTestToken(privatePEM string, kid string, modifier func(*jwt.MapClaims)) string {
	key, _ := jwt.ParseRSAPrivateKeyFromPEM([]byte(privatePEM))

	claims := jwt.MapClaims{
		"sub":  "user-123",
		"role": "CUSTOMER",
		"iss":  "cartgo-user-service",
		"aud":  "cartgo-test-service",
		"nbf":  time.Now().Add(-1 * time.Minute).Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
	}

	if modifier != nil {
		modifier(&claims)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if kid != "" {
		token.Header["kid"] = kid
	}
	tokenStr, _ := token.SignedString(key)
	return tokenStr
}

func TestValidateToken(t *testing.T) {
	privPEM1, pubPEM1 := generateTestKeyPair()
	privPEM2, _ := generateTestKeyPair() // Different key

	publicKeys := map[string]string{
		"key-1": pubPEM1,
	}

	tests := []struct {
		name        string
		token       string
		expectedErr string
	}{
		{
			name:        "Valid Token",
			token:       createTestToken(privPEM1, "key-1", nil),
			expectedErr: "",
		},
		{
			name:        "Expired Token",
			token:       createTestToken(privPEM1, "key-1", func(c *jwt.MapClaims) { (*c)["exp"] = time.Now().Add(-1 * time.Minute).Unix() }),
			expectedErr: "token is expired",
		},
		{
			name:        "Wrong Issuer",
			token:       createTestToken(privPEM1, "key-1", func(c *jwt.MapClaims) { (*c)["iss"] = "attacker-service" }),
			expectedErr: "invalid issuer: attacker-service",
		},
		{
			name:        "Wrong Audience",
			token:       createTestToken(privPEM1, "key-1", func(c *jwt.MapClaims) { (*c)["aud"] = "other-service" }),
			expectedErr: "invalid audience",
		},
		{
			name:        "Invalid Signature (Used Wrong Private Key)",
			token:       createTestToken(privPEM2, "key-1", nil),
			expectedErr: "crypto/rsa: verification error",
		},
		{
			name:        "Unknown Key ID (kid)",
			token:       createTestToken(privPEM1, "unknown-key", nil),
			expectedErr: "unknown key ID (kid): unknown-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateToken(tt.token, publicKeys, "cartgo-test-service")
			if err != nil {
				if tt.expectedErr == "" {
					t.Errorf("Expected no error, got %v", err)
				} else if err.Error() != tt.expectedErr && !containsError(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedErr, err)
				}
			} else {
				if tt.expectedErr != "" {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedErr)
				}
			}
		})
	}
}

func containsError(actual, expected string) bool {
	return strings.Contains(actual, expected)
}
