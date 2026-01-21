package models

import (
	"time"

	"github.com/google/uuid"
)

type UserCertificate struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"not null" json:"updatedAt"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`
	Certificate   []byte    `gorm:"type:bytea;not null" json:"-"`
	CertificatePEM string   `gorm:"type:text;not null" json:"certificatePem"`
	PublicKey     string    `gorm:"type:text;not null" json:"publicKey"`
	Fingerprint   []byte    `gorm:"type:bytea;not null;index" json:"fingerprint"`
	Active        bool      `gorm:"type:boolean;not null;default:false;index" json:"active"`
}
