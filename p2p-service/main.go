package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	GridAPIURL     = "http://localhost:8080"
	InternalAPIKey = "your_internal_service_key"
)

type P2PSession struct {
	ID         string `json:"session_id"`
	ProverUUID string `json:"prover_uuid"`
	Challenge  string `json:"challenge"`
	Status     string `json:"status"`
}

var (
	sessions = make(map[string]*P2PSession)
	mu       sync.RWMutex
)

func main() {
	r := gin.Default()

	r.POST("/p2p/create", func(c *gin.Context) {
		var req struct {
			ProverUUID string `json:"prover_uuid"`
		}
		c.ShouldBindJSON(&req)
		id := uuid.New().String()
		s := &P2PSession{ID: id, ProverUUID: req.ProverUUID, Status: "awaiting_challenge"}
		mu.Lock()
		sessions[id] = s
		mu.Unlock()
		c.JSON(200, s)
	})

	r.POST("/p2p/post-challenge", func(c *gin.Context) {
		var req struct{ SessionID, Challenge string }
		c.ShouldBindJSON(&req)
		mu.Lock()
		if s, ok := sessions[req.SessionID]; ok {
			s.Challenge = req.Challenge
			s.Status = "awaiting_consent"
		}
		mu.Unlock()
		c.Status(200)
	})

	r.POST("/p2p/submit-signature", func(c *gin.Context) {
		var req struct{ SessionID, Signature string }
		c.ShouldBindJSON(&req)

		mu.RLock()
		s, ok := sessions[req.SessionID]
		mu.RUnlock()

		if !ok || req.Signature == "REJECTED" {
			c.JSON(200, gin.H{"valid": false})
			return
		}

		// Verify against Blockchain via Grid API
		payload := map[string]string{
			"uuid": s.ProverUUID, "challenge": s.Challenge, "signature": req.Signature,
		}
		jsonB, _ := json.Marshal(payload)
		hReq, _ := http.NewRequest("POST", GridAPIURL+"/verify", bytes.NewBuffer(jsonB))
		hReq.Header.Set("X-API-KEY", InternalAPIKey)
		hReq.Header.Set("Content-Type", "application/json")

		resp, err := (&http.Client{}).Do(hReq)
		var result struct {
			Valid bool `json:"valid"`
		}
		if err == nil {
			json.NewDecoder(resp.Body).Decode(&result)
			mu.Lock()
			if result.Valid {
				s.Status = "success"
			} else {
				s.Status = "failed"
			}
			mu.Unlock()
		}
		c.JSON(200, result)
	})

	r.GET("/p2p/status/:id", func(c *gin.Context) {
		mu.RLock()
		s, ok := sessions[c.Param("id")]
		mu.RUnlock()
		if !ok {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		c.JSON(200, s)
	})

	r.Run("0.0.0.0:8083")
}
