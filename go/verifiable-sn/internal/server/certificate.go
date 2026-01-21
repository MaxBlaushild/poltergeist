package server

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	ethereum_transactor "github.com/MaxBlaushild/poltergeist/pkg/ethereum_transactor"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EnrollCertificateRequest struct {
	PublicKey          string `json:"publicKey" binding:"required"`
	ChallengeSignature string `json:"challengeSignature" binding:"required"`
}

type EnrollCertificateResponse struct {
	CertificatePEM string `json:"certificatePem"`
	Fingerprint    string `json:"fingerprint"`
	PublicKey      string `json:"publicKey"`
}

func (s *server) CheckCertificate(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	cert, err := s.dbClient.UserCertificate().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	hasCertificate := cert != nil
	ctx.JSON(http.StatusOK, gin.H{
		"hasCertificate": hasCertificate,
	})
}

func (s *server) EnrollCertificate(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if user already has a certificate
	existingCert, err := s.dbClient.UserCertificate().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if existingCert != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "user already has a certificate",
		})
		return
	}

	var requestBody EnrollCertificateRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse the public key from PEM
	block, _ := pem.Decode([]byte(requestBody.PublicKey))
	if block == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to decode public key PEM",
		})
		return
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to parse public key: %v", err),
		})
		return
	}

	// Verify proof of possession: decode and verify the challenge signature
	challengeSignature, err := base64.StdEncoding.DecodeString(requestBody.ChallengeSignature)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to decode challenge signature",
		})
		return
	}

	// Create a challenge based on the user ID and public key
	challenge := createChallenge(user.ID, requestBody.PublicKey)

	// Verify the signature
	if err := verifySignature(challenge, challengeSignature, publicKey); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid signature: %v", err),
		})
		return
	}

	// Issue certificate (1 year validity)
	certificateDER, certificatePEM, fingerprint, err := s.certClient.IssueCertificate(requestBody.PublicKey, user.ID, 365*24*time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to issue certificate: %v", err),
		})
		return
	}

	// Store certificate in database (created as inactive by default)
	_, err = s.dbClient.UserCertificate().Create(ctx, user.ID, certificateDER, certificatePEM, requestBody.PublicKey, fingerprint)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to store certificate: %v", err),
		})
		return
	}

	// Extract issuer and subject from X.509 certificate
	issuer, subject, err := extractIssuerAndSubject(certificateDER)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to extract certificate fields: %v", err),
		})
		return
	}

	// Generate ABI-encoded registerCertificate function call
	encodedData, err := encodeRegisterCertificateCall(fingerprint, issuer, subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to encode function call: %v", err),
		})
		return
	}

	// Create blockchain transaction via ethereum-transactor service
	if s.ethereumTransactorClient != nil && s.c2PAContractAddress != "" {
		dataHex := "0x" + hex.EncodeToString(encodedData)
		_, err := s.ethereumTransactorClient.CreateTransaction(ctx, ethereum_transactor.CreateTransactionRequest{
			To:   &s.c2PAContractAddress,
			Value: "0",
			Data: &dataHex,
		})
		if err != nil {
			// Log error but don't fail the enrollment - certificate is created, just not registered on-chain yet
			// In production, you might want to queue this for retry
			fmt.Printf("Warning: failed to create blockchain transaction: %v\n", err)
		}
		// Note: The ethereum-transactor service stores the transaction in its database.
		// The job runner will find transactions by type "registerCertificate" and match fingerprints
		// from the transaction data to activate certificates when they confirm.
	}

	ctx.JSON(http.StatusOK, EnrollCertificateResponse{
		CertificatePEM: certificatePEM,
		Fingerprint:    fmt.Sprintf("%x", fingerprint),
		PublicKey:      requestBody.PublicKey,
	})
}

func (s *server) GetCertificate(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	cert, err := s.dbClient.UserCertificate().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if cert == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "certificate not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"certificatePem": cert.CertificatePEM,
		"fingerprint":    fmt.Sprintf("%x", cert.Fingerprint),
		"publicKey":      cert.PublicKey,
		"createdAt":      cert.CreatedAt,
	})
}

// createChallenge creates a deterministic challenge based on user ID and public key
func createChallenge(userID uuid.UUID, publicKeyPEM string) []byte {
	data := fmt.Sprintf("%s:%s", userID.String(), publicKeyPEM)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// verifySignature verifies an ECDSA signature over a message hash
func verifySignature(messageHash []byte, signature []byte, publicKey crypto.PublicKey) error {
	ecdsaPubKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key is not ECDSA")
	}

	// ECDSA signature from Secure Enclave is typically 64 bytes (r and s, each 32 bytes)
	if len(signature) != 64 {
		return fmt.Errorf("invalid signature length: expected 64 bytes, got %d", len(signature))
	}

	// Split signature into r and s
	rBytes := signature[:32]
	sBytes := signature[32:]

	// Convert bytes to big.Int
	var rInt, sInt big.Int
	rInt.SetBytes(rBytes)
	sInt.SetBytes(sBytes)

	// Create ASN.1 DER encoded signature
	sigDER, err := encodeECDSASignature(&rInt, &sInt)
	if err != nil {
		return err
	}

	valid := ecdsa.VerifyASN1(ecdsaPubKey, messageHash, sigDER)
	if !valid {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// encodeECDSASignature encodes r and s as ASN.1 DER for ECDSA signature verification
func encodeECDSASignature(r, s *big.Int) ([]byte, error) {
	type ecdsaSignature struct {
		R, S *big.Int
	}

	sig := ecdsaSignature{
		R: r,
		S: s,
	}

	return asn1.Marshal(sig)
}

// encodeRegisterCertificateCall ABI encodes the registerCertificate(bytes32,string,string) function call
func encodeRegisterCertificateCall(fingerprint []byte, issuer string, subject string) ([]byte, error) {
	// Function signature: registerCertificate(bytes32,string,string)
	// Function selector: keccak256("registerCertificate(bytes32,string,string)")[:4]
	// We'll use the ABI package to encode this properly
	
	// Define the ABI for the function
	abiJSON := `[{"constant":false,"inputs":[{"name":"fingerprint","type":"bytes32"},{"name":"issuer","type":"string"},{"name":"subject","type":"string"}],"name":"registerCertificate","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`
	
	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Convert fingerprint to bytes32
	var fingerprintBytes32 [32]byte
	copy(fingerprintBytes32[:], fingerprint)

	// Encode the function call
	data, err := contractABI.Pack("registerCertificate", fingerprintBytes32, issuer, subject)
	if err != nil {
		return nil, fmt.Errorf("failed to pack function call: %w", err)
	}

	return data, nil
}

// extractIssuerAndSubject extracts issuer and subject strings from an X.509 certificate
func extractIssuerAndSubject(certDER []byte) (string, string, error) {
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Extract issuer - format as "CN=..., O=..., ..."
	issuer := cert.Issuer.String()
	
	// Extract subject - format as "CN=..., O=..., ..."
	subject := cert.Subject.String()

	// If the strings are empty, provide defaults
	if issuer == "" {
		issuer = "Unknown Issuer"
	}
	if subject == "" {
		subject = "Unknown Subject"
	}

	return issuer, subject, nil
}
