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
	BankAPI     = "http://localhost:8082"
	P2PServer   = "http://localhost:8083"
	InternalKey = "your_internal_service_key" // MUST match Grid API config
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
	fmt.Println("🚀 DEEP SYSTEM INTEGRATION TEST")

	// 1. SETUP
	userA := createTestUser("User_A")
	userB := createTestUser("User_B")

	// --- SECTION 1: BLOCKCHAIN ONBOARDING ---
	fmt.Println("\n--- [Test 1: Blockchain Onboarding] ---")
	onboardUser(userA)
	onboardUser(userB)

	// --- SECTION 2: BANK FLOW (ENTERPRISE) ---
	fmt.Println("\n--- [Test 2: Bank/Enterprise Flow] ---")
	bankResp := getRequest(BankAPI + "/bank/init-login")
	sessID := bankResp["session_id"].(string)
	fmt.Printf("Bank Session Created: %s\n", sessID)

	fmt.Println("🧐 Fetching challenge through Bank API...")
	chalResp := getRequest(BankAPI + "/api/get-challenge/" + userA.UUID)
	challenge := chalResp["challenge"].(string)

	sigA := base64.StdEncoding.EncodeToString(signData(userA.PrivKey, challenge))
	payload := map[string]string{
		"uuid":       userA.UUID,
		"challenge":  challenge,
		"signature":  sigA,
		"session_id": sessID,
	}
	fmt.Println("🧐 Submitting proof to Bank API...")
	postToAPI(BankAPI+"/api/submit-proof", payload, "")

	statusResp := getRequest(BankAPI + "/bank/status/" + sessID)
	fmt.Printf("🏦 Bank Session Status: %s (UUID: %v)\n", statusResp["status"], statusResp["uuid"])

	// --- SECTION 3: P2P RELAY FLOW ---
	fmt.Println("\n--- [Test 3: P2P Relay Flow] ---")
	p2pInit := postToAPI(P2PServer+"/p2p/create", map[string]string{"prover_uuid": userB.UUID}, "")
	var p2pSess map[string]interface{}
	json.Unmarshal([]byte(p2pInit.body), &p2pSess)
	p2pID := p2pSess["session_id"].(string)
	fmt.Printf("P2P Session ID: %s\n", p2pID)

	p2pChal := "p2p-random-challenge-123"
	postToAPI(P2PServer+"/p2p/post-challenge", map[string]string{"session_id": p2pID, "challenge": p2pChal}, "")

	sigB := base64.StdEncoding.EncodeToString(signData(userB.PrivKey, p2pChal))
	postToAPI(P2PServer+"/p2p/submit-signature", map[string]string{"session_id": p2pID, "signature": sigB}, "")

	p2pFinal := getRequest(P2PServer + "/p2p/status/" + p2pID)
	fmt.Printf("🤝 P2P Handshake Status: %s\n", p2pFinal["status"])

	// --- SECTION 4: SECURITY EDGE CASES ---
	fmt.Println("\n--- [Test 4: Security Edge Cases] ---")

	// Fixed assignment mismatch here (struct return, not multi-value)
	fmt.Println("🧐 Case: Accessing Grid API with WRONG API Key (Should FAIL)")
	gridResp := postToAPI(GridAPI+"/verify", map[string]string{"uuid": userA.UUID}, "WRONG_KEY")
	if gridResp.code == 401 {
		fmt.Println("✅ Success: Grid API rejected unauthorized request.")
	}

	// Fixed unused variable 'badChal' by using it in the check
	fmt.Println("🧐 Case: Signature with Expired/Incorrect Challenge (Should FAIL)")
	badChal := fetchChallenge(userA.UUID)
	fmt.Printf("Generated challenge: %s... now attempting verification with wrong data\n", badChal[:8])

	// We use sigA (signed for 'challenge') against 'badChal'
	verifyResult(userA.UUID, badChal, sigA, "Signature-Challenge Mismatch")
}

// --- HELPERS ---

func createTestUser(label string) TestUser {
	priv, _ := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	pubHex := fmt.Sprintf("%064x%064x", priv.PublicKey.X, priv.PublicKey.Y)
	u := uuid.New().String()
	fmt.Printf("Created %s: %s\n", label, u)
	return TestUser{UUID: u, PrivKey: priv, PubKey: pubHex}
}

func onboardUser(u TestUser) {
	hAadhaar := sha256.Sum256([]byte(u.UUID))
	aHash := hex.EncodeToString(hAadhaar[:])
	sig := base64.StdEncoding.EncodeToString(signData(u.PrivKey, aHash))
	data := u.UUID + aHash + sig + u.PubKey + "SHA256"
	hCombined := sha256.Sum256([]byte(data))
	cHash := hex.EncodeToString(hCombined[:])

	id := Identity{
		UUID: u.UUID, AadhaarHash: aHash, DigitalSig: sig,
		PublicKey: u.PubKey, HashFunction: "SHA256", CombinedHash: cHash,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	postToAPI(GridAPI+"/create", id, "")
}

func fetchChallenge(uid string) string {
	resp := getRequest(VerifierAPI + "/challenge/" + uid)
	return resp["challenge"].(string)
}

func verifyResult(uid, chal, sig, testName string) {
	payload := map[string]string{"uuid": uid, "challenge": chal, "signature": sig}
	resp := postToAPI(VerifierAPI+"/verify-uuid", payload, "")
	if strings.Contains(resp.body, `"valid":true`) {
		fmt.Printf("🟢 %s: VALID\n", testName)
	} else {
		fmt.Printf("🔴 %s: INVALID\n", testName)
	}
}

func signData(priv *ecdsa.PrivateKey, data string) []byte {
	hash := sha256.Sum256([]byte(data))
	r, s, _ := ecdsa.Sign(rand.Reader, priv, hash[:])
	sig := append(r.Bytes(), s.Bytes()...)
	return sig
}

type apiResponse struct {
	code int
	body string
}

func postToAPI(url string, payload interface{}, key string) apiResponse {
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("X-API-KEY", key)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("API Offline: %s", url)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return apiResponse{resp.StatusCode, string(body)}
}

func getRequest(url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("API Offline: %s", url)
	}
	defer resp.Body.Close()
	var r map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&r)
	return r
}
