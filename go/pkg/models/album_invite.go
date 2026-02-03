package models

import (
	"time"

	"github.com/google/uuid"
)

type AlbumInvite struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	AlbumID       uuid.UUID `gorm:"type:uuid;not null;index" json:"albumId"`
	InvitedUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"invitedUserId"`
	InvitedByID   uuid.UUID `gorm:"type:uuid;not null" json:"invitedById"`
	Role          string    `gorm:"type:text;not null;default:'poster'" json:"role"` // "admin" or "poster"
	Status        string     `gorm:"type:text;not null;default:'pending'" json:"status"`
	CreatedAt     time.Time  `gorm:"not null" json:"createdAt"`
	AcceptedAt    *time.Time `json:"acceptedAt,omitempty"`
}
