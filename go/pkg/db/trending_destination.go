package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type trendingDestinationHandler struct {
	db *gorm.DB
}

func (h *trendingDestinationHandler) Upsert(ctx context.Context, destination *models.TrendingDestination) error {
	// Set ID and timestamps if not set
	if destination.ID == uuid.Nil {
		destination.ID = uuid.New()
	}
	if destination.CreatedAt.IsZero() {
		destination.CreatedAt = time.Now()
	}
	destination.UpdatedAt = time.Now()

	// Since we clear all entries before inserting, we can just use Create
	// But we'll use Save which will update if it exists (based on unique constraint)
	// or create if it doesn't
	return h.db.WithContext(ctx).Save(destination).Error
}

func (h *trendingDestinationHandler) FindByType(ctx context.Context, locationType models.LocationType) ([]models.TrendingDestination, error) {
	var destinations []models.TrendingDestination
	if err := h.db.WithContext(ctx).
		Where("location_type = ?", locationType).
		Order("rank ASC").
		Find(&destinations).Error; err != nil {
		return nil, err
	}
	return destinations, nil
}

func (h *trendingDestinationHandler) DeleteAll(ctx context.Context) error {
	return h.db.WithContext(ctx).
		Delete(&models.TrendingDestination{}).Error
}

type LocationCountResult struct {
	PlaceID          string
	Name             string
	FormattedAddress string
	Latitude         float64
	Longitude        float64
	LocationType     string
	DocumentCount    int
}

func (h *trendingDestinationHandler) GetTopLocationsByType(ctx context.Context, locationType models.LocationType, since time.Time, limit int) ([]LocationCountResult, error) {
	var results []LocationCountResult
	query := `
		SELECT 
			dl.place_id,
			dl.name,
			dl.formatted_address,
			dl.latitude,
			dl.longitude,
			dl.location_type,
			COUNT(DISTINCT d.id) as document_count
		FROM document_locations dl
		INNER JOIN documents d ON dl.document_id = d.id
		WHERE dl.location_type = ?
			AND d.created_at >= ?
		GROUP BY dl.place_id, dl.name, dl.formatted_address, dl.latitude, dl.longitude, dl.location_type
		ORDER BY document_count DESC, MAX(d.created_at) DESC
		LIMIT ?
	`

	if err := h.db.WithContext(ctx).Raw(query, string(locationType), since, limit).Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
