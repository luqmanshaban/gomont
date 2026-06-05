package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/luqmanshaban/gomont/internals/api/utils"
)

type contextKey string
const UserClaimsContextKey contextKey = "userClaims"

// AuthMiddleware now accepts the jwtKey from your server configuration
func AuthMiddleware(jwtKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.WriteJson(w, http.StatusUnauthorized, map[string]string{"message": "missing authorization header"})
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				utils.WriteJson(w, http.StatusUnauthorized, map[string]string{"message": "invalid authorization header format"})
				return
			}

			tokenString := parts[1]

			// FIX: Pass the jwtKey into your validation utility
			claims, err := utils.ValidateToken(jwtKey, tokenString)
			if err != nil {
				utils.WriteJson(w, http.StatusUnauthorized, map[string]string{"message": "invalid or expired token"})
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}