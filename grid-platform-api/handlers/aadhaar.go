package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/Ashwin-VR/grid-platform-api/models"
	"github.com/gin-gonic/gin"
)

// Mock OTP storage (in production, use Redis with TTL)
var otpStore = make(map[string]string)

// GenerateOTP handles POST /api/v1/aadhaar/otp/generate
func GenerateOTP(c *gin.Context) {
	var req models.OTPGenerateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate mock OTP (6 digits)
	rand.Seed(time.Now().UnixNano())
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Generate transaction ID
	txnID := fmt.Sprintf("TXN%d", time.Now().Unix())

	// Store OTP (in production, store in Redis with 5 min TTL)
	otpStore[req.AadhaarNumber] = otp

	// Mock response (in production, AUA API would send SMS)
	c.JSON(http.StatusOK, models.OTPGenerateResponse{
		Success:       true,
		Message:       fmt.Sprintf("OTP sent successfully (Mock: %s)", otp),
		AadhaarNumber: req.AadhaarNumber,
		TransactionID: txnID,
		MockOTP:       otp, // Only for demo - remove in production
	})
}

// VerifyOTP handles POST /api/v1/aadhaar/otp/verify
func VerifyOTP(c *gin.Context) {
	var req models.OTPVerifyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if OTP exists and matches
	storedOTP, exists := otpStore[req.AadhaarNumber]

	if !exists {
		c.JSON(http.StatusBadRequest, models.OTPVerifyResponse{
			Success:  false,
			Message:  "No OTP found for this Aadhaar number",
			Verified: false,
		})
		return
	}

	if storedOTP != req.OTP {
		c.JSON(http.StatusOK, models.OTPVerifyResponse{
			Success:  true,
			Message:  "Invalid OTP",
			Verified: false,
		})
		return
	}

	// OTP verified successfully
	// Clean up OTP
	delete(otpStore, req.AadhaarNumber)

	c.JSON(http.StatusOK, models.OTPVerifyResponse{
		Success:  true,
		Message:  "OTP verified successfully",
		Verified: true,
	})
}
