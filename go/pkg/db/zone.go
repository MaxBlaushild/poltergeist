package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type zoneHandler struct {
	db *gorm.DB
}

func (h *zoneHandler) Create(ctx context.Context, zone *models.Zone) error {
	if zone.Boundary == "" {
		zone.Boundary = "010300000000000000" // Empty polygon in WKB hex format
	}

	return h.db.WithContext(ctx).Create(zone).Error
}

func (h *zoneHandler) FindAll(ctx context.Context) ([]*models.Zone, error) {
	var zones []*models.Zone
	if err := h.db.WithContext(ctx).Preload("Points").Find(&zones).Error; err != nil {
		return nil, err
	}
	return zones, nil
}

func (h *zoneHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.Zone, error) {
	var zone models.Zone
	if err := h.db.WithContext(ctx).Preload("Points").Where("id = ?", id).First(&zone).Error; err != nil {
		return nil, err
	}
	return &zone, nil
}

func (h *zoneHandler) Update(ctx context.Context, zone *models.Zone) error {
	return h.db.WithContext(ctx).Save(zone).Error
}

func (h *zoneHandler) Delete(ctx context.Context, zoneID uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Zone{}, "id = ?", zoneID).Error
}

func (h *zoneHandler) AddPointOfInterestToZone(ctx context.Context, zoneID uuid.UUID, pointOfInterestID uuid.UUID) error {
	return h.db.WithContext(ctx).Create(&models.PointOfInterestZone{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		ZoneID:            zoneID,
		PointOfInterestID: pointOfInterestID,
	}).Error
}

func (h *zoneHandler) RemovePointOfInterestFromZone(ctx context.Context, zoneID uuid.UUID, pointOfInterestID uuid.UUID) error {
	return h.db.WithContext(ctx).
		Where("zone_id = ? AND point_of_interest_id = ?", zoneID, pointOfInterestID).
		Delete(&models.PointOfInterestZone{}).Error
}

func (h *zoneHandler) UpdateBoundary(ctx context.Context, zoneID uuid.UUID, boundary [][]float64) error {
	// Create a PostGIS Polygon WKT string
	// Format: POLYGON((lng1 lat1, lng2 lat2, ...))
	var coords []string
	for _, point := range boundary {
		// Format coordinates with 6 decimal places and ensure proper spacing
		coords = append(coords, fmt.Sprintf("%.6f %.6f", point[0], point[1]))
	}
	// Close the polygon by repeating the first point
	if len(boundary) > 0 {
		coords = append(coords, fmt.Sprintf("%.6f %.6f", boundary[0][0], boundary[0][1]))
	}

	// Format the WKT string with proper spacing and parentheses
	wkt := fmt.Sprintf("POLYGON((%s))", strings.Join(coords, ", "))

	// Use ST_GeomFromText to properly parse the WKT
	updates := map[string]interface{}{
		"boundary": gorm.Expr("ST_GeomFromText(?, 4326)", wkt),
	}

	if err := h.db.WithContext(ctx).Model(&models.Zone{}).Where("id = ?", zoneID).Updates(updates).Error; err != nil {
		return err
	}

	pointHandle := pointHandler{db: h.db}

	if err := h.db.WithContext(ctx).Where("zone_id = ?", zoneID).Delete(&models.BoundaryPoint{}).Error; err != nil {
		return err
	}

	for _, point := range boundary {
		point, err := pointHandle.CreatePoint(ctx, point[0], point[1])
		if err != nil {
			return err
		}

		if err := h.db.WithContext(ctx).Create(&models.BoundaryPoint{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ZoneID:    zoneID,
			PointID:   point.ID,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (h *zoneHandler) FindByPointOfInterestID(ctx context.Context, pointOfInterestID uuid.UUID) (*models.Zone, error) {
	var pointOfInterestZone models.PointOfInterestZone
	if err := h.db.WithContext(ctx).Where("point_of_interest_id = ?", pointOfInterestID).First(&pointOfInterestZone).Error; err != nil {
		return nil, err
	}

	var zone models.Zone
	if err := h.db.WithContext(ctx).Where("id = ?", pointOfInterestZone.ZoneID).First(&zone).Error; err != nil {
		return nil, err
	}

	return &zone, nil
}

func (h *zoneHandler) UpdateNameAndDescription(ctx context.Context, zoneID uuid.UUID, name string, description string) error {
	return h.db.WithContext(ctx).Model(&models.Zone{}).Where("id = ?", zoneID).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
	}).Error
}
