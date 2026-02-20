package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type zoneImportHandle struct {
	db *gorm.DB
}

func (h *zoneImportHandle) Create(ctx context.Context, item *models.ZoneImport) error {
	return h.db.WithContext(ctx).Create(item).Error
}

func (h *zoneImportHandle) Update(ctx context.Context, item *models.ZoneImport) error {
	return h.db.WithContext(ctx).Save(item).Error
}

func (h *zoneImportHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ZoneImport, error) {
	var item models.ZoneImport
	if err := h.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (h *zoneImportHandle) FindRecent(ctx context.Context, limit int) ([]models.ZoneImport, error) {
	var items []models.ZoneImport
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (h *zoneImportHandle) FindByMetroName(ctx context.Context, metroName string, limit int) ([]models.ZoneImport, error) {
	var items []models.ZoneImport
	q := h.db.WithContext(ctx).Where("metro_name = ?", metroName).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
