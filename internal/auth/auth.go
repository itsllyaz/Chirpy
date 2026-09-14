package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)
var JwtKey = []byte("super_secret")

type AccessTokenClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email  string   `json:"email"`
	jwt.RegisteredClaims
		
}

type RefreshTokenClaims struct{
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}
func HashPassword(password string) (string, error) {
	HashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil{
		return "", err
	}

	return HashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	Match, err := argon2id.ComparePasswordAndHash(password, hash)	
	if err != nil{
		return false, err
	}

	return Match, nil
}


func GenerateTokens(userID uuid.UUID, Email string, jwtSecretKey []byte) (string, string, error){
	accessClaim := AccessTokenClaims{
		UserID:  userID,
		Email: Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Second)), // Short lifespan
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "chirpy-access",
		},
	}

	// 2. Create the token using HS256 algorithm and claims
	acessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaim)

	// 3. Sign the token with your secret key
	acessToken, err := acessTokenObj.SignedString(jwtSecretKey)
	if err != nil {
		return "", "", err
	}
	
	refreshClaim := RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaim)
	refreshToken, err := refreshTokenObj.SignedString(jwtSecretKey) 
	return acessToken , refreshToken, nil
}


func ValidateAccessToken(tokenStr string, accessSecret []byte) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return accessSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired access token")
	} 
	return claims, nil
}

// ValidateRefreshToken parses and checks the long-lived refresh token
func ValidateRefreshToken(tokenStr string, refreshSecret []byte) (*RefreshTokenClaims, error) {
	claims := &RefreshTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return refreshSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}
	return claims, nil
}





func ValidateJWT(tokenString string, jwtSecretKey []byte) (uuid.UUID, error){
	extractedClaims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, extractedClaims, func(token *jwt.Token)(any,error){
		return jwtSecretKey, nil
	})
	
	if err != nil{
		return uuid.Nil, err
	}

	if token.Valid{
		return extractedClaims.UserID, nil
	}
	return uuid.Nil, errors.New("Invalid token")
}


// func GetBearerToken(headers http.Header) (string, error){
// 	headers	 
// }
