package handlers

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

func GetChallengeHandler(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(400, gin.H{"error": "UUID is required"})
		return
	}

	// Generate a secure random 32-byte challenge
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate challenge"})
		return
	}
	challenge := hex.EncodeToString(b)

	// In production, save this challenge in Redis mapping to the UUID
	// for verification in the next step.
	c.JSON(200, gin.H{
		"uuid":      uuid,
		"challenge": challenge,
		"message":   "Please sign this challenge using your private key (Base64 r||s format)",
	})
}
