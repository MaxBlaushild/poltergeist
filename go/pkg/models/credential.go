package models

import (
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/encoding"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type Credential struct {
	ID           uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	CredentialID string    // base64url
	PublicKey    string    // base64url
	User         User
	UserID       uuid.UUID
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
