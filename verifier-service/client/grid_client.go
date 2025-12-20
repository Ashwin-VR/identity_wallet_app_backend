package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Ashwin-VR/verifier-service/models"
)

// GridPlatformClient handles communication with Grid Platform API
type GridPlatformClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// NewGridPlatformClient creates a new Grid Platform client
func NewGridPlatformClient() *GridPlatformClient {
	return &GridPlatformClient{
		BaseURL: os.Getenv("GRID_PLATFORM_URL"),
		APIKey:  os.Getenv("VERIFIER_API_KEY"),
		Client:  &http.Client{},
	}
}

// GetIdentity retrieves identity from Grid Platform
func (g *GridPlatformClient) GetIdentity(uuid string) (*models.Identity, error) {
	url := fmt.Sprintf("%s/api/v1/internal/identity/%s", g.BaseURL, uuid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add API Key header 🔑
	req.Header.Set("X-API-Key", g.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var identity models.Identity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &identity, nil
}

// VerifyIdentity verifies identity through Grid Platform
func (g *GridPlatformClient) VerifyIdentity(req models.VerifyRequest) (*models.VerifyResponse, error) {
	url := fmt.Sprintf("%s/api/v1/internal/identity/verify", g.BaseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add API Key header 🔑
	httpReq.Header.Set("X-API-Key", g.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var verifyResp models.VerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &verifyResp, nil
}
