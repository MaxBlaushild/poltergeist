package travelcalendar

import (
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxTripDocuments = 100
const maxTripGuideBytes = 1 << 20

// These shared copies must never reuse document_trip_references: that table
// contains personal bookmarks and is expressly not a sharing grant.
type TripDocument struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	StopID        uuid.UUID `gorm:"type:uuid"`
	AddedBy       uuid.UUID `gorm:"type:uuid"`
	Title         string
	Kind          string
	Format        string
	Link          string
	Content       string
	Filename      string
	ContentType   string
	SizeBytes     int64
	StorageBucket string
	StorageKey    string
	CreatedAt     time.Time
}

func (TripDocument) TableName() string { return "tc_trip_documents" }

func (s *Service) registerTripDocumentRoutes(g *gin.RouterGroup) {
	g.GET("/calendar/stops/:id/documents", s.listTripDocuments)
	g.POST("/calendar/stops/:id/documents", s.createTripDocument)
	g.POST("/calendar/stops/:id/documents/upload", s.uploadTripDocument)
	g.GET("/calendar/stops/:id/documents/:documentId", s.getTripDocument)
	g.DELETE("/calendar/stops/:id/documents/:documentId", s.deleteTripDocument)
	g.GET("/calendar/stops/:id/documents/:documentId/file", s.getTripDocumentFile)
	g.GET("/calendar/stops/:id/documents/:documentId/comments", s.listTripDocumentComments)
	g.POST("/calendar/stops/:id/documents/:documentId/comments", s.createTripDocumentComment)
	g.PUT("/calendar/stops/:id/documents/:documentId/comments/:commentId", s.updateTripDocumentComment)
	g.DELETE("/calendar/stops/:id/documents/:documentId/comments/:commentId", s.deleteTripDocumentComment)
}

func (s *Service) tripDocumentCapabilities(tx *gorm.DB, stop *Stop, viewer Viewer) (bool, bool) {
	// An unrelated verified guest session must not lend its grant to a claimed
	// account. Account-linked recipient grants are resolved by hasGrant itself.
	viewer.RecipientID = uuid.Nil
	canView := viewer.UserID != uuid.Nil && s.canViewStop(tx, stop, viewer) &&
		(s.canManageStop(tx, stop, viewer) || s.tripAttendee(tx, stop, viewer))
	content := stop.Published
	if content == nil {
		content = &stop.Draft
	}
	return canView, canView && content.Status != "cancelled"
}

// Every document representation uses the trip's current permission lock order.
// A remembered RSVP is not a grant after a link is rotated or access is revoked.
func (s *Service) tripDocumentAccess(tx *gorm.DB, c *gin.Context, viewer Viewer, contribute bool) (*Stop, Viewer, error) {
	stop, err := s.lockedStop(tx, c.Param("id"))
	if err != nil {
		return nil, viewer, err
	}
	viewer = s.currentTripMediaViewer(tx, stop, c, viewer)
	canView, canContribute := s.tripDocumentCapabilities(tx, stop, viewer)
	if !canView {
		return nil, viewer, errUnavailable
	}
	if contribute && !canContribute {
		return nil, viewer, errForbidden
	}
	return stop, viewer, nil
}

func loadTripDocument(tx *gorm.DB, stopID uuid.UUID, rawID string) (*TripDocument, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, errUnavailable
	}
	var document TripDocument
	err = tx.Where("id = ? AND stop_id = ?", id, stopID).First(&document).Error
	return &document, err
}

func (s *Service) tripDocumentIdentity(tx *gorm.DB, id uuid.UUID) gin.H {
	identity := s.tripSafeIdentity(tx, id)
	identity["id"] = identity["userId"]
	delete(identity, "userId")
	return identity
}

func (s *Service) tripDocumentDTO(tx *gorm.DB, stop *Stop, document *TripDocument, viewer Viewer, detail bool) (gin.H, error) {
	var comments int64
	if err := tx.Model(&TripDocumentComment{}).Where("document_id = ?", document.ID).Count(&comments).Error; err != nil {
		return nil, err
	}
	_, canContribute := s.tripDocumentCapabilities(tx, stop, viewer)
	dto := gin.H{
		"id": document.ID, "title": document.Title, "kind": document.Kind, "format": document.Format,
		"createdAt": document.CreatedAt, "addedBy": s.tripDocumentIdentity(tx, document.AddedBy),
		"canDelete":     viewer.UserID == document.AddedBy || s.canManageStop(tx, stop, viewer),
		"canContribute": canContribute, "commentCount": comments,
	}
	if document.Link != "" {
		dto["link"] = document.Link
	}
	if detail && document.Content != "" {
		dto["content"] = document.Content
	}
	if document.Kind == "file" {
		dto["filename"], dto["contentType"], dto["sizeBytes"] = document.Filename, document.ContentType, document.SizeBytes
		dto["fileUrl"] = "/travel-angels/calendar/stops/" + stop.ID.String() + "/documents/" + document.ID.String() + "/file"
	}
	return dto, nil
}

