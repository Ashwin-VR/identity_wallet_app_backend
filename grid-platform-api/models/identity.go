package models

// Identity represents the digital identity JSON structure
type Identity struct {
	UUID          string `json:"uuid" binding:"required"`
	AadhaarHash   string `json:"aadharhash" binding:"required"`
	DigitalSig    string `json:"Digsig" binding:"required"`
	CombinedHash  string `json:"combinedhash" binding:"required"`
	PublicKey     string `json:"public_key" binding:"required"`
	HashFunction  string `json:"hashfunction" binding:"required"`
	TwoFARequired bool   `json:"2FA_Required"`
}

// OTPGenerateRequest represents OTP generation request
type OTPGenerateRequest struct {
	AadhaarNumber string `json:"aadhaar_number" binding:"required,len=12"`
}

// OTPGenerateResponse represents OTP generation response
type OTPGenerateResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	AadhaarNumber string `json:"aadhaar_number"`
	TransactionID string `json:"transaction_id"`
	MockOTP       string `json:"mock_otp,omitempty"` // Only for demo
}

// OTPVerifyRequest represents OTP verification request
type OTPVerifyRequest struct {
	AadhaarNumber string `json:"aadhaar_number" binding:"required,len=12"`
	OTP           string `json:"otp" binding:"required,len=6"`
}

// OTPVerifyResponse represents OTP verification response
type OTPVerifyResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Verified bool   `json:"verified"`
}

// VerifyRequest represents identity verification request
type VerifyRequest struct {
	UUID         string `json:"uuid" binding:"required"`
	AadhaarHash  string `json:"aadharhash" binding:"required"`
	DigitalSig   string `json:"Digsig" binding:"required"`
	PublicKey    string `json:"public_key" binding:"required"`
	CombinedHash string `json:"combinedhash" binding:"required"`
	HashFunction string `json:"hashfunction" binding:"required"`
}

// VerifyResponse represents identity verification response
type VerifyResponse struct {
	Verified bool   `json:"verified"`
	Reason   string `json:"reason,omitempty"`
}
