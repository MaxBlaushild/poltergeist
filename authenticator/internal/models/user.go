package models

import (
	"encoding/binary"
	"strings"
	"unicode"

	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

type AuthUser struct {
	gorm.Model
	Name        string       `json:"name"`
	PhoneNumber string       `json:"phoneNumber"`
	Credentials []Credential `json:"credentials"`
}

func (user *AuthUser) WebAuthnID() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(user.ID))
	return buf
}

func (user *AuthUser) WebAuthnName() string {
	s := strings.ToLower(user.Name)

	var builder strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func (user *AuthUser) WebAuthnDisplayName() string {
	return user.Name
}

func (user *AuthUser) WebAuthnIcon() string {
	return ""
}

func (user *AuthUser) WebAuthnCredentials() []webauthn.Credential {
	var credentials []webauthn.Credential

	for _, cred := range user.Credentials {
		webauthnCred, err := cred.ToWebauthnCredential()
		if err == nil {
			credentials = append(credentials, *webauthnCred)
		}
	}

	return credentials
}
