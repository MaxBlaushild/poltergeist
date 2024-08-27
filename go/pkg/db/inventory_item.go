package db

import (
	"context"
	"errors"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inventoryItemHandler struct {
	db *gorm.DB
}

func (h *inventoryItemHandler) CreateOrIncrementInventoryItem(ctx context.Context, teamID uuid.UUID, inventoryItemID int) error {
	var item models.TeamInventoryItem
	result := h.db.Where("team_id = ? AND inventory_item_id = ?", teamID, inventoryItemID).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			newItem := models.TeamInventoryItem{
				TeamID:          teamID,
				InventoryItemID: inventoryItemID,
				Quantity:        1,
			}
			return h.db.Create(&newItem).Error
		}
		return result.Error
	}
	item.Quantity += 1
	return h.db.Save(&item).Error
}

func (h *inventoryItemHandler) GetInventoryItem(ctx context.Context, teamID uuid.UUID, inventoryItemID int) (*models.TeamInventoryItem, error) {
	var item models.TeamInventoryItem
	result := h.db.Where("team_id = ? AND inventory_item_id = ?", teamID, inventoryItemID).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &item, nil
}

func (h *inventoryItemHandler) UseInventoryItem(ctx context.Context, teamInventoryItemID uuid.UUID) error {
	var item models.TeamInventoryItem
	result := h.db.Where("id = ?", teamInventoryItemID).First(&item)
	if result.Error != nil {
		return result.Error
	}
	item.Quantity -= 1
	return h.db.Save(&item).Error
}

func (h *inventoryItemHandler) ApplyInventoryItem(ctx context.Context, inventoryItemID int, teamID uuid.UUID) error {
	return h.db.Delete(&models.TeamInventoryItem{}, "id = ?", teamInventoryItemID).Error
}
