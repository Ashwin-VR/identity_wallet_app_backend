package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth validates API key for internal endpoints
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from environment
		expectedKey := os.Getenv("VERIFIER_API_KEY")

		if expectedKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Server configuration error: API key not set",
			})
			c.Abort()
			return
		}

		// Get API key from request header
		providedKey := c.GetHeader("X-API-Key")

		// Validate API key
		if providedKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing API key",
				"hint":  "Include X-API-Key header",
			})
			c.Abort()
			return
		}

		if providedKey != expectedKey {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		// API key is valid, proceed
		c.Next()
	}
}
