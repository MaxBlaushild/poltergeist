package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type zoneKindHandle struct {
	db *gorm.DB
}

func (h *zoneKindHandle) Create(ctx context.Context, zoneKind *models.ZoneKind) error {
	return h.db.WithContext(ctx).Create(zoneKind).Error
}

func (h *zoneKindHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ZoneKind, error) {
	var zoneKind models.ZoneKind
	if err := h.db.WithContext(ctx).First(&zoneKind, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &zoneKind, nil
}

func (h *zoneKindHandle) FindBySlug(ctx context.Context, slug string) (*models.ZoneKind, error) {
	var zoneKind models.ZoneKind
	if err := h.db.WithContext(ctx).First(&zoneKind, "slug = ?", models.NormalizeZoneKind(slug)).Error; err != nil {
		return nil, err
	}
	return &zoneKind, nil
}

func (h *zoneKindHandle) FindAll(ctx context.Context) ([]models.ZoneKind, error) {
	var zoneKinds []models.ZoneKind
	if err := h.db.WithContext(ctx).Order("name ASC, slug ASC").Find(&zoneKinds).Error; err != nil {
		return nil, err
	}
	return zoneKinds, nil
}

func (h *zoneKindHandle) Update(ctx context.Context, zoneKind *models.ZoneKind) error {
	return h.db.WithContext(ctx).Save(zoneKind).Error
}

func (h *zoneKindHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.ZoneKind{}, "id = ?", id).Error
}
