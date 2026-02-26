package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	GridAPIURL     = "http://localhost:8080"
	InternalAPIKey = "your_internal_service_key"
)

var (
	challenges = make(map[string]string)
	mu         sync.RWMutex
)

func main() {
	r := gin.Default()

	r.GET("/challenge/:uuid", func(c *gin.Context) {
		b := make([]byte, 32)
		rand.Read(b)
		val := hex.EncodeToString(b)
		mu.Lock()
		challenges[c.Param("uuid")] = val
		mu.Unlock()
		c.JSON(200, gin.H{"uuid": c.Param("uuid"), "challenge": val})
	})

	r.POST("/verify-uuid", func(c *gin.Context) {
		var req struct {
			UUID      string `json:"uuid"`
			Challenge string `json:"challenge"`
			Signature string `json:"signature"`
		}
		c.ShouldBindJSON(&req)

		// 1. Internal check: Did we actually issue this challenge?
		mu.RLock()
		stored, ok := challenges[req.UUID]
		mu.RUnlock()
		if !ok || stored != req.Challenge {
			c.JSON(200, gin.H{"valid": false, "error": "Invalid challenge"})
			return
		}

		// 2. Call Grid API (Blockchain)
		jsonB, _ := json.Marshal(req)
		hReq, _ := http.NewRequest("POST", GridAPIURL+"/verify", bytes.NewBuffer(jsonB))
		hReq.Header.Set("X-API-KEY", InternalAPIKey)
		hReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(hReq)
		if err != nil {
			c.JSON(500, gin.H{"valid": false, "error": "Blockchain API unreachable"})
			return
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		c.JSON(200, result)
	})

	r.Run("0.0.0.0:8081")
}
