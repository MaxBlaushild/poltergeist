package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type neighboringPointsOfInterestHandle struct {
	db *gorm.DB
}

func (h *neighboringPointsOfInterestHandle) Create(ctx context.Context, pointOfInterestOneID uuid.UUID, pointOfInterestTwoID uuid.UUID) error {
	return h.db.WithContext(ctx).Create(&models.NeighboringPointsOfInterest{
		PointOfInterestOneID: pointOfInterestOneID,
		PointOfInterestTwoID: pointOfInterestTwoID,
	}).Error
}

func (h *neighboringPointsOfInterestHandle) FindAll(ctx context.Context) ([]models.NeighboringPointsOfInterest, error) {
	var neighbors []models.NeighboringPointsOfInterest

	if err := h.db.WithContext(ctx).Find(&neighbors).Error; err != nil {
		return nil, err
	}

	return neighbors, nil
}
