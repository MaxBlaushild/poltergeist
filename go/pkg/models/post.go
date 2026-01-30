package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt       time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"not null" json:"updatedAt"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	ImageURL        string    `gorm:"type:text;not null" json:"imageUrl"`
	MediaType       *string   `gorm:"type:text" json:"mediaType,omitempty"` // "image" or "video"
	Caption         *string   `gorm:"type:text" json:"caption,omitempty"`
	ManifestHash    []byte    `gorm:"type:bytea" json:"manifestHash,omitempty"`
	ManifestURI     *string   `gorm:"type:text" json:"manifestUri,omitempty"`
	CertFingerprint []byte    `gorm:"type:bytea" json:"certFingerprint,omitempty"`
	AssetID         *string   `gorm:"type:text" json:"assetId,omitempty"`
	Tags            []string  `gorm:"-" json:"tags,omitempty"` // Populated from post_tags, not stored on posts table
}

