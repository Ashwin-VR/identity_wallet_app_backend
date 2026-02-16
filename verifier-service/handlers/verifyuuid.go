package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// --- WEBSOCKET INFRASTRUCTURE ---

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan VerificationEvent)
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

type VerificationEvent struct {
	UUID      string `json:"uuid"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func InitHub() {
	go handleBroadcasts()
}

func handleBroadcasts() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	clients[conn] = true
}

// --- VERIFICATION HANDLER ---

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

	// BROADCAST IF SUCCESSFUL
	if resp.StatusCode == 200 && result["valid"] == true {
		broadcast <- VerificationEvent{
			UUID:      req.UUID,
			Status:    "SUCCESS",
			Timestamp: time.Now().Format(time.RFC3339),
		}
		// Cleanup the P2P session after successful verification
		SessionLock.Lock()
		delete(P2PSessions, req.UUID)
		SessionLock.Unlock()
	}

	c.JSON(resp.StatusCode, result)
}
