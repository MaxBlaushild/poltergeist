package server

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/fxamacker/cbor/v2"
)

// C2PAManifest represents a simplified C2PA manifest structure
type C2PAManifest struct {
	Claim      map[string]interface{}   `cbor:"claim"`
	Assertions []map[string]interface{} `cbor:"assertions"`
	Signatures []map[string]interface{} `cbor:"signatures,omitempty"`
}

// ValidateManifest validates a C2PA manifest
// Returns the manifest hash, certificate fingerprint, and any validation errors
func ValidateManifest(manifestBytes []byte) (manifestHash []byte, certFingerprint []byte, err error) {
	// Compute manifest hash
	hash := sha256.Sum256(manifestBytes)
	manifestHash = hash[:]

	// Parse CBOR manifest
	var manifest C2PAManifest
	if err := cbor.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, nil, fmt.Errorf("failed to parse CBOR manifest: %w", err)
	}

	// Validate manifest structure
	if manifest.Claim == nil {
		return nil, nil, fmt.Errorf("manifest missing claim")
	}

	if len(manifest.Assertions) == 0 {
		return nil, nil, fmt.Errorf("manifest missing assertions")
	}

	// Find signature assertion
	var signatureAssertion map[string]interface{}
	for _, assertion := range manifest.Assertions {
		if label, ok := assertion["label"].(string); ok && label == "c2pa.signature" {
			signatureAssertion = assertion
			break
		}
	}

	if signatureAssertion == nil {
		return nil, nil, fmt.Errorf("manifest missing signature assertion")
	}

	// Extract certificate from signature assertion
	data, ok := signatureAssertion["data"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("invalid signature assertion data")
	}

	certChain, ok := data["cert_chain"].([]interface{})
	if !ok || len(certChain) == 0 {
		return nil, nil, fmt.Errorf("signature assertion missing certificate chain")
	}

	// Get first certificate (leaf certificate)
	certBase64, ok := certChain[0].(string)
	if !ok {
		return nil, nil, fmt.Errorf("invalid certificate format in chain")
	}

	// Decode certificate
	certDER, err := base64.StdEncoding.DecodeString(certBase64)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode certificate: %w", err)
	}

	// Parse X.509 certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse X.509 certificate: %w", err)
	}

	// Validate certificate chain (basic validation)
	// In a full implementation, we would validate the entire chain
	if cert == nil {
		return nil, nil, fmt.Errorf("invalid certificate")
	}

	// Compute certificate fingerprint
	certFingerprintHash := sha256.Sum256(certDER)
	certFingerprint = certFingerprintHash[:]

	// Verify signature (basic check - in full implementation, verify actual signature)
	signature, ok := data["signature"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("signature assertion missing signature")
	}

	// Decode signature
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	if len(signatureBytes) == 0 {
		return nil, nil, fmt.Errorf("empty signature")
	}

	// Note: Full signature verification would require:
	// 1. Extract the signed payload from the manifest
	// 2. Verify the ECDSA signature using the certificate's public key
	// 3. Validate the certificate chain against a trusted root
	// For now, we're doing basic structure validation

	return manifestHash, certFingerprint, nil
}

// DownloadManifestFromS3 downloads a manifest from an S3 URL
func DownloadManifestFromS3(manifestURI string) ([]byte, error) {
	// Parse URL
	parsedURL, err := url.Parse(manifestURI)
	if err != nil {
		return nil, fmt.Errorf("invalid manifest URI: %w", err)
	}

	// Download from URL (works for S3 public URLs or presigned URLs)
	resp, err := http.Get(manifestURI)
	if err != nil {
		return nil, fmt.Errorf("failed to download manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download manifest: status %d", resp.StatusCode)
	}

	// Read manifest bytes
	manifestBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	return manifestBytes, nil
}

// HexToBytes converts a hex string to bytes
func HexToBytes(hexStr string) ([]byte, error) {
	// Remove 0x prefix if present
	cleanHex := strings.TrimPrefix(hexStr, "0x")
	cleanHex = strings.ReplaceAll(cleanHex, " ", "")

	bytes, err := hex.DecodeString(cleanHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}

	return bytes, nil
}

// BytesToHex converts bytes to hex string
func BytesToHex(bytes []byte) string {
	return hex.EncodeToString(bytes)
}
