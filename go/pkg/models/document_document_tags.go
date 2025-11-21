package models

import (
	"time"

	"github.com/google/uuid"
)

type DocumentDocumentTag struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	DocumentID uuid.UUID `json:"documentId"`
	Document   Document  `json:"document" gorm:"foreignKey:DocumentID"`
	TagID      uuid.UUID `json:"tagId"`
	Tag        Tag       `json:"tag" gorm:"foreignKey:TagID"`
}
