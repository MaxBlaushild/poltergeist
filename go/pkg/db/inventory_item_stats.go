package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inventoryItemStatsHandler struct {
	db *gorm.DB
}

func (h *inventoryItemStatsHandler) FindByInventoryItemID(ctx context.Context, inventoryItemID int) (*models.InventoryItemStats, error) {
	var stats models.InventoryItemStats
	result := h.db.WithContext(ctx).Where("inventory_item_id = ?", inventoryItemID).First(&stats)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No stats found, which is valid
		}
		return nil, result.Error
	}
	return &stats, nil
}

func (h *inventoryItemStatsHandler) Create(ctx context.Context, stats *models.InventoryItemStats) error {
	return h.db.WithContext(ctx).Create(stats).Error
}

func (h *inventoryItemStatsHandler) Update(ctx context.Context, stats *models.InventoryItemStats) error {
	return h.db.WithContext(ctx).Save(stats).Error
}

func (h *inventoryItemStatsHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.InventoryItemStats{}, id).Error
}

func (h *inventoryItemStatsHandler) DeleteByInventoryItemID(ctx context.Context, inventoryItemID int) error {
	return h.db.WithContext(ctx).Where("inventory_item_id = ?", inventoryItemID).Delete(&models.InventoryItemStats{}).Error
}

func (h *inventoryItemStatsHandler) CreateOrUpdate(ctx context.Context, stats *models.InventoryItemStats) error {
	existing, err := h.FindByInventoryItemID(ctx, stats.InventoryItemID)
	if err != nil {
		return err
	}

	if existing == nil {
		// Create new stats
		return h.Create(ctx, stats)
	} else {
		// Update existing stats
		stats.ID = existing.ID
		stats.CreatedAt = existing.CreatedAt
		return h.Update(ctx, stats)
	}
}