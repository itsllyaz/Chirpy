package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// LoginRequest is the incoming payload body from the client
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshRequest is used when asking for a new access token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func main() {
	// Public routes
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/refresh", refreshHandler)

	// Protected route (wrapped inside our AuthMiddleware guard)
	http.HandleFunc("/dashboard", AuthMiddleware(dashboardHandler))

	fmt.Println("Server running on port :8081...")
	http.ListenAndServe(":8081", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	json.NewDecoder(r.Body).Decode(&req)

	// Simulate password checking (e.g., if password matches database)
	if req.Username == "alex" && req.Password == "password124" {
		// Create BOTH tokens (User ID = "usr_100", Email = "alex@example.com")
		accessToken, refreshToken, _ := GenerateTokens("usr_100", "alex@example.com")

		// Return both tokens to the frontend
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
		return
	}

	http.Error(w, "Bad Credentials", http.StatusUnauthorized)
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	json.NewDecoder(r.Body).Decode(&req)

	// 1. Verify the incoming long-lived refresh token
	claims, err := ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token. Log in again.", http.StatusForbidden)
		return
	}

	// 2. It's valid! Since we have the UserID from the claims, look up their email
	// (Simulating database lookup here)
	userEmail := "alex@example.com" 

	// 3. Issue a fresh new short-lived access token (and a brand new refresh token)
	newAccessToken, newRefreshToken, _ := GenerateTokens(claims.UserID, userEmail)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Read the user ID that the middleware safely extracted for us
	userID := r.Context().Value("user_id").(string)

	w.Write([]byte(fmt.Sprintf("Welcome to your private dashboard! Your UserID is: %s", userID)))
}

