package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping returns a pong message (Level 1).
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}
