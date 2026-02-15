package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func VerifyUUIDHandler(c *gin.Context) {
	var req struct {
		UUID      string `json:"uuid"`
		Challenge string `json:"challenge"`
		Signature string `json:"signature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	gridURL := os.Getenv("GRID_URL") + "/verify"
	apiKey := os.Getenv("INTERNAL_SERVICE_KEY")

	jsonData, _ := json.Marshal(req)
	clientReq, _ := http.NewRequest("POST", gridURL, bytes.NewBuffer(jsonData))
	clientReq.Header.Set("Content-Type", "application/json")
	clientReq.Header.Set("X-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(clientReq)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Grid API unreachable"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(resp.StatusCode, result)
}
