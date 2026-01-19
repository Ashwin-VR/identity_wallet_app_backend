package fabric

import (
	"crypto/x509"
	"fmt"
	"os"
	"path"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	contract *client.Contract
}

func NewClient() (*Client, error) {
	mspID := getenv("FABRIC_MSP_ID", "Org1MSP")
	cryptoPath := getenv("FABRIC_CRYPTO_PATH", "../../test-network/organizations/peerOrganizations/org1.example.com")
	certDir := getenv("FABRIC_CERT_DIR", cryptoPath+"/users/User1@org1.example.com/msp/signcerts")
	keyDir := getenv("FABRIC_KEY_DIR", cryptoPath+"/users/User1@org1.example.com/msp/keystore")
	tlsCertPath := getenv("FABRIC_TLS_CERT_PATH", cryptoPath+"/peers/peer0.org1.example.com/tls/ca.crt")
	peerEndpoint := getenv("FABRIC_PEER_ENDPOINT", "dns:///localhost:7051")
	gatewayPeer := getenv("FABRIC_GATEWAY_PEER", "peer0.org1.example.com")
	channelName := getenv("FABRIC_CHANNEL_NAME", "mychannel")
	chaincodeName := getenv("FABRIC_CHAINCODE_NAME", "identity")

	conn, err := newGrpcConnection(tlsCertPath, peerEndpoint, gatewayPeer)
	if err != nil {
		return nil, err
	}

	id, err := newIdentity(mspID, certDir)
	if err != nil {
		return nil, err
	}

	sign, err := newSign(keyDir)
	if err != nil {
		return nil, err
	}

	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithHash(hash.SHA256),
		client.WithClientConnection(conn),
	)
	if err != nil {
		return nil, err
	}

	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	return &Client{contract: contract}, nil
}

func (c *Client) OnboardIdentity(aadhar, pubKey string) error {
	_, err := c.contract.SubmitTransaction("OnboardIdentity", aadhar, pubKey)
	return err
}

func (c *Client) GetIdentity(aadhar string) (string, error) {
	res, err := c.contract.EvaluateTransaction("GetIdentity", aadhar)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

// helpers

func newGrpcConnection(tlsCertPath, peerEndpoint, gatewayPeer string) (*grpc.ClientConn, error) {
	certPEM, err := os.ReadFile(tlsCertPath)
	if err != nil {
		return nil, fmt.Errorf("read TLS cert: %w", err)
	}

	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	cp := x509.NewCertPool()
	cp.AddCert(cert)
	creds := credentials.NewClientTLSFromCert(cp, gatewayPeer)

	conn, err := grpc.NewClient(peerEndpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("grpc connect: %w", err)
	}
	return conn, nil
}

func newIdentity(mspID, certDir string) (*identity.X509Identity, error) {
	certPEM, err := readFirstFile(certDir)
	if err != nil {
		return nil, err
	}

	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	return identity.NewX509Identity(mspID, cert)
}

func newSign(keyDir string) (identity.Sign, error) {
	keyPEM, err := readFirstFile(keyDir)
	if err != nil {
		return nil, err
	}

	privKey, err := identity.PrivateKeyFromPEM(keyPEM)
	if err != nil {
		return nil, err
	}

	return identity.NewPrivateKeySign(privKey)
}

func readFirstFile(dirPath string) ([]byte, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	names, err := dir.Readdirnames(1)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path.Join(dirPath, names[0]))
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
