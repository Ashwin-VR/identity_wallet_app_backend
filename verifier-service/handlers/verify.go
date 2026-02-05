package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

func VerifyHandler(c *fiber.Ctx) error {
	uuid := c.Query("uuid")
	if uuid == "" {
		return c.Status(400).JSON(fiber.Map{"error": "UUID is required"})
	}

	// 1. Prepare request to Grid Platform
	gridURL := os.Getenv("GRID_PLATFORM_URL")
	verifyKey := os.Getenv("VERIFIER_API_KEY")

	fullURL := fmt.Sprintf("%s/api/v1/internal/identity/verify?uuid=%s", gridURL, uuid)

	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create request"})
	}

	// 2. Add the Auth Header that Grid API expects
	req.Header.Set("X-Verifier-Key", verifyKey)

	// 3. Execute call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Grid platform unreachable"})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	return c.Status(resp.StatusCode).Send(body)
}
