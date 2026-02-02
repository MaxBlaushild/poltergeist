package models

import (
	"time"

	"github.com/google/uuid"
)

type UserDeviceToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	Token     string    `gorm:"type:text;not null" json:"token"`
	Platform  string    `gorm:"type:text;not null" json:"platform"` // ios, android
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
}
