package models

import (
	"time"

	"github.com/google/uuid"
)

type CloudDocumentProvider string

const (
	CloudDocumentProviderUnknown CloudDocumentProvider = "unknown"
	CloudDocumentProviderGoogleDocs CloudDocumentProvider = "google_docs"
	CloudDocumentProviderGoogleSheets CloudDocumentProvider = "google_sheets"
	CloudDocumentProviderInternal CloudDocumentProvider = "internal"
)

type Document struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Title string `json:"title"`
	Provider CloudDocumentProvider `json:"provider"`
	UserID uuid.UUID `json:"userId"`
	User User `json:"user" gorm:"foreignKey:UserID"`
	DocumentTags []DocumentTag `json:"documentTags" gorm:"many2many:document_document_tags;"`
	Link *string `json:"link"`
	Content *string `json:"content"`
}