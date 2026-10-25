package models

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type CloudDocumentProvider string

const (
	CloudDocumentProviderUnknown      CloudDocumentProvider = "unknown"
	CloudDocumentProviderGoogleDocs   CloudDocumentProvider = "google_docs"
	CloudDocumentProviderGoogleSheets CloudDocumentProvider = "google_sheets"
	CloudDocumentProviderInternal     CloudDocumentProvider = "internal"
)

type Document struct {
	ID                uuid.UUID             `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt         time.Time             `json:"createdAt"`
	UpdatedAt         time.Time             `json:"updatedAt"`
	Title             string                `json:"title"`
	Provider          CloudDocumentProvider `json:"provider"`
	UserID            uuid.UUID             `json:"userId"`
	User              User                  `json:"user" gorm:"foreignKey:UserID"`
	DocumentTags      []DocumentTag         `json:"documentTags" gorm:"many2many:document_document_tags;"`
	DocumentLocations []DocumentLocation    `json:"documentLocations" gorm:"foreignKey:DocumentID"`
	Link              *string               `json:"link"`
	Content           *string               `json:"content"`
	Visibility        string                `json:"visibility" gorm:"not null;default:private"`
	Format            string                `json:"format" gorm:"not null;default:guide"`
	PersonalNote      string                `json:"personalNote"`
	Destination       string                `json:"destination"`
}

// NormalizeLibraryMetadata is shared by file, Drive and ordinary document
// creation so an omitted audience can never publish a new import to friends.
func (d *Document) NormalizeLibraryMetadata() error {
	if d.Visibility == "" {
		d.Visibility = "private"
	}
	if d.Visibility != "private" && d.Visibility != "network" && d.Visibility != "selected" {
		return fmt.Errorf("invalid document visibility")
	}
	if d.Format == "" {
		d.Format = "guide"
		if d.Provider == CloudDocumentProviderGoogleSheets {
			d.Format = "spreadsheet"
		}
		if d.Link != nil {
			if u, err := url.Parse(*d.Link); err == nil {
				path := strings.ToLower(u.Path)
				if strings.Contains(path, "/videos/") || strings.HasSuffix(path, ".mp4") || strings.HasSuffix(path, ".mov") || strings.HasSuffix(path, ".m4v") {
					d.Format = "video"
				}
			}
		}
	}
	switch d.Format {
	case "spreadsheet", "itinerary", "guide", "video":
	default:
		return fmt.Errorf("invalid document format")
	}
	d.Title = strings.TrimSpace(d.Title)
	d.PersonalNote = strings.TrimSpace(d.PersonalNote)
	d.Destination = strings.TrimSpace(d.Destination)
	if d.Title == "" || utf8.RuneCountInString(d.Title) > 500 {
		return fmt.Errorf("title must contain 1 to 500 characters")
	}
	if utf8.RuneCountInString(d.PersonalNote) > 5000 || utf8.RuneCountInString(d.Destination) > 300 {
		return fmt.Errorf("document note or destination is too long")
	}
	return nil
}

func ValidateDocumentSourceLink(link *string) error {
	if link == nil || *link == "" {
		return nil
	}
	u, err := url.Parse(*link)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		return fmt.Errorf("document link must be an HTTP or HTTPS URL")
	}
	return nil
}
