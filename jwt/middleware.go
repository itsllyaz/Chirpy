package main

import (
	"context"
	"net/http"
	"strings"
)

// AuthMiddleware intercepts requests to make sure the Access Token is valid
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Grab the "Authorization: Bearer <token>" header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		// 2. Split the word "Bearer" away from the token string
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization Header Format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// 3. Use our core logic to validate the token
		claims, err := ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// 4. Put the extracted UserID into the request Context so endpoints can see who it is
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

