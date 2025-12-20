package models

// AuthenticateRequest from Flutter app
type AuthenticateRequest struct {
	UUID      string `json:"uuid" binding:"required"`
	SessionID string `json:"session_id,omitempty"`
}

// AuthenticateResponse to Flutter/Client
type AuthenticateResponse struct {
	Success       bool   `json:"success"`
	Authenticated bool   `json:"authenticated"`
	Message       string `json:"message,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
}

// Identity from Grid Platform
type Identity struct {
	UUID          string `json:"uuid"`
	AadhaarHash   string `json:"aadharhash"`   // ← Back to old
	DigitalSig    string `json:"Digsig"`       // ← Back to old
	CombinedHash  string `json:"combinedhash"` // ← Back to old
	PublicKey     string `json:"public_key"`
	HashFunction  string `json:"hashfunction"` // ← Back to old
	TwoFARequired bool   `json:"2FA_Required"` // ← Back to old
}

// VerifyRequest to Grid Platform
type VerifyRequest struct {
	UUID         string `json:"uuid"`
	AadhaarHash  string `json:"aadharhash"` // ← Back to old
	DigitalSig   string `json:"Digsig"`     // ← Back to old
	PublicKey    string `json:"public_key"`
	CombinedHash string `json:"combinedhash"` // ← Back to old
	HashFunction string `json:"hashfunction"` // ← Back to old
}

// VerifyResponse from Grid Platform
type VerifyResponse struct {
	Verified bool   `json:"verified"`
	Reason   string `json:"reason,omitempty"`
}
