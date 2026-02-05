package main

import (
	"fmt"
	"log"
	"os"

	// THE FIX: Use the full module path from go.mod
	"github.com/Ashwin-VR/grid-platform-api/fabric"
	"github.com/Ashwin-VR/grid-platform-api/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}

	app := fiber.New()
	app.Use(logger.New())

	// Initialize Fabric Client
	fabricClient, err := fabric.NewFabricClient()
	if err != nil {
		log.Fatalf("Critical: Could not connect to Fabric Network: %v", err)
	}

	h := &handlers.IdentityHandler{Fabric: fabricClient}

	// Middleware for Verifier Service
	authInternal := keyauth.New(keyauth.Config{
		KeyLookup: "header:X-Verifier-Key",
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			return key == os.Getenv("VERIFIER_API_KEY"), nil
		},
	})

	api := app.Group("/api/v1/internal/identity")

	// Endpoint 1: Onboard (Wallet)
	api.Post("/onboard", h.Onboard)

	// Endpoint 2: Verify (Verifier Service Only)
	api.Post("/verify", authInternal, h.Verify)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Grid Platform API starting on port %s...\n", port)
	log.Fatal(app.Listen(":" + port))
}
