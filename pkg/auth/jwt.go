package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID string `json:"sub,omitempty"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// ValidateToken parses and extensively validates a JWT based on CartGO security requirements
func ValidateToken(tokenStr string, publicKeys map[string]string, expectedAudience string) (*TokenClaims, error) {
	// 1. Initial Parse and Verification of Signing Method and Key
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Enforce RS256 algorithm ONLY
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Extract Key ID (kid) from Header
		kidRaw, ok := token.Header["kid"]
		if !ok {
			return nil, errors.New("missing 'kid' header in token")
		}
		kid, ok := kidRaw.(string)
		if !ok {
			return nil, errors.New("'kid' header must be a string")
		}

		// Look up public key for this kid
		pubKeyPEM, exists := publicKeys[kid]
		if !exists {
			return nil, fmt.Errorf("unknown key ID (kid): %s", kid)
		}

		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pubKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key for kid %s: %v", kid, err)
		}

		return key, nil
	})

	if err != nil {
		return nil, err
	}

	// 2. Extract Claims
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token or malformed claims")
	}

	// 3. Strict Payload Verification
	
	// Issuer MUST be "cartgo-user-service"
	issuer, err := claims.GetIssuer()
	if err != nil || issuer != "cartgo-user-service" {
		return nil, fmt.Errorf("invalid issuer: %v", issuer)
	}

	// Audience MUST contain target service or "cartgo-api"
	audiences, err := claims.GetAudience()
	if err != nil {
		return nil, errors.New("missing or invalid audience")
	}
	
	validAudience := false
	for _, aud := range audiences {
		if aud == expectedAudience || aud == "cartgo-api" {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return nil, fmt.Errorf("invalid audience, expected %s or cartgo-api", expectedAudience)
	}

	// Subject must exist
	subject, err := claims.GetSubject()
	if err != nil || subject == "" {
		// Backwards compatibility handling (in case sub isn't properly registered as string)
		if claims.UserID == "" {
			return nil, errors.New("missing subject (user_id)")
		}
	} else {
	    // Standardize
	    claims.UserID = subject
	}
	
	// Role must exist for meaningful internal operations
	if claims.Role == "" {
		return nil, errors.New("missing role claim")
	}

	// Verify Issued At (iat)
	iat, err := claims.GetIssuedAt()
	if err != nil || iat == nil {
		return nil, errors.New("missing issued at (iat) claim")
	}
	// Small skew allowance for clock differences (e.g. 5 mins)
	if iat.Time.After(time.Now().Add(5 * time.Minute)) {
		return nil, errors.New("token issued in the future (iat drift)")
	}

	return claims, nil
}
