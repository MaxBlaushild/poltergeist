package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestImportHandle struct {
	db *gorm.DB
}

func (h *pointOfInterestImportHandle) Create(ctx context.Context, item *models.PointOfInterestImport) error {
	return h.db.WithContext(ctx).Create(item).Error
}

func (h *pointOfInterestImportHandle) Update(ctx context.Context, item *models.PointOfInterestImport) error {
	return h.db.WithContext(ctx).Save(item).Error
}

func (h *pointOfInterestImportHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestImport, error) {
	var item models.PointOfInterestImport
	if err := h.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (h *pointOfInterestImportHandle) FindRecent(ctx context.Context, limit int) ([]models.PointOfInterestImport, error) {
	var items []models.PointOfInterestImport
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (h *pointOfInterestImportHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID, limit int) ([]models.PointOfInterestImport, error) {
	var items []models.PointOfInterestImport
	q := h.db.WithContext(ctx).Where("zone_id = ?", zoneID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
