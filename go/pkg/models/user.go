package models

import (
	"strings"
	"time"
	"unicode"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID    `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	Name        string       `json:"name"`
	PhoneNumber string       `json:"phoneNumber" gorm:"unique"`
	Credentials []Credential `json:"credentials"`
	Active      bool         `json:"active"`
}

func (user *User) WebAuthnID() []byte {
	return []byte(user.ID.String())
}

func (user *User) WebAuthnName() string {
	s := strings.ToLower(user.Name)

	var builder strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func (user *User) WebAuthnDisplayName() string {
	return user.Name
}

func (user *User) WebAuthnIcon() string {
	return ""
}

func (user *User) WebAuthnCredentials() []webauthn.Credential {
	var credentials []webauthn.Credential

	for _, cred := range user.Credentials {
		webauthnCred, err := cred.ToWebauthnCredential()
		if err == nil {
			credentials = append(credentials, *webauthnCred)
		}
	}

	return credentials
}
