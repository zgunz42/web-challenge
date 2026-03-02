package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs detailed request and response information
func RequestLogger(c *gin.Context) {
	// Capture request body
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
		// Restore the body for the actual handler
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}

	// Log request details
	log.Printf("\n=== INCOMING REQUEST ===")
	log.Printf("Method: %s", c.Request.Method)
	log.Printf("Path: %s", c.Request.URL.Path)
	log.Printf("Query: %s", c.Request.URL.RawQuery)
	log.Printf("Headers:")
	for name, values := range c.Request.Header {
		for _, value := range values {
			log.Printf("  %s: %s", name, value)
		}
	}
	
	// Log Authorization specifically
	auth := c.GetHeader("Authorization")
	if auth != "" {
		log.Printf("Auth Header: '%s'", auth)
		log.Printf("Auth Length: %d", len(auth))
		log.Printf("Auth Bytes: %v", []byte(auth))
	} else {
		log.Printf("Auth Header: MISSING/EMPTY")
	}

	if len(requestBody) > 0 {
		log.Printf("Request Body: %s", string(requestBody))
		// Try to pretty print JSON
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, requestBody, "", "  "); err == nil {
			log.Printf("Request Body (formatted):\n%s", prettyJSON.String())
		}
	} else {
		log.Printf("Request Body: EMPTY")
	}

	// Create a response writer that captures the response
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw

	// Record start time
	start := time.Now()

	// Process request
	c.Next()

	// Calculate duration
	duration := time.Since(start)

	// Log response details
	log.Printf("\n=== OUTGOING RESPONSE ===")
	log.Printf("Status: %d", blw.Status())
	log.Printf("Duration: %v", duration)
	log.Printf("Response Headers:")
	for name, values := range blw.Header() {
		for _, value := range values {
			log.Printf("  %s: %s", name, value)
		}
	}
	
	responseBody := blw.body.String()
	if len(responseBody) > 0 {
		log.Printf("Response Body: %s", responseBody)
		// Try to pretty print JSON
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, []byte(responseBody), "", "  "); err == nil {
			log.Printf("Response Body (formatted):\n%s", prettyJSON.String())
		}
	} else {
		log.Printf("Response Body: EMPTY")
	}
	
	log.Printf("=== END REQUEST ===\n")
}

// bodyLogWriter wraps gin.ResponseWriter to capture response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
