package models

import (
	"github.com/google/uuid"
)

type PostTag struct {
	PostID uuid.UUID `gorm:"type:uuid;primaryKey" json:"postId"`
	Tag    string    `gorm:"type:text;primaryKey" json:"tag"`
}
