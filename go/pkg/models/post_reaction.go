package models

import (
	"time"

	"github.com/google/uuid"
)

type PostReaction struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"not null" json:"updatedAt"`
	PostID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"postId"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	Emoji     string     `gorm:"type:text;not null;index" json:"emoji"`
}
