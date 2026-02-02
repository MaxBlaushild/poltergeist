package models

import (
	"time"

	"github.com/google/uuid"
)

type AlbumMember struct {
	AlbumID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"albumId"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	Role      string    `gorm:"type:text;not null" json:"role"` // "admin" or "poster"
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
}
