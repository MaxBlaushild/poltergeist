package travellibrary

import (
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type inputError string

func (e inputError) Error() string { return string(e) }
func bad(message string) error     { return inputError(message) }

var forbidden = errors.New("forbidden")

func respondError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "This document or trip is unavailable"})
		return
	}
	if errors.Is(err, forbidden) {
		c.JSON(403, gin.H{"error": "Only the owner can change document sharing"})
		return
	}
	var invalid inputError
	if errors.As(err, &invalid) {
		c.JSON(400, gin.H{"error": invalid.Error()})
		return
	}
	c.JSON(500, gin.H{"error": "Unable to complete this library action"})
}
func (s *Service) requireUser(c *gin.Context) (*models.User, bool) {
	u, err := s.auth(c)
	if err != nil || u == nil || u.ID == uuid.Nil {
		c.JSON(401, gin.H{"error": "Sign in to use the library"})
		return nil, false
	}
	return u, true
}
func documentID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, gorm.ErrRecordNotFound)
		return uuid.Nil, false
	}
	return id, true
}
func (s *Service) RegisterRoutes(router *gin.Engine) {
	g := router.Group("/travel-angels/documents")
	g.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-store")
		c.Header("X-Robots-Tag", "noindex, nofollow")
		c.Next()
	})
	g.GET("/library", func(c *gin.Context) { s.ListScope(c, c.DefaultQuery("scope", "network")) })
	g.GET("/:id", s.get)
	g.PUT("/:id/sharing", s.setSharing)
	g.PUT("/:id/saved", s.setSaved)
	g.GET("/:id/trips", s.getTrips)
	g.PUT("/:id/trips/:tripId", s.setTrip)
}
func (s *Service) get(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	id, ok := documentID(c)
	if !ok {
		return
	}
	s.RespondDocument(c, id, u.ID, 200)
}
func (s *Service) lockedDocument(tx *gorm.DB, id, user uuid.UUID) (*models.Document, error) {
	var doc models.Document
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&doc, "id = ?", id).Error; err != nil {
		return nil, err
	}
	// Check again after obtaining the lock so a save cannot race revocation.
	return s.readableDocument(tx, id, user)
}
func (s *Service) setSharing(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	id, ok := documentID(c)
	if !ok {
		return
	}
	var body struct {
		Visibility string      `json:"visibility"`
		UserIDs    []uuid.UUID `json:"userIds"`
	}
	if c.ShouldBindJSON(&body) != nil || len(body.UserIDs) > 100 {
		respondError(c, bad("Choose a valid audience and at most 100 people"))
		return
	}
	if body.Visibility != "private" && body.Visibility != "network" && body.Visibility != "selected" {
		respondError(c, bad("Choose private, network or selected"))
		return
	}
	err := s.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		doc, err := s.lockedDocument(tx, id, u.ID)
		if err != nil {
			return err
		}
		if doc.UserID != u.ID {
			return forbidden
		}
		ids := map[uuid.UUID]bool{}
		if body.Visibility == "selected" {
			for _, person := range body.UserIDs {
				if person == uuid.Nil {
					return bad("Choose valid accounts")
				}
				if person != u.ID {
					ids[person] = true
				}
			}
			for person := range ids {
				var count int64
				if err := tx.Model(&models.User{}).Where("id = ?", person).Count(&count).Error; err != nil {
					return err
				}
				if count != 1 {
					return bad("One or more selected accounts is unavailable")
				}
			}
		}
		if err := tx.Where("document_id = ?", id).Delete(&share{}).Error; err != nil {
			return err
		}
		for person := range ids {
			if err := tx.Create(&share{DocumentID: id, UserID: person, CreatedAt: time.Now()}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&models.Document{}).Where("id = ?", id).Updates(map[string]interface{}{"visibility": body.Visibility, "updated_at": time.Now()}).Error
	})
	if err != nil {
		respondError(c, err)
		return
	}
	s.RespondDocument(c, id, u.ID, 200)
}
func (s *Service) setSaved(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	id, ok := documentID(c)
	if !ok {
		return
	}
	var body struct {
		Saved *bool `json:"saved"`
	}
	if c.ShouldBindJSON(&body) != nil || body.Saved == nil {
		respondError(c, bad("saved is required"))
		return
	}
	err := s.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if _, err := s.lockedDocument(tx, id, u.ID); err != nil {
			return err
		}
		if *body.Saved {
			return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&save{DocumentID: id, UserID: u.ID, CreatedAt: time.Now()}).Error
		}
		return tx.Where("document_id = ? AND user_id = ?", id, u.ID).Delete(&save{}).Error
	})
	if err != nil {
		respondError(c, err)
		return
	}
	s.RespondDocument(c, id, u.ID, 200)
}
func (s *Service) getTrips(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	id, ok := documentID(c)
	if !ok {
		return
	}
	tx := s.db.WithContext(c)
	if _, err := s.readableDocument(tx, id, u.ID); err != nil {
		respondError(c, err)
		return
	}
	rows := []map[string]interface{}{}
	if s.trips != nil {
		var err error
		rows, err = s.trips(tx, u.ID)
		if err != nil {
			respondError(c, err)
			return
		}
	}
	refs := []tripReference{}
	if err := tx.Where("document_id = ? AND user_id = ?", id, u.ID).Find(&refs).Error; err != nil {
		respondError(c, err)
		return
	}
	saved := map[string]bool{}
	for _, ref := range refs {
		saved[ref.TripID.String()] = true
	}
	for _, row := range rows {
		row["isSaved"] = saved[row["id"].(string)]
	}
	c.JSON(200, rows)
}
func (s *Service) setTrip(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	id, ok := documentID(c)
	if !ok {
		return
	}
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil || s.trip == nil {
		respondError(c, gorm.ErrRecordNotFound)
		return
	}
	var body struct {
		Saved *bool `json:"saved"`
	}
	if c.ShouldBindJSON(&body) != nil || body.Saved == nil {
		respondError(c, bad("saved is required"))
		return
	}
	err = s.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if _, err := s.lockedDocument(tx, id, u.ID); err != nil {
			return err
		}
		if _, err := s.trip(tx, u.ID, tripID); err != nil {
			return err
		}
		if !*body.Saved {
			return tx.Where("document_id = ? AND user_id = ? AND trip_id = ?", id, u.ID, tripID).Delete(&tripReference{}).Error
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&save{DocumentID: id, UserID: u.ID, CreatedAt: time.Now()}).Error; err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&tripReference{DocumentID: id, UserID: u.ID, TripID: tripID, CreatedAt: time.Now()}).Error
	})
	if err != nil {
		respondError(c, err)
		return
	}
	s.RespondDocument(c, id, u.ID, 200)
}
