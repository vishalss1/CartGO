package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	pkgAuth "github.com/vishalss1/CartGO/pkg/auth"
)

func GenerateAccessToken(userID, role, privateKeyPEM, kid, expiryStr string) (string, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 15 * time.Minute
	}

	claims := pkgAuth.TokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    "cartgo-user-service",
			Audience:  jwt.ClaimStrings{"cartgo-api"},
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid

	return token.SignedString(key)
}

func GenerateRefreshToken(userID, privateKeyPEM, kid, expiryStr string) (string, time.Time, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse private key: %v", err)
	}

	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 7 * 24 * time.Hour
	}

	expiresAt := time.Now().Add(expiry)
	claims := pkgAuth.TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    "cartgo-user-service",
			Audience:  jwt.ClaimStrings{"cartgo-api"},
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	
	tokenStr, err := token.SignedString(key)
	return tokenStr, expiresAt, err
}

func ValidateToken(tokenStr string, publicKeys map[string]string) (*pkgAuth.TokenClaims, error) {
	return pkgAuth.ValidateToken(tokenStr, publicKeys, "cartgo-api")
}
