package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/gin-gonic/gin"
)

// In-memory store for P2P Matchmaking
// UUID -> Active Challenge
var (
	P2PSessions = make(map[string]string)
	SessionLock sync.RWMutex
)

func GetChallengeHandler(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(400, gin.H{"error": "UUID is required"})
		return
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate challenge"})
		return
	}
	challenge := hex.EncodeToString(b)

	// Store for P2P retrieval
	SessionLock.Lock()
	P2PSessions[uuid] = challenge
	SessionLock.Unlock()

	c.JSON(200, gin.H{
		"uuid":      uuid,
		"challenge": challenge,
		"message":   "Challenge generated and stored for verification",
	})
}

// P2PPollHandler allows the person showing the QR to check if someone requested a challenge
func P2PPollHandler(c *gin.Context) {
	uuid := c.Param("uuid")

	SessionLock.RLock()
	challenge, exists := P2PSessions[uuid]
	SessionLock.RUnlock()

	if !exists {
		c.JSON(200, gin.H{"pending": false})
		return
	}

	c.JSON(200, gin.H{
		"pending":   true,
		"challenge": challenge,
	})
}
