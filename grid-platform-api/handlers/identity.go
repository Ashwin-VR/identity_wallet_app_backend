package handlers

import (
	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type Identity struct {
	UUID          string `json:"uuid"`
	AadhaarHash   string `json:"aadhaarHash"`
	DigitalSig    string `json:"digitalSig"`
	PublicKey     string `json:"publicKey"`
	HashFunction  string `json:"hashFunction"`
	TwoFARequired bool   `json:"twoFARequired"`
	CombinedHash  string `json:"combinedHash"`
	Timestamp     string `json:"timestamp"`
}

func CreateIdentityHandler(contract *client.Contract) gin.HandlerFunc {
	return func(c *gin.Context) {
		var id Identity
		if err := c.ShouldBindJSON(&id); err != nil {
			c.JSON(400, gin.H{"success": false, "error": "Invalid Identity format"})
			return
		}

		// Ground Truth: Chaincode expects the raw JSON string as the first argument
		idJSON, _ := json.Marshal(id)
		_, err := contract.SubmitTransaction("CreateIdentity", string(idJSON))
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": "Identity onboarded successfully"})
	}
}

func VerifyIdentityHandler(contract *client.Contract) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Internal Security Check
		if c.GetHeader("X-API-KEY") != os.Getenv("INTERNAL_SERVICE_KEY") {
			c.JSON(401, gin.H{"success": false, "error": "Unauthorized"})
			return
		}

		var req struct {
			UUID      string `json:"uuid"`
			Challenge string `json:"challenge"`
			Signature string `json:"signature"` // Base64 r||s
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": "Invalid request"})
			return
		}

		// Ground Truth: Chaincode expects 3 string arguments
		result, err := contract.EvaluateTransaction("VerifyIdentity", req.UUID, req.Challenge, req.Signature)
		if err != nil {
			c.JSON(200, gin.H{"success": true, "valid": false, "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"success": true, "valid": string(result) == "true"})
	}
}
