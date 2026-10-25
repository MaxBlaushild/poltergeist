// Package travellibrary owns the audience boundary shared by every document
// representation. Saves and trip references are personal bookmarks, never grants.
package travellibrary

import (
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TripLookup func(*gorm.DB, uuid.UUID, uuid.UUID) (map[string]interface{}, error)
type TripList func(*gorm.DB, uuid.UUID) ([]map[string]interface{}, error)
type Service struct {
	db    *gorm.DB
	auth  func(*gin.Context) (*models.User, error)
	trip  TripLookup
	trips TripList
}

func New(database *gorm.DB, auth func(*gin.Context) (*models.User, error), trip TripLookup, trips TripList) *Service {
	return &Service{db: database, auth: auth, trip: trip, trips: trips}
}

type share struct {
	DocumentID uuid.UUID `gorm:"primaryKey"`
	UserID     uuid.UUID `gorm:"primaryKey"`
	CreatedAt  time.Time
}

func (share) TableName() string { return "document_shares" }

type save struct {
	DocumentID uuid.UUID `gorm:"primaryKey"`
	UserID     uuid.UUID `gorm:"primaryKey"`
	CreatedAt  time.Time
}

func (save) TableName() string { return "document_saves" }

type tripReference struct {
	DocumentID uuid.UUID `gorm:"primaryKey"`
	UserID     uuid.UUID `gorm:"primaryKey"`
	TripID     uuid.UUID `gorm:"primaryKey"`
	CreatedAt  time.Time
}

func (tripReference) TableName() string { return "document_trip_references" }

// Readable always uses the documents table name, including in aggregate joins.
// Neither a saved reference nor a friend's own grants provide transitive access.
func Readable(tx *gorm.DB, user uuid.UUID) *gorm.DB {
	return tx.Where(`(documents.user_id = ? OR
	 (documents.visibility = 'network' AND EXISTS
	  (SELECT 1 FROM friends f WHERE (f.first_user_id = documents.user_id AND f.second_user_id = ?) OR (f.second_user_id = documents.user_id AND f.first_user_id = ?))) OR
	 (documents.visibility = 'selected' AND EXISTS
	  (SELECT 1 FROM document_shares ds WHERE ds.document_id = documents.id AND ds.user_id = ?)))`, user, user, user, user)
}
func documentQuery(tx *gorm.DB) *gorm.DB {
	return tx.Model(&models.Document{}).Preload("DocumentTags").Preload("DocumentLocations").Preload("User", func(q *gorm.DB) *gorm.DB {
		return q.Select("id", "name", "username", "profile_picture_url")
	})
}
func (s *Service) readableDocument(tx *gorm.DB, id, user uuid.UUID) (*models.Document, error) {
	var document models.Document
	if err := Readable(documentQuery(tx), user).First(&document, "documents.id = ?", id).Error; err != nil {
		return nil, err
	}
	return &document, nil
}
func (s *Service) dto(tx *gorm.DB, document *models.Document, user uuid.UUID) (gin.H, error) {
	var count int64
	if err := tx.Model(&save{}).Where("document_id = ? AND user_id = ?", document.ID, user).Count(&count).Error; err != nil {
		return nil, err
	}
	refs := []tripReference{}
	if err := tx.Where("document_id = ? AND user_id = ?", document.ID, user).Find(&refs).Error; err != nil {
		return nil, err
	}
	tripIDs := []uuid.UUID{}
	for _, ref := range refs {
		if s.trip == nil {
			continue
		}
		if _, err := s.trip(tx, user, ref.TripID); err == nil {
			tripIDs = append(tripIDs, ref.TripID)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	tags := document.DocumentTags
	if tags == nil {
		tags = []models.DocumentTag{}
	}
	locations := document.DocumentLocations
	if locations == nil {
		locations = []models.DocumentLocation{}
	}
	result := gin.H{
		"id": document.ID, "createdAt": document.CreatedAt, "updatedAt": document.UpdatedAt,
		"title": document.Title, "provider": document.Provider, "userId": document.UserID,
		"link": document.Link, "content": document.Content, "documentTags": tags, "documentLocations": locations,
		"visibility": document.Visibility, "format": document.Format, "personalNote": document.PersonalNote, "destination": document.Destination,
		"isSaved": count > 0, "savedTripIds": tripIDs,
		"user": gin.H{"id": document.UserID, "name": document.User.Name, "username": document.User.Username, "profilePictureUrl": document.User.ProfilePictureUrl},
	}
	if document.UserID == user {
		ids := []uuid.UUID{}
		if err := tx.Model(&share{}).Where("document_id = ?", document.ID).Order("user_id").Pluck("user_id", &ids).Error; err != nil {
			return nil, err
		}
		result["sharedWithUserIds"] = ids
	}
	return result, nil
}

func (s *Service) RespondDocument(c *gin.Context, documentID, userID uuid.UUID, status int) {
	c.Header("Cache-Control", "private, no-store")
	doc, err := s.readableDocument(s.db.WithContext(c), documentID, userID)
	if err != nil {
		respondError(c, err)
		return
	}
	result, err := s.dto(s.db.WithContext(c), doc, userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(status, result)
}
func (s *Service) ListScope(c *gin.Context, scope string) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	tx := s.db.WithContext(c)
	query := Readable(documentQuery(tx), u.ID)
	switch scope {
	case "yours":
		query = query.Where("documents.user_id = ?", u.ID)
	case "network":
		query = query.Where("documents.user_id <> ?", u.ID)
	case "saved":
		if rawTrip := c.Query("tripId"); rawTrip != "" {
			tripID, err := uuid.Parse(rawTrip)
			if err != nil || s.trip == nil {
				respondError(c, gorm.ErrRecordNotFound)
				return
			}
			if _, err := s.trip(tx, u.ID, tripID); err != nil {
				respondError(c, err)
				return
			}
			query = query.Where("EXISTS (SELECT 1 FROM document_trip_references tr WHERE tr.document_id = documents.id AND tr.user_id = ? AND tr.trip_id = ?)", u.ID, tripID)
		} else {
			query = query.Where("EXISTS (SELECT 1 FROM document_saves sv WHERE sv.document_id = documents.id AND sv.user_id = ?)", u.ID)
		}
	default:
		respondError(c, bad("Unknown library scope"))
		return
	}
	documents := []models.Document{}
	if err := query.Order("documents.created_at DESC, documents.id").Find(&documents).Error; err != nil {
		respondError(c, err)
		return
	}
	result := []gin.H{}
	for i := range documents {
		row, err := s.dto(tx, &documents[i], u.ID)
		if err != nil {
			respondError(c, err)
			return
		}
		result = append(result, row)
	}
	c.Header("Cache-Control", "private, no-store")
	c.JSON(200, result)
}

func (s *Service) Trending(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	result := gin.H{}
	for _, kind := range []string{"city", "country"} {
		rows := []models.TrendingDestination{}
		query := s.db.WithContext(c).Table("documents").Joins("JOIN document_locations dl ON dl.document_id = documents.id")
		err := Readable(query, u.ID).Where("dl.location_type = ?", kind).
			Select("dl.place_id, dl.name, dl.formatted_address, dl.latitude, dl.longitude, dl.location_type, COUNT(DISTINCT documents.id) AS document_count").
			Group("dl.place_id, dl.name, dl.formatted_address, dl.latitude, dl.longitude, dl.location_type").
			Order("document_count DESC, dl.name, dl.place_id").Limit(5).Scan(&rows).Error
		if err != nil {
			respondError(c, err)
			return
		}
		for i := range rows {
			rows[i].Rank = i + 1
		}
		key := "cities"
		if kind == "country" {
			key = "countries"
		}
		result[key] = rows
	}
	c.Header("Cache-Control", "private, no-store")
	c.JSON(200, result)
}
