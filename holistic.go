package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/google/uuid"
)

const (
	GridAPI     = "http://localhost:8080"
	VerifierAPI = "http://localhost:8081"
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

type TestUser struct {
	UUID    string
	PrivKey *ecdsa.PrivateKey
	PubKey  string
}

func main() {
	fmt.Println("🧪 Starting Holistic Edge-Case Testing...")

	// 1. SETUP TWO IDENTITIES
	userA := createTestUser("User_A")
	userB := createTestUser("User_B")

	fmt.Println("\n--- [Test 1: Parallel Onboarding] ---")
	onboardUser(userA)
	onboardUser(userB)

	// 2. TEST CASE: CROSS-VERIFICATION (EXPECTED FAIL)
	fmt.Println("\n--- [Test 2: Cross-Verification Attack (Should FAIL)] ---")
	chalA := fetchChallenge(userA.UUID)
	// Sign User A's challenge with User B's private key
	sigB := base64.StdEncoding.EncodeToString(signData(userB.PrivKey, chalA))
	verifyResult(userA.UUID, chalA, sigB, "Cross-User Signing")

	// 3. TEST CASE: TAMPERED CHALLENGE (EXPECTED FAIL)
	fmt.Println("\n--- [Test 3: Tampered Challenge (Should FAIL)] ---")
	chalB := fetchChallenge(userB.UUID)
	sigActualB := base64.StdEncoding.EncodeToString(signData(userB.PrivKey, chalB))
	// Change one character in the challenge before sending to API
	tamperedChal := chalB[:len(chalB)-1] + "0"
	verifyResult(userB.UUID, tamperedChal, sigActualB, "Tampered Challenge Data")

	// 4. TEST CASE: DUPLICATE REGISTRATION (EXPECTED FAIL)
	fmt.Println("\n--- [Test 4: Duplicate UUID Registration (Should FAIL)] ---")
	status, _ := onboardUser(userA)
	if status != 200 {
		fmt.Printf("✅ Success: Blockchain rejected duplicate UUID (Status: %d)\n", status)
	} else {
		fmt.Println("❌ Failure: Blockchain allowed duplicate registration!")
	}

	// 5. TEST CASE: VALID FLOW FOR USER B (EXPECTED SUCCESS)
	fmt.Println("\n--- [Test 5: Valid Handshake for User B (Should PASS)] ---")
	verifyResult(userB.UUID, chalB, sigActualB, "Valid Logic")
}

// --- HELPER FUNCTIONS ---

func createTestUser(label string) TestUser {
	priv, _ := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	pubHex := fmt.Sprintf("04%064x%064x", priv.PublicKey.X, priv.PublicKey.Y)
	u := uuid.New().String()
	fmt.Printf("Created %s: %s\n", label, u)
	return TestUser{UUID: u, PrivKey: priv, PubKey: pubHex}
}

func onboardUser(u TestUser) (int, string) {
	// Fix: Separate hash calculation to ensure it is addressable
	hAadhaar := sha256.Sum256([]byte(u.UUID))
	aHash := hex.EncodeToString(hAadhaar[:])

	sig := base64.StdEncoding.EncodeToString(signData(u.PrivKey, aHash))
	cStr := u.UUID + aHash + sig + u.PubKey + "SHA256"

	hCombined := sha256.Sum256([]byte(cStr))
	cHash := hex.EncodeToString(hCombined[:])

	id := Identity{
		UUID: u.UUID, AadhaarHash: aHash, DigitalSig: sig,
		PublicKey: u.PubKey, HashFunction: "SHA256", CombinedHash: cHash,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	return postToAPI(GridAPI+"/create", id)
}

func fetchChallenge(uid string) string {
	resp := getRequest(VerifierAPI + "/challenge/" + uid)
	chal, _ := resp["challenge"].(string)
	return chal
}

func verifyResult(uid, chal, sig, testName string) {
	fmt.Printf("🧐 Running: %s\n", testName)
	payload := map[string]string{"uuid": uid, "challenge": chal, "signature": sig}
	_, body := postToAPI(VerifierAPI+"/verify-uuid", payload)

	if strings.Contains(body, `"valid":true`) {
		fmt.Println("🟢 RESULT: VALID")
	} else {
		fmt.Println("🔴 RESULT: INVALID")
	}
}

func signData(priv *ecdsa.PrivateKey, data string) []byte {
	hash := sha256.Sum256([]byte(data))
	r, s, _ := ecdsa.Sign(rand.Reader, priv, hash[:])

	signature := make([]byte, 64)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(signature[32-len(rBytes):32], rBytes)
	copy(signature[64-len(sBytes):64], sBytes)
	return signature
}

func postToAPI(url string, payload interface{}) (int, string) {
	b, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func getRequest(url string) map[string]interface{} {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	var r map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&r)
	return r
}
