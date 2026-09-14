package internal

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/itsllyaz/Chirpy/internal/auth"
)
type MiddlewareConfig struct {
	JWTSecretKey []byte
	// You can add your db queries here later if you need to check DB records in middleware:
	// DBQueries *database.Queries 
}

func New(secret []byte) *MiddlewareConfig {
	return &MiddlewareConfig{JWTSecretKey: secret}
}
func (m *MiddlewareConfig) LoginAuthMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
	// 1. Grab the "Authorization: Bearer <token>" header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}
		
		token := strings.TrimPrefix(authHeader, "Bearer")
		// 2. Split the word "Bearer" away from the token string
		// parts := strings.Split(authHeader, " ")
		// if len(parts) != 2 || parts[0] != "Bearer" {
		// 	http.Error(w, "Invalid Authorization Header Format", http.StatusUnauthorized)
		// 	return
		// }
		// tokenString := parts[1]

		// 3. Use our core logic to validate the token
		claims, err := auth.ValidateAccessToken(token, m.JWTSecretKey)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			fmt.Println("THE ERROR: ", err)
			return
		}

		// 4. Put the extracted UserID into the request Context so endpoints can see who it is
		ctx := context.WithValue(r.Context(), "user_id", claims.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
