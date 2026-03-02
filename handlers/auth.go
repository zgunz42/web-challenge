package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthTokenRequest is the body for POST /auth/token.
type AuthTokenRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthTokenResponse is the response for POST /auth/token.
type AuthTokenResponse struct {
	Token string `json:"token"`
}

// Valid credentials for the challenge (Level 5).
const (
	AuthUsername = "admin"
	AuthPassword = "password"
	JWTSecret    = "supersecretkey-change-in-production" // In production, use env var
)

// JWTClaims represents the JWT payload
type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthToken issues a JWT token for valid credentials (Level 5).
func AuthToken(c *gin.Context) {
	var req AuthTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
		return
	}
	if req.Username != AuthUsername || req.Password != AuthPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	
	// Generate JWT token
	claims := JWTClaims{
		Username: req.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   req.Username,
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	
	c.JSON(http.StatusOK, AuthTokenResponse{Token: tokenString})
}
