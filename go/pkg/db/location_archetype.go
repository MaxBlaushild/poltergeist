package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type locationArchetypeHandle struct {
	db *gorm.DB
}

func (h *locationArchetypeHandle) Create(ctx context.Context, locationArchetype *models.LocationArchetype) error {
	return h.db.WithContext(ctx).Create(locationArchetype).Error
}

func (h *locationArchetypeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.LocationArchetype, error) {
	var locationArchetype models.LocationArchetype
	if err := h.db.WithContext(ctx).First(&locationArchetype, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &locationArchetype, nil
}

func (h *locationArchetypeHandle) FindAll(ctx context.Context) ([]*models.LocationArchetype, error) {
	var locationArchetypes []*models.LocationArchetype
	if err := h.db.WithContext(ctx).Find(&locationArchetypes).Error; err != nil {
		return nil, err
	}
	return locationArchetypes, nil
}

func (h *locationArchetypeHandle) Update(ctx context.Context, locationArchetype *models.LocationArchetype) error {
	return h.db.WithContext(ctx).Save(locationArchetype).Error
}

func (h *locationArchetypeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.LocationArchetype{}, "id = ?", id).Error
}
