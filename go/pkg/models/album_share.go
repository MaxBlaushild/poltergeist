package models

import (
	"time"

	"github.com/google/uuid"
)

type AlbumShare struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	AlbumID   uuid.UUID `gorm:"type:uuid;not null;index" json:"albumId"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null;index" json:"createdBy"`
	Token     string    `gorm:"type:text;not null;uniqueIndex" json:"token"`
}
