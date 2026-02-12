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
	"time"

	"github.com/fxamacker/cbor/v2"
)

// ExtractManifestTimestamp tries to read the C2PA claim created timestamp from the manifest.
// Returns nil if no timestamp is found.
func ExtractManifestTimestamp(manifestBytes []byte) (*time.Time, error) {
	var manifestInterface interface{}
	if err := cbor.Unmarshal(manifestBytes, &manifestInterface); err != nil {
		return nil, fmt.Errorf("failed to parse CBOR manifest (size: %d bytes): %w", len(manifestBytes), err)
	}

	manifestMap, ok := coerceStringMap(manifestInterface)
	if !ok || len(manifestMap) == 0 {
		return nil, fmt.Errorf("manifest is not a map (type: %T)", manifestInterface)
	}

	claimInterface, exists := manifestMap["claim"]
	if !exists {
		return nil, fmt.Errorf("manifest missing claim field")
	}

	claimMap, ok := coerceStringMap(claimInterface)
	if !ok {
		return nil, fmt.Errorf("manifest claim has invalid type: %T", claimInterface)
	}

	createdValue, exists := claimMap["created"]
	if !exists || createdValue == nil {
		return nil, nil
	}

	parsed, err := parseManifestTime(createdValue)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

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

	// Parse CBOR manifest into a generic map first to handle flexible structure
	var manifestInterface interface{}
	if err := cbor.Unmarshal(manifestBytes, &manifestInterface); err != nil {
		return nil, nil, fmt.Errorf("failed to parse CBOR manifest (size: %d bytes): %w", len(manifestBytes), err)
	}

	// Convert to map[string]interface{} handling both types
	var manifestMap map[string]interface{}
	if m, ok := manifestInterface.(map[string]interface{}); ok {
		manifestMap = m
	} else if m, ok := manifestInterface.(map[interface{}]interface{}); ok {
		// Convert map[interface{}]interface{} to map[string]interface{}
		manifestMap = make(map[string]interface{})
		for k, v := range m {
			if keyStr, ok := k.(string); ok {
				manifestMap[keyStr] = v
			}
		}
	} else {
		return nil, nil, fmt.Errorf("manifest is not a map (type: %T)", manifestInterface)
	}

	if len(manifestMap) == 0 {
		return nil, nil, fmt.Errorf("manifest map is empty")
	}

	// Validate manifest structure
	claimInterface, exists := manifestMap["claim"]
	if !exists {
		fmt.Printf("Manifest keys: %v\n", getMapKeysFromInterface(manifestMap))
		return nil, nil, fmt.Errorf("manifest missing claim field")
	}

	var claim map[string]interface{}
	if c, ok := claimInterface.(map[string]interface{}); ok {
		claim = c
	} else if c, ok := claimInterface.(map[interface{}]interface{}); ok {
		// Convert map[interface{}]interface{} to map[string]interface{}
		claim = make(map[string]interface{})
		for k, v := range c {
			if keyStr, ok := k.(string); ok {
				claim[keyStr] = v
			}
		}
	} else {
		fmt.Printf("Claim type: %T, value: %v\n", claimInterface, claimInterface)
		return nil, nil, fmt.Errorf("manifest claim has invalid type: %T", claimInterface)
	}

	if claim == nil || len(claim) == 0 {
		return nil, nil, fmt.Errorf("manifest claim is empty")
	}

	assertionsInterface, exists := manifestMap["assertions"]
	if !exists {
		return nil, nil, fmt.Errorf("manifest missing assertions field")
	}

	var assertions []interface{}
	switch v := assertionsInterface.(type) {
	case []interface{}:
		assertions = v
	case []map[string]interface{}:
		// Convert to []interface{}
		assertions = make([]interface{}, len(v))
		for i, item := range v {
			assertions[i] = item
		}
	case []map[interface{}]interface{}:
		// Convert to []interface{}
		assertions = make([]interface{}, len(v))
		for i, item := range v {
			assertions[i] = item
		}
	default:
		return nil, nil, fmt.Errorf("assertions has invalid type: %T", assertionsInterface)
	}

	if len(assertions) == 0 {
		return nil, nil, fmt.Errorf("manifest assertions array is empty")
	}

	// Find signature assertion
	var signatureAssertion map[string]interface{}
	for _, assertionInterface := range assertions {
		var assertion map[string]interface{}

		// Handle map[string]interface{}
		if a, ok := assertionInterface.(map[string]interface{}); ok {
			assertion = a
		} else if a, ok := assertionInterface.(map[interface{}]interface{}); ok {
			// Convert map[interface{}]interface{} to map[string]interface{}
			assertion = make(map[string]interface{})
			for k, v := range a {
				if keyStr, ok := k.(string); ok {
					assertion[keyStr] = v
				}
			}
		} else {
			fmt.Printf("Assertion has unexpected type: %T\n", assertionInterface)
			continue
		}

		label, ok := assertion["label"].(string)
		if !ok {
			continue
		}

		if label == "c2pa.signature" {
			signatureAssertion = assertion
			break
		}
	}

	if signatureAssertion == nil {
		return nil, nil, fmt.Errorf("manifest missing signature assertion")
	}

	// Extract certificate from signature assertion
	// Debug: log the signature assertion structure
	fmt.Printf("Signature assertion keys: %v\n", getMapKeys(signatureAssertion))
	if dataInterface, exists := signatureAssertion["data"]; exists {
		fmt.Printf("Data type: %T, value: %v\n", dataInterface, dataInterface)
	}

	data, ok := signatureAssertion["data"].(map[string]interface{})
	if !ok {
		// Try to handle if data is encoded differently
		if dataMap, ok2 := signatureAssertion["data"].(map[interface{}]interface{}); ok2 {
			// Convert map[interface{}]interface{} to map[string]interface{}
			data = make(map[string]interface{})
			for k, v := range dataMap {
				if keyStr, ok := k.(string); ok {
					data[keyStr] = v
				}
			}
		} else {
			return nil, nil, fmt.Errorf("invalid signature assertion data (type: %T)", signatureAssertion["data"])
		}
	}

	certChainInterface, exists := data["cert_chain"]
	if !exists {
		return nil, nil, fmt.Errorf("signature assertion missing cert_chain field")
	}

	var certChain []interface{}
	switch v := certChainInterface.(type) {
	case []interface{}:
		certChain = v
	case []map[string]interface{}:
		// Convert to []interface{}
		certChain = make([]interface{}, len(v))
		for i, item := range v {
			certChain[i] = item
		}
	default:
		return nil, nil, fmt.Errorf("cert_chain has invalid type: %T", certChainInterface)
	}

	if len(certChain) == 0 {
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
	signatureInterface, exists := data["signature"]
	if !exists {
		return nil, nil, fmt.Errorf("signature assertion missing signature field")
	}

	signature, ok := signatureInterface.(string)
	if !ok {
		return nil, nil, fmt.Errorf("signature has invalid type: %T, expected string", signatureInterface)
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
	_, err := url.Parse(manifestURI)
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

func coerceStringMap(value interface{}) (map[string]interface{}, bool) {
	switch m := value.(type) {
	case map[string]interface{}:
		return m, true
	case map[interface{}]interface{}:
		converted := make(map[string]interface{}, len(m))
		for k, v := range m {
			if keyStr, ok := k.(string); ok {
				converted[keyStr] = v
			}
		}
		return converted, true
	default:
		return nil, false
	}
}

func parseManifestTime(value interface{}) (*time.Time, error) {
	switch v := value.(type) {
	case string:
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return &t, nil
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return &t, nil
		}
		return nil, fmt.Errorf("unsupported created timestamp format: %q", v)
	case []byte:
		str := string(v)
		if t, err := time.Parse(time.RFC3339Nano, str); err == nil {
			return &t, nil
		}
		if t, err := time.Parse(time.RFC3339, str); err == nil {
			return &t, nil
		}
		return nil, fmt.Errorf("unsupported created timestamp format: %q", str)
	case int64:
		t := time.Unix(v, 0).UTC()
		return &t, nil
	case uint64:
		t := time.Unix(int64(v), 0).UTC()
		return &t, nil
	case int:
		t := time.Unix(int64(v), 0).UTC()
		return &t, nil
	case float64:
		t := time.Unix(int64(v), 0).UTC()
		return &t, nil
	default:
		return nil, fmt.Errorf("unsupported created timestamp type: %T", value)
	}
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

// getMapKeys returns all keys from a map for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getMapKeysFromInterface returns all keys from a map[interface{}]interface{} for debugging
func getMapKeysFromInterface(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
