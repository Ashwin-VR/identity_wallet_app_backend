package utils

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

// ComputeCombinedHash computes SHA-256 hash of concatenated fields
func ComputeCombinedHash(uuid, digsig, publicKey, hashFunction string) string {
	combined := uuid + digsig + publicKey + hashFunction
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// VerifyECDSASignature verifies ECDSA signature
func VerifyECDSASignature(aadhaarHash, signatureHex, publicKeyHex string) (bool, error) {
	// Decode public key from hex
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false, fmt.Errorf("failed to decode public key: %v", err)
	}

	// Parse ECDSA public key
	pubKeyInterface, err := x509.ParsePKIXPublicKey(pubKeyBytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %v", err)
	}

	pubKey, ok := pubKeyInterface.(*ecdsa.PublicKey)
	if !ok {
		return false, errors.New("not an ECDSA public key")
	}

	// Decode signature from hex
	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %v", err)
	}

	// Parse signature (assuming r||s format, 32 bytes each for P-256)
	if len(sigBytes) != 64 {
		return false, fmt.Errorf("invalid signature length: expected 64, got %d", len(sigBytes))
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	// Hash the aadhaar hash (message to verify)
	messageBytes, err := hex.DecodeString(aadhaarHash)
	if err != nil {
		return false, fmt.Errorf("failed to decode aadhaar hash: %v", err)
	}

	// Verify signature
	valid := ecdsa.Verify(pubKey, messageBytes, r, s)
	return valid, nil
}
