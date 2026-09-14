package main

import (
	"fmt"
	"time"

	"://github.com"
)

var (
	// In production, keep these in environment variables!
	accessSecret  = []byte("short-term-secret-key")
	refreshSecret = []byte("long-term-secret-key")
)

// AccessTokenClaims holds data needed for every request (User ID, Email)
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims only needs the UserID to re-verify who they are
type RefreshTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateTokens creates BOTH tokens at the same time
func GenerateTokens(userID, email string) (string, string, error) {
	// 1. Generate Access Token (Short-lived: 15 Minutes)
	accessClaims := AccessTokenClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := accessTokenObj.SignedString(accessSecret)
	if err != nil {
		return "", "", err
	}

	// 2. Generate Refresh Token (Long-lived: 7 Days)
	refreshClaims := RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString(refreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateAccessToken parses and checks the short-lived access token
func ValidateAccessToken(tokenStr string) (*AccessTokenClaims, error) {
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
func ValidateRefreshToken(tokenStr string) (*RefreshTokenClaims, error) {
	claims := &RefreshTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return refreshSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}
	return claims, nil
}

