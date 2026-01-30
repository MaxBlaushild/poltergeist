package models

import (
	"time"

	"github.com/google/uuid"
)

type Album struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	Name      string    `gorm:"type:text;not null" json:"name"`
}

type AlbumTag struct {
	AlbumID uuid.UUID `gorm:"type:uuid;primaryKey" json:"albumId"`
	Tag     string    `gorm:"type:text;primaryKey" json:"tag"`
}
