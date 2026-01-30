package models

import (
	"time"

	"github.com/google/uuid"
)

type PostFlag struct {
	PostID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"postId"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
}
