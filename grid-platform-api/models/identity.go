package models

// Identity defines the schema for the digital identity stored on-chain
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

// OnboardRequest wraps the Identity for the onboarding endpoint
type OnboardRequest struct {
	Identity
}
