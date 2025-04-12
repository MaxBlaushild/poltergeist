package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type zoneHandler struct {
	db *gorm.DB
}

func (h *zoneHandler) Create(ctx context.Context, zone *models.Zone) error {
	return h.db.WithContext(ctx).Create(zone).Error
}

func (h *zoneHandler) FindAll(ctx context.Context) ([]*models.Zone, error) {
	var zones []*models.Zone
	if err := h.db.WithContext(ctx).Find(&zones).Error; err != nil {
		return nil, err
	}
	return zones, nil
}

func (h *zoneHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.Zone, error) {
	var zone models.Zone
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&zone).Error; err != nil {
		return nil, err
	}
	return &zone, nil
}

func (h *zoneHandler) Update(ctx context.Context, zone *models.Zone) error {
	return h.db.WithContext(ctx).Save(zone).Error
}
