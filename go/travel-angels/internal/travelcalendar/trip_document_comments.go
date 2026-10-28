package travelcalendar

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TripDocumentComment struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	StopID     uuid.UUID `gorm:"type:uuid"`
	DocumentID uuid.UUID `gorm:"type:uuid"`
	AuthorID   uuid.UUID `gorm:"type:uuid"`
	Body       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (TripDocumentComment) TableName() string { return "tc_trip_document_comments" }

func (s *Service) tripDocumentCommentDTO(tx *gorm.DB, stop *Stop, comment *TripDocumentComment, viewer Viewer) gin.H {
	_, canContribute := s.tripDocumentCapabilities(tx, stop, viewer)
	return gin.H{
		"id": comment.ID, "body": comment.Body,
		"author":    s.tripDocumentIdentity(tx, comment.AuthorID),
		"createdAt": comment.CreatedAt, "updatedAt": comment.UpdatedAt,
		"canEdit":   canContribute && viewer.UserID == comment.AuthorID,
		"canDelete": viewer.UserID == comment.AuthorID || s.canManageStop(tx, stop, viewer),
	}
}

func (s *Service) listTripDocumentComments(c *gin.Context) {
	viewer := s.viewer(c)
	limit := 100
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100 {
			respondError(c, bad("Choose a comment page size between 1 and 100"))
			return
		}
		limit = parsed
	}
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
		query := tx.Where("stop_id = ? AND document_id = ?", stop.ID, document.ID)
		if before := c.Query("before"); before != "" {
			id, err := uuid.Parse(before)
			if err != nil {
				return bad("Invalid comment page")
			}
			var cursor TripDocumentComment
			if err = tx.Where("id = ? AND stop_id = ? AND document_id = ?", id, stop.ID, document.ID).First(&cursor).Error; err != nil {
				return errUnavailable
			}
			query = query.Where("(created_at, id) < (?, ?)", cursor.CreatedAt, cursor.ID)
		}
		var comments []TripDocumentComment
		if err = query.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&comments).Error; err != nil {
			return err
		}
		hasMore := len(comments) > limit
		if hasMore {
			comments = comments[:limit]
		}
		rows := []gin.H{}
		for i := len(comments) - 1; i >= 0; i-- {
			rows = append(rows, s.tripDocumentCommentDTO(tx, stop, &comments[i], current))
		}
		_, canContribute := s.tripDocumentCapabilities(tx, stop, current)
		result = gin.H{"comments": rows, "hasMore": hasMore, "canContribute": canContribute}
		if hasMore {
			result["nextBefore"] = comments[len(comments)-1].ID
		}
		return nil
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, result)
}

func tripDocumentCommentBody(c *gin.Context) (string, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<10)
	var input struct {
		Body string `json:"body"`
	}
	if c.ShouldBindJSON(&input) != nil {
		return "", bad("Add a comment up to 4,000 characters")
	}
	input.Body = strings.TrimSpace(input.Body)
	if input.Body == "" || !utf8.ValidString(input.Body) || utf8.RuneCountInString(input.Body) > 4000 || strings.ContainsRune(input.Body, 0) {
		return "", bad("Add a comment up to 4,000 characters")
	}
	return input.Body, nil
}

func (s *Service) createTripDocumentComment(c *gin.Context) {
	s.mutateTripDocumentComment(c, "create")
}
func (s *Service) updateTripDocumentComment(c *gin.Context) {
	s.mutateTripDocumentComment(c, "update")
}
func (s *Service) deleteTripDocumentComment(c *gin.Context) {
	s.mutateTripDocumentComment(c, "delete")
}

func (s *Service) mutateTripDocumentComment(c *gin.Context, action string) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	body := ""
	if action != "delete" {
		var err error
		body, err = tripDocumentCommentBody(c)
		if err != nil {
			respondError(c, err)
			return
		}
	}
	viewer := s.viewer(c)
	viewer.UserID = user.ID
	var result gin.H
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, current, err := s.tripDocumentAccess(tx, c, viewer, action != "delete")
		if err != nil {
			return err
		}
		document, err := loadTripDocument(tx, stop.ID, c.Param("documentId"))
		if err != nil {
			return err
		}
		comment := TripDocumentComment{ID: uuid.New(), StopID: stop.ID, DocumentID: document.ID, AuthorID: user.ID, Body: body, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		if action == "create" {
			var count int64
			if err = tx.Model(&TripDocumentComment{}).Where("document_id = ?", document.ID).Count(&count).Error; err != nil {
				return err
			}
			if count >= 1000 {
				return bad("This document has reached its 1,000 comment limit")
			}
			if err = tx.Create(&comment).Error; err != nil {
				return err
			}
		} else {
			id, err := uuid.Parse(c.Param("commentId"))
			if err != nil {
				return errUnavailable
			}
			comment = TripDocumentComment{}
			if err = tx.Where("id = ? AND stop_id = ? AND document_id = ?", id, stop.ID, document.ID).First(&comment).Error; err != nil {
				return err
			}
			if comment.AuthorID != user.ID && (action != "delete" || !s.canManageStop(tx, stop, current)) {
				return errForbidden
			}
			if action == "delete" {
				return tx.Delete(&comment).Error
			}
			comment.Body, comment.UpdatedAt = body, time.Now().UTC()
			if err = tx.Save(&comment).Error; err != nil {
				return err
			}
		}
		result = s.tripDocumentCommentDTO(tx, stop, &comment, current)
		return nil
	})
	if err != nil {
		respondError(c, err)
		return
	}
	if action == "delete" {
		c.JSON(200, gin.H{"deleted": true})
		return
	}
	status := http.StatusOK
	if action == "create" {
		status = http.StatusCreated
	}
	c.JSON(status, gin.H{"comment": result})
}
