package main

import (
	"fmt"
	"log"
	"os"

	// FIXED IMPORT: Use the full module path from go.mod
	"github.com/Ashwin-VR/verifier-service/handlers"

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

	// Middleware: Verify Banking Site requests
	authVerifier := keyauth.New(keyauth.Config{
		KeyLookup: "header:X-Verifier-Key",
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			return key == os.Getenv("VERIFIER_API_KEY"), nil
		},
	})

	// Banking site calls this to check a user
	app.Post("/api/v1/verify-user", authVerifier, handlers.VerifyHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Printf("Verifier Service starting on port %s...\n", port)
	log.Fatal(app.Listen(":" + port))
}
