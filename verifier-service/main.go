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

	r := gin.Default()

	// External endpoints
	r.GET("/challenge/:uuid", handlers.GetChallengeHandler)
	r.POST("/verify-uuid", handlers.VerifyUUIDHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	fmt.Printf("Verifier Service running on :%s\n", port)
	r.Run(":" + port)
}
