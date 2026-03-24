package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type districtHandle struct {
	db *gorm.DB
}

func (h *districtHandle) Create(ctx context.Context, district *models.District, zoneIDs []uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(district).Error; err != nil {
			return err
		}
		if err := replaceDistrictZonesTx(tx.WithContext(ctx), district.ID, zoneIDs); err != nil {
			return err
		}
		return tx.Preload("Zones", func(db *gorm.DB) *gorm.DB {
			return db.Order("zones.name ASC")
		}).First(district, "id = ?", district.ID).Error
	})
}

func (h *districtHandle) FindAll(ctx context.Context) ([]*models.District, error) {
	var districts []*models.District
	if err := h.db.WithContext(ctx).
		Preload("Zones", func(db *gorm.DB) *gorm.DB {
			return db.Order("zones.name ASC")
		}).
		Order("updated_at DESC").
		Find(&districts).Error; err != nil {
		return nil, err
	}
	return districts, nil
}

func (h *districtHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.District, error) {
	var district models.District
	if err := h.db.WithContext(ctx).
		Preload("Zones", func(db *gorm.DB) *gorm.DB {
			return db.Order("zones.name ASC")
		}).
		First(&district, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &district, nil
}

func (h *districtHandle) Update(ctx context.Context, districtID uuid.UUID, name string, description string, zoneIDs []uuid.UUID) (*models.District, error) {
	var district models.District
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.District{}).
			Where("id = ?", districtID).
			Updates(map[string]interface{}{
				"name":        name,
				"description": description,
				"updated_at":  time.Now(),
			}).Error; err != nil {
			return err
		}
		if err := replaceDistrictZonesTx(tx.WithContext(ctx), districtID, zoneIDs); err != nil {
			return err
		}
		return tx.Preload("Zones", func(db *gorm.DB) *gorm.DB {
			return db.Order("zones.name ASC")
		}).First(&district, "id = ?", districtID).Error
	})
	if err != nil {
		return nil, err
	}
	return &district, nil
}

func (h *districtHandle) Delete(ctx context.Context, districtID uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.District{}, "id = ?", districtID).Error
}

func replaceDistrictZonesTx(tx *gorm.DB, districtID uuid.UUID, zoneIDs []uuid.UUID) error {
	dedupedZoneIDs := dedupeUUIDs(zoneIDs)

	if err := tx.Where("district_id = ?", districtID).Delete(&models.DistrictZone{}).Error; err != nil {
		return err
	}

	if len(dedupedZoneIDs) == 0 {
		return nil
	}

	joins := make([]models.DistrictZone, 0, len(dedupedZoneIDs))
	now := time.Now()
	for _, zoneID := range dedupedZoneIDs {
		joins = append(joins, models.DistrictZone{
			ID:         uuid.New(),
			CreatedAt:  now,
			UpdatedAt:  now,
			DistrictID: districtID,
			ZoneID:     zoneID,
		})
	}

	return tx.Create(&joins).Error
}

func dedupeUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}

	seen := make(map[uuid.UUID]struct{}, len(ids))
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	return result
}
