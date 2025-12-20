package main

import (
	"log"
	"os"

	"github.com/Ashwin-VR/verifier-service/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Validate required environment variables
	if os.Getenv("VERIFIER_API_KEY") == "" {
		log.Fatal("❌ VERIFIER_API_KEY not set in environment variables")
	}
	if os.Getenv("GRID_PLATFORM_URL") == "" {
		log.Fatal("❌ GRID_PLATFORM_URL not set in environment variables")
	}

	// Initialize Gin router
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "Verifier Service",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/authenticate", handlers.AuthenticateIdentity) // ✅ Single endpoint
	}

	// Start server
	log.Printf("🚀 Verifier Service running on port %s", port)
	log.Printf("🔗 Connected to Grid Platform: %s", os.Getenv("GRID_PLATFORM_URL"))
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
