package main

import (
	"fmt"
	"grid-platform-api/fabric"
	"grid-platform-api/handlers"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// 1. Setup Fabric Connection
	clientConn, err := fabric.NewGrpcConnection(
		os.Getenv("FABRIC_TLS_CERT_PATH"),
		os.Getenv("FABRIC_PEER_ENDPOINT"),
		os.Getenv("FABRIC_GATEWAY_PEER"),
	)
	if err != nil {
		log.Fatalf("Failed to create gRPC connection: %v", err)
	}
	defer clientConn.Close()

	id, err := fabric.NewIdentity(os.Getenv("FABRIC_CERT_PATH"), os.Getenv("FABRIC_MSP_ID"))
	if err != nil {
		log.Fatalf("Failed to create identity: %v", err)
	}

	sign, err := fabric.NewSign(os.Getenv("FABRIC_KEY_PATH"))
	if err != nil {
		log.Fatalf("Failed to create sign: %v", err)
	}

	// 2. Create Gateway
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConn),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network := gw.GetNetwork(os.Getenv("FABRIC_CHANNEL"))
	contract := network.GetContract(os.Getenv("FABRIC_CHAINCODE"))

	// 3. Routes
	r := gin.Default()

	// Public onboarding
	r.POST("/create", handlers.CreateIdentityHandler(contract))

	// Internal verification (Requires API Key)
	r.POST("/verify", handlers.VerifyIdentityHandler(contract))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Grid Platform API running on :%s\n", port)
	r.Run("0.0.0.0:" + port)
}
