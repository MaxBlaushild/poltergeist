package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointHandler struct {
	db *gorm.DB
}

func NewPointHandler(db *gorm.DB) *pointHandler {
	return &pointHandler{db: db}
}

func (h *pointHandler) CreatePoint(ctx context.Context, latitude float64, longitude float64) (*models.GeometryPoint, error) {
	// Check for existing points within 100 meters
	var existingPoint models.GeometryPoint
	err := h.db.Where("ST_DWithin(geometry, ST_SetSRID(ST_MakePoint(?, ?), 4326), 100)", longitude, latitude).First(&existingPoint).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err != gorm.ErrRecordNotFound {
		existingPoint.Latitude = latitude
		existingPoint.Longitude = longitude
		if err := h.db.Save(&existingPoint).Error; err != nil {
			return nil, err
		}
		return &existingPoint, nil
	}

	point := &models.GeometryPoint{
		Latitude:  latitude,
		Longitude: longitude,
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = h.db.Create(point).Error
	if err != nil {
		return nil, err
	}
	return point, nil
}
