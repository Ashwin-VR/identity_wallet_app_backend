package handlers

import (
	"log"
	"net/http"

	"github.com/Ashwin-VR/verifier-service/client"
	"github.com/Ashwin-VR/verifier-service/models"
	"github.com/gin-gonic/gin"
)

// AuthenticateIdentity handles POST /api/v1/authenticate
func AuthenticateIdentity(c *gin.Context) {
	var req models.AuthenticateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("🔍 Authentication request for UUID: %s", req.UUID)

	// Initialize Grid Platform client
	gridClient := client.NewGridPlatformClient()

	// Step 1: Get identity from Grid Platform
	identity, err := gridClient.GetIdentity(req.UUID)
	if err != nil {
		log.Printf("❌ Failed to retrieve identity: %v", err)
		c.JSON(http.StatusNotFound, models.AuthenticateResponse{
			Success:       false,
			Authenticated: false,
			Message:       "Identity not found",
		})
		return
	}

	log.Printf("✅ Identity retrieved: UUID=%s", identity.UUID)

	// Step 2: Verify identity through Grid Platform
	verifyReq := models.VerifyRequest{
		UUID:         identity.UUID,
		AadhaarHash:  identity.AadhaarHash,
		DigitalSig:   identity.DigitalSig,
		PublicKey:    identity.PublicKey,
		CombinedHash: identity.CombinedHash,
		HashFunction: identity.HashFunction,
	}

	verifyResp, err := gridClient.VerifyIdentity(verifyReq)
	if err != nil {
		log.Printf("❌ Verification failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.AuthenticateResponse{
			Success:       false,
			Authenticated: false,
			Message:       "Verification process failed: " + err.Error(),
		})
		return
	}

	// Step 3: Return result
	if verifyResp.Verified {
		log.Printf("✅ Identity verified successfully: UUID=%s", req.UUID)
		c.JSON(http.StatusOK, models.AuthenticateResponse{
			Success:       true,
			Authenticated: true,
			Message:       "Authentication successful",
			SessionID:     req.SessionID,
		})
	} else {
		log.Printf("❌ Identity verification failed: %s", verifyResp.Reason)
		c.JSON(http.StatusOK, models.AuthenticateResponse{
			Success:       true,
			Authenticated: false,
			Message:       verifyResp.Reason,
			SessionID:     req.SessionID,
		})
	}
}
