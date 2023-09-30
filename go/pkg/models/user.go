package models

import (
	"encoding/binary"
	"strings"
	"unicode"

	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name        string       `json:"name"`
	PhoneNumber string       `json:"phoneNumber" gorm:"unique"`
	Credentials []Credential `json:"credentials"`
	Active      bool         `json:"active"`
}

func (user *User) WebAuthnID() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(user.ID))
	return buf
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

func (u *User) TableName() string {
	return "geist_users"
}
