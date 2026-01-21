package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	ImageURL  string    `gorm:"type:text;not null" json:"imageUrl"`
	Caption   *string   `gorm:"type:text" json:"caption,omitempty"`
}

