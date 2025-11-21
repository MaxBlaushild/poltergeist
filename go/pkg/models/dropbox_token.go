package models

import (
	"time"

	"github.com/google/uuid"
)

type DropboxToken struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	UserID       uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"userId"`
	User         User      `json:"user" gorm:"foreignKey:UserID"`
	AccessToken  string    `gorm:"type:text;not null" json:"-"` // Don't expose in JSON
	RefreshToken string    `gorm:"type:text;not null" json:"-"` // Don't expose in JSON
	ExpiresAt    time.Time `gorm:"not null" json:"expiresAt"`
	TokenType    string    `gorm:"type:varchar(50);default:'Bearer'" json:"tokenType"`
}

