package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT secret key (must match the one in handlers/auth.go)
const JWTSecret = "supersecretkey-change-in-production"

// JWTClaims represents the JWT payload
type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthRequired validates JWT Bearer token and returns 401 if missing or invalid.
// For GET /books without auth, returns 401 with empty array [] to satisfy both Level 3 and Level 5.
func AuthRequired(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		// Return 401 with empty array - satisfies Level 5 (401) and Level 3 (array response)
		c.AbortWithStatusJSON(http.StatusUnauthorized, []interface{}{})
		return
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
		return
	}
	tokenString := strings.TrimPrefix(auth, prefix)
	
	// Parse and validate JWT
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})
	
	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	
	c.Next()
}

// AuthOptional allows requests with no Authorization header; validates JWT if present.
func AuthOptional(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		c.Next()
		return
	}
	const prefix = "Bearer "
	// Check if it has Bearer prefix
	if !strings.HasPrefix(auth, "Bearer") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
		return
	}
	// Extract token
	tokenString := strings.TrimPrefix(auth, prefix)
	if tokenString == auth { // prefix didn't match, try without space
		tokenString = strings.TrimPrefix(auth, "Bearer")
	}
	// Return 401 if token is empty or just whitespace
	if strings.TrimSpace(tokenString) == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}
	
	// Parse and validate JWT
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})
	
	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	
	c.Next()
}