func (s *Service) listTripDocuments(c *gin.Context) {
	viewer := s.viewer(c)
	var result gin.H
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, err := s.lockedStop(tx, c.Param("id"))
		if err != nil {
			return err
		}
		viewer = s.currentTripMediaViewer(tx, stop, c, viewer)
		if !s.canViewStop(tx, stop, viewer) {
			return errUnavailable
		}
		canView, canContribute := s.tripDocumentCapabilities(tx, stop, viewer)
		rows := []gin.H{}
		result = gin.H{"documents": rows, "canView": canView, "canContribute": canContribute}
		if !canView {
			return nil
		}
		var documents []TripDocument
		if err = tx.Where("stop_id = ?", stop.ID).Omit("content").Order("created_at DESC, id DESC").Limit(maxTripDocuments).Find(&documents).Error; err != nil {
			return err
		}
		for i := range documents {
			dto, err := s.tripDocumentDTO(tx, stop, &documents[i], viewer, false)
			if err != nil {
				return err
			}
			rows = append(rows, dto)
		}
		result["documents"] = rows
		return nil
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, result)
}

func (s *Service) getTripDocument(c *gin.Context) {
	viewer := s.viewer(c)
	var result gin.H
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, current, err := s.tripDocumentAccess(tx, c, viewer, false)
		if err != nil {
			return err
		}
		document, err := loadTripDocument(tx, stop.ID, c.Param("documentId"))
		if err != nil {
			return err
		}
		result, err = s.tripDocumentDTO(tx, stop, document, current, true)
		return err
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"document": result})
}

func validateTripDocumentTitle(title string) error {
	if !utf8.ValidString(title) || title == "" || utf8.RuneCountInString(title) > 500 || strings.ContainsRune(title, 0) {
		return bad("Add a document title up to 500 characters")
	}
	return nil
}

func validateTripDocumentLink(link string) error {
	if link == "" || len(link) > 8192 || !utf8.ValidString(link) || strings.ContainsAny(link, "\r\n\x00") || models.ValidateDocumentSourceLink(&link) != nil {
		return bad("Use an HTTP or HTTPS document link without account credentials")
	}
	return nil
}

func checkTripDocumentCapacity(tx *gorm.DB, stopID uuid.UUID) error {
	var count int64
	if err := tx.Model(&TripDocument{}).Where("stop_id = ?", stopID).Count(&count).Error; err != nil {
		return err
	}
	if count >= maxTripDocuments {
		return bad("This trip has 100 documents. Remove a document before adding another")
	}
	return nil
}

func (s *Service) createTripDocument(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var input struct {
		Title            string `json:"title"`
		Link             string `json:"link"`
		SourceDocumentID string `json:"sourceDocumentId"`
	}
	if c.ShouldBindJSON(&input) != nil || (input.SourceDocumentID == "") == (strings.TrimSpace(input.Link) == "") {
		respondError(c, bad("Choose a document link or one of your library guides"))
		return
	}
	viewer := s.viewer(c)
	viewer.UserID = user.ID
	document := TripDocument{ID: uuid.New(), AddedBy: user.ID, Title: strings.TrimSpace(input.Title), Kind: "link", Format: "guide", Link: strings.TrimSpace(input.Link), CreatedAt: time.Now().UTC()}
	var result gin.H
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, current, err := s.tripDocumentAccess(tx, c, viewer, true)
		if err != nil {
			return err
		}
		document.StopID = stop.ID
		if err = checkTripDocumentCapacity(tx, stop.ID); err != nil {
			return err
		}
		if input.SourceDocumentID != "" {
			id, err := uuid.Parse(input.SourceDocumentID)
			if err != nil {
				return errUnavailable
			}
			// Lock and recheck ownership. Being allowed to read another person's
			// guide never permits republishing it to a different audience.
			var source models.Document
			if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id, user_id, title, link, content, format").Where("id = ? AND user_id = ?", id, user.ID).First(&source).Error; err != nil {
				return err
			}
			document.Kind = "guide"
			if document.Title == "" {
				document.Title = source.Title
			}
			if source.Content != nil {
				document.Content = *source.Content
			}
			if source.Link != nil {
				document.Link = strings.TrimSpace(*source.Link)
			}
			if len(document.Content) > maxTripGuideBytes || !utf8.ValidString(document.Content) || strings.ContainsRune(document.Content, 0) {
				return bad("This guide is too large to share. Add its source link instead")
			}
			if document.Content == "" && document.Link == "" {
				return bad("This guide has no content or source link to share")
			}
			switch source.Format {
			case "guide", "itinerary", "spreadsheet", "video":
				document.Format = source.Format
			}
		}
		if err = validateTripDocumentTitle(document.Title); err != nil {
			return err
		}
		if document.Link != "" {
			if err = validateTripDocumentLink(document.Link); err != nil {
				return err
			}
		}
		if err = tx.Create(&document).Error; err != nil {
			return err
		}
		result, err = s.tripDocumentDTO(tx, stop, &document, current, true)
		return err
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(201, gin.H{"document": result})
}

func (s *Service) deleteTripDocument(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	viewer := s.viewer(c)
	viewer.UserID = user.ID
	var deleted *TripDocument
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, current, err := s.tripDocumentAccess(tx, c, viewer, false)
		if err != nil {
			return err
		}
		document, err := loadTripDocument(tx, stop.ID, c.Param("documentId"))
		if err != nil {
			return err
		}
		if document.AddedBy != user.ID && !s.canManageStop(tx, stop, current) {
			return errForbidden
		}
		if err = tx.Delete(document).Error; err != nil {
			return err
		}
		deleted = document
		return nil
	})
	if err != nil {
		respondError(c, err)
		return
	}
	s.deleteTripDocumentObject(deleted)
	c.JSON(200, gin.H{"deleted": true})
}
