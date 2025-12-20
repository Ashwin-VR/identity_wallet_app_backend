package main

import (
	"log"
	"os"

	"github.com/Ashwin-VR/grid-platform-api/handlers"
	"github.com/Ashwin-VR/grid-platform-api/middleware"
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
		port = "8080"
	}

	// Validate API key is set
	if os.Getenv("VERIFIER_API_KEY") == "" {
		log.Fatal("❌ VERIFIER_API_KEY not set in environment variables")
	}

	// Initialize Gin router
	router := gin.Default()

	// CORS configuration for Flutter app
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "Grid Platform API",
		})
	})

	// Public API routes (for Flutter App) - NO AUTH REQUIRED
	v1 := router.Group("/api/v1")
	{
		// Aadhaar OTP endpoints
		aadhaar := v1.Group("/aadhaar")
		{
			aadhaar.POST("/otp/generate", handlers.GenerateOTP)
			aadhaar.POST("/otp/verify", handlers.VerifyOTP)
		}

		// Identity onboarding endpoint (public)
		v1.POST("/identity/onboard", handlers.OnboardIdentity)
	}

	// Internal API routes (for Verifier Service) - API KEY REQUIRED 🔐
	internal := router.Group("/api/v1/internal")
	internal.Use(middleware.APIKeyAuth())
	{
		internal.GET("/identity/:uuid", handlers.GetIdentity)
		internal.POST("/identity/verify", handlers.VerifyIdentity)
	}

	// Start server
	log.Printf("🚀 Grid Platform API running on port %s", port)
	log.Printf("🔐 API Key authentication enabled for internal endpoints")
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
