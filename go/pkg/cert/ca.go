package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

type Client interface {
	IssueCertificate(publicKeyPEM string, userID uuid.UUID, validityPeriod time.Duration) (certificateDER []byte, certificatePEM string, fingerprint []byte, err error)
	ComputeFingerprint(certificateDER []byte) []byte
	GetCACertificate() *x509.Certificate
}

type client struct {
	caCert       *x509.Certificate
	caPrivateKey *rsa.PrivateKey
}

// NewClient creates a new CA client. If caPrivateKeyPEM is empty, it generates a new CA.
// If caPrivateKeyPEM is provided, it loads the CA from the PEM-encoded private key.
func NewClient(caPrivateKeyPEM string) (Client, error) {
	var caCert *x509.Certificate
	var caPrivateKey *rsa.PrivateKey
	var err error

	if caPrivateKeyPEM == "" {
		// Generate new CA
		caCert, caPrivateKey, err = generateCA()
		if err != nil {
			return nil, fmt.Errorf("failed to generate CA: %w", err)
		}
	} else {
		// Load CA from PEM
		block, _ := pem.Decode([]byte(caPrivateKeyPEM))
		if block == nil {
			return nil, fmt.Errorf("failed to decode CA private key PEM")
		}

		caPrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CA private key: %w", err)
		}

		// Generate CA certificate from the private key
		caCert, err = generateCACertificate(caPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to generate CA certificate: %w", err)
		}
	}

	return &client{
		caCert:       caCert,
		caPrivateKey: caPrivateKey,
	}, nil
}

func generateCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	// Generate CA private key
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Generate CA certificate
	caCert, err := generateCACertificate(caPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	return caCert, caPrivateKey, nil
}

func generateCACertificate(caPrivateKey *rsa.PrivateKey) (*x509.Certificate, error) {
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Verifiable SN"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "Verifiable SN CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years validity
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, err
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return nil, err
	}

	return caCert, nil
}

func (c *client) IssueCertificate(publicKeyPEM string, userID uuid.UUID, validityPeriod time.Duration) ([]byte, string, []byte, error) {
	// Parse the public key from PEM
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, "", nil, fmt.Errorf("failed to decode public key PEM")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, big.NewInt(0).Exp(big.NewInt(2), big.NewInt(159), nil))
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Verifiable SN"},
			Country:      []string{"US"},
			CommonName:   userID.String(),
			SerialNumber: userID.String(),
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validityPeriod),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: false,
		IsCA:                  false,
	}

	// Sign the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, c.caCert, publicKey, c.caPrivateKey)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Convert to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Compute fingerprint
	fingerprint := c.ComputeFingerprint(certDER)

	return certDER, string(certPEM), fingerprint, nil
}

func (c *client) ComputeFingerprint(certificateDER []byte) []byte {
	hash := sha256.Sum256(certificateDER)
	return hash[:]
}

func (c *client) GetCACertificate() *x509.Certificate {
	return c.caCert
}
