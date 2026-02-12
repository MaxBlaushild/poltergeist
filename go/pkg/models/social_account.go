package models

import (
	"time"

	"github.com/google/uuid"
)

type SocialAccount struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	UserID       uuid.UUID  `gorm:"type:uuid;index;not null" json:"userId"`
	User         User       `json:"user" gorm:"foreignKey:UserID"`
	Provider     string     `gorm:"type:varchar(32);not null" json:"provider"`
	AccountID    *string    `gorm:"type:varchar(128)" json:"accountId,omitempty"`
	Username     *string    `gorm:"type:varchar(128)" json:"username,omitempty"`
	AccessToken  string     `gorm:"type:text;not null" json:"-"`
	RefreshToken string     `gorm:"type:text" json:"-"`
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
	Scopes       *string    `gorm:"type:text" json:"scopes,omitempty"`
}
