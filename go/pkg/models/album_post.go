package models

import (
	"time"

	"github.com/google/uuid"
)

type AlbumPost struct {
	AlbumID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"albumId"`
	PostID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"postId"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
}
