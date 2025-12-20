package handlers

import (
	"log"
	"net/http"

	"github.com/Ashwin-VR/grid-platform-api/models"
	"github.com/gin-gonic/gin"
)

// Mock storage (in production, this will be Hyperledger Fabric)
var identityStore = make(map[string]models.Identity)

// OnboardIdentity handles POST /api/v1/identity/onboard
func OnboardIdentity(c *gin.Context) {
	var identity models.Identity

	if err := c.ShouldBindJSON(&identity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Call Hyperledger Fabric to store identity
	// For now, store in memory
	identityStore[identity.UUID] = identity

	log.Printf("✅ Identity onboarded: UUID=%s", identity.UUID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Identity onboarded successfully",
		"uuid":    identity.UUID,
	})
}

// GetIdentity handles GET /api/v1/internal/identity/:uuid
func GetIdentity(c *gin.Context) {
	uuid := c.Param("uuid")

	// TODO: Query Hyperledger Fabric ledger
	// For now, query from memory
	identity, exists := identityStore[uuid]

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Identity not found",
		})
		return
	}

	c.JSON(http.StatusOK, identity)
}

// VerifyIdentity handles POST /api/v1/internal/identity/verify
func VerifyIdentity(c *gin.Context) {
	var req models.VerifyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("🔍 Verifying identity: UUID=%s", req.UUID)

	// Get identity from store
	identity, exists := identityStore[req.UUID]
	if !exists {
		log.Printf("❌ Identity not found: UUID=%s", req.UUID)
		c.JSON(http.StatusNotFound, models.VerifyResponse{
			Verified: false,
			Reason:   "Identity not found",
		})
		return
	}

	log.Printf("✅ Identity found in store: UUID=%s", req.UUID)

	// Simple verification: Check if all fields match (NO CRYPTO FOR NOW)
	verified := true
	reason := "Identity verified successfully"

	if identity.AadhaarHash != req.AadhaarHash {
		verified = false
		reason = "Aadhaar hash mismatch"
		log.Printf("❌ Aadhaar hash mismatch")
	} else if identity.DigitalSig != req.DigitalSig {
		verified = false
		reason = "Digital signature mismatch"
		log.Printf("❌ Digital signature mismatch")
	} else if identity.PublicKey != req.PublicKey {
		verified = false
		reason = "Public key mismatch"
		log.Printf("❌ Public key mismatch")
	} else if identity.CombinedHash != req.CombinedHash {
		verified = false
		reason = "Combined hash mismatch"
		log.Printf("❌ Combined hash mismatch")
	}

	if verified {
		log.Printf("✅ Identity fully verified: UUID=%s", req.UUID)
	}

	c.JSON(http.StatusOK, models.VerifyResponse{
		Verified: verified,
		Reason:   reason,
	})
}
