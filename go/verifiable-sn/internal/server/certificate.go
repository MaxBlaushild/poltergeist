package server

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"time"

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

	// Store certificate in database
	_, err = s.dbClient.UserCertificate().Create(ctx, user.ID, certificateDER, certificatePEM, requestBody.PublicKey, fingerprint)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to store certificate: %v", err),
		})
		return
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
