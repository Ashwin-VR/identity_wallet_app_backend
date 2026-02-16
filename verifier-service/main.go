package main

import (
	"fmt"
	"log"
	"os"
	"verifier-service/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the Global Hub for WebSockets
	handlers.InitHub()

	r := gin.Default()

	// Standard Endpoints
	r.GET("/challenge/:uuid", handlers.GetChallengeHandler)
	r.POST("/verify-uuid", handlers.VerifyUUIDHandler)

	// --- P2P & NOTIFICATION ENDPOINTS ---

	// WebSocket endpoint for both Enterprise and P2P Listeners
	r.GET("/ws", handlers.WebSocketHandler)

	// P2P Polling: Prover checks if a Scanner has requested a challenge from them
	r.GET("/p2p/poll/:uuid", handlers.P2PPollHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	fmt.Printf("Verifier Service (Enhanced) running on :%s\n", port)
	r.Run(":" + port)
}
