package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	Type      string     `gorm:"type:text;not null" json:"type"` // album_invite, album_invite_accepted, album_photo_added
	ActorID   uuid.UUID  `gorm:"type:uuid;not null" json:"actorId"`
	AlbumID   uuid.UUID  `gorm:"type:uuid;not null" json:"albumId"`
	PostID    *uuid.UUID `gorm:"type:uuid" json:"postId,omitempty"`
	InviteID  *uuid.UUID `gorm:"type:uuid" json:"inviteId,omitempty"`
	ReadAt    *time.Time `json:"readAt,omitempty"`
	CreatedAt time.Time  `gorm:"not null" json:"createdAt"`
}
