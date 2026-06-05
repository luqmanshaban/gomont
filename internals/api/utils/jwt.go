package utils

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	Email string `json:"email"`
	ID    int    `json:"id"`
	jwt.RegisteredClaims
}

func GenerateJWT(jwtKey, email string, id int) (string, error) {
	claims := CustomClaims{
		Email: email,
		ID:    id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "my-go-auth-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		slog.Error("failed to generate jwt", "email", email, "error", err)
		return "", err
	}

	return tokenStr, nil
}

// ValidateToken parses and validates the given JWT string
func ValidateToken(jwtKey, tokenString string) (*CustomClaims, error) {
	// Parse the token with the custom claims structure
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method is what we expect (HMAC / HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			slog.Error("error parsing token", "error", fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtKey), nil
	})

	if err != nil {
		slog.Error("failed to validate token", "error", err)
		return nil, err
	}

	// Extract and validate claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	slog.Error("invalid token", "error", fmt.Errorf("invalid token"))
	return nil, fmt.Errorf("invalid token")
}
