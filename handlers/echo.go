package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Echo returns the request body byte-for-byte to preserve key order (Level 2).
func Echo(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		c.Data(http.StatusOK, "application/json", []byte("{}"))
		return
	}
	c.Data(http.StatusOK, "application/json", body)
}
