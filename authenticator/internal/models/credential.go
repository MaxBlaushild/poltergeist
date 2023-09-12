package models

import (
	"fmt"

	"github.com/MaxBlaushild/authenticator/internal/encoding"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

type Credential struct {
	gorm.Model
	CredentialID string // base64url
	PublicKey    string // base64url
	AuthUser     AuthUser
	AuthUserID   uint
}

func (c *Credential) ToWebauthnCredential() (*webauthn.Credential, error) {
	publicKey, err := encoding.DecodeBase64UrlEncodedString(c.PublicKey)
	if err != nil {
		fmt.Println("error decoding public key")
		return nil, err
	}

	credentialID, err := encoding.DecodeBase64UrlEncodedString(c.CredentialID)
	if err != nil {
		fmt.Println("error decoding credential id")
		return nil, err
	}
	return &webauthn.Credential{
		ID:        credentialID,
		PublicKey: publicKey,
		Transport: []protocol.AuthenticatorTransport{"internal"},
	}, nil
}
