package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const VerifierURL = "http://localhost:8081"

type SessionInfo struct {
	Status string `json:"status"`
	UUID   string `json:"uuid,omitempty"`
}

var (
	sessions    = make(map[string]*SessionInfo)
	sessionLock sync.RWMutex
)

func main() {
	r := gin.Default()

	// Web Interface: Initialize a session and get a QR link
	r.GET("/bank/init-login", func(c *gin.Context) {
		sessionID := fmt.Sprintf("sess-%d", time.Now().UnixNano())
		sessionLock.Lock()
		sessions[sessionID] = &SessionInfo{Status: "pending"}
		sessionLock.Unlock()

		c.JSON(200, gin.H{
			"session_id":    sessionID,
			"bank_endpoint": "http://localhost:8082/api", // Path for Flutter app
		})
	})

	// Web Interface: Poll to see if Flutter user finished signing
	r.GET("/bank/status/:session_id", func(c *gin.Context) {
		sessionLock.RLock()
		info, exists := sessions[c.Param("session_id")]
		sessionLock.RUnlock()
		if !exists {
			c.JSON(404, gin.H{"error": "Session not found"})
			return
		}
		c.JSON(200, info)
	})

	// Flutter App: Request a challenge via the Bank
	r.GET("/api/get-challenge/:uuid", func(c *gin.Context) {
		resp, err := http.Get(VerifierURL + "/challenge/" + c.Param("uuid"))
		if err != nil {
			c.JSON(500, gin.H{"error": "Verifier down"})
			return
		}
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		c.JSON(resp.StatusCode, res)
	})

	// Flutter App: Submit the signed proof
	r.POST("/api/submit-proof", func(c *gin.Context) {
		var body struct {
			UUID      string `json:"uuid"`
			Challenge string `json:"challenge"`
			Signature string `json:"signature"`
			SessionID string `json:"session_id"`
		}
		c.ShouldBindJSON(&body)

		jsonB, _ := json.Marshal(body)
		resp, err := http.Post(VerifierURL+"/verify-uuid", "application/json", bytes.NewBuffer(jsonB))
		if err != nil {
			c.JSON(500, gin.H{"error": "Verifier down"})
			return
		}
		defer resp.Body.Close()

		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)

		if res["valid"] == true && body.SessionID != "" {
			sessionLock.Lock()
			sessions[body.SessionID] = &SessionInfo{Status: "success", UUID: body.UUID}
			sessionLock.Unlock()
		}

		c.JSON(resp.StatusCode, res)
	})

	r.Run("0.0.0.0:8082")
}
