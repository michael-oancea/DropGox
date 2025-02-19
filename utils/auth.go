package auth

import (
	"context"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var jwtPublicKey = loadPublicKey()

// loadPublicKey retrieves the RSA public key from the JWT_PUBLIC_KEY environment variable.
// The key must be in PEM format.
func loadPublicKey() *rsa.PublicKey {
	pubKeyPEM := os.Getenv("JWT_PUBLIC_KEY")
	if pubKeyPEM == "" {
		panic("JWT_PUBLIC_KEY environment variable not set")
	}

	// Decode PEM block
	block, _ := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		panic("failed to parse PEM block containing the public key")
	}

	pub, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pubKeyPEM))
	if err != nil {
		panic(fmt.Sprintf("failed to parse RSA public key: %v", err))
	}
	return pub
}

// KeycloakMiddleware validates JWT tokens using RS256.
func KeycloakMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the Authorization header.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// Parse and validate the token using RS256.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure that the token method is RS256.
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtPublicKey, nil
		})
		if err != nil || !token.Valid {
			log.Printf("Token parsing error: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract claims.
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Check the audience claim: it might be a string or an array.
		audClaim := claims["aud"]
		audValid := false

		switch aud := audClaim.(type) {
		case string:
			// If the audience is a string, check if it matches.
			if aud == "dropgox-backend" || aud == "account" {
				audValid = true
			}
		case []interface{}:
			// If the audience is an array, ensure "dropgox-backend" is included.
			for _, a := range aud {
				if aStr, ok := a.(string); ok && aStr == "dropgox-backend" {
					audValid = true
					break
				}
			}
		default:
			http.Error(w, "Invalid audience format", http.StatusUnauthorized)
			return
		}

		if !audValid {
			log.Printf("Audience claim does not include dropgox-backend: %v", audClaim)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Optional: if the audience array contains "account", check that azp is "dropgox-backend".
		if audArray, ok := audClaim.([]interface{}); ok {
			hasAccount := false
			for _, a := range audArray {
				if aStr, ok := a.(string); ok && aStr == "account" {
					hasAccount = true
					break
				}
			}
			if hasAccount {
				if azp, ok := claims["azp"].(string); !ok || azp != "dropgox-backend" {
					log.Printf("Token azp (%v) does not match expected client", claims["azp"])
					http.Error(w, "Invalid token", http.StatusUnauthorized)
					return
				}
			}
		}

		// Add claims to context.
		ctx := context.WithValue(r.Context(), "claims", claims)
		r = r.WithContext(ctx)

		// Continue with the next handler.
		next.ServeHTTP(w, r)
	})
}
