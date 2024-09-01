package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inventoryItemHandler struct {
	db *gorm.DB
}

func (h *inventoryItemHandler) GetTeamsItems(ctx context.Context, teamID uuid.UUID) ([]models.TeamInventoryItem, error) {
	var items []models.TeamInventoryItem
	result := h.db.Where("team_id = ?", teamID).Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (h *inventoryItemHandler) StealItems(ctx context.Context, thiefTeamID uuid.UUID, victimTeamID uuid.UUID) error {
	items, err := h.GetTeamsItems(ctx, victimTeamID)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := h.CreateOrIncrementInventoryItem(ctx, thiefTeamID, item.InventoryItemID, item.Quantity); err != nil {
			return err
		}
		item.Quantity = 0
		h.db.Save(&item)
	}
	return nil
}

func (h *inventoryItemHandler) StealItem(ctx context.Context, thiefTeamID uuid.UUID, victimTeamID uuid.UUID, inventoryItemID int) error {
	items, err := h.GetTeamsItems(ctx, victimTeamID)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.InventoryItemID == inventoryItemID {
			if err := h.CreateOrIncrementInventoryItem(ctx, thiefTeamID, item.InventoryItemID, item.Quantity); err != nil {
				return err
			}
			item.Quantity = 0
			h.db.Save(&item)
		}
	}
	return nil
}

func (h *inventoryItemHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.TeamInventoryItem, error) {
	var item models.TeamInventoryItem
	result := h.db.Where("id = ?", id).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}

func (h *inventoryItemHandler) CreateOrIncrementInventoryItem(ctx context.Context, teamID uuid.UUID, inventoryItemID int, quantity int) error {
	var item models.TeamInventoryItem
	result := h.db.Where("team_id = ? AND inventory_item_id = ?", teamID, inventoryItemID).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			newItem := models.TeamInventoryItem{
				ID:              uuid.New(),
				TeamID:          teamID,
				InventoryItemID: inventoryItemID,
				Quantity:        quantity,
			}
			return h.db.Create(&newItem).Error
		}
		return result.Error
	}
	item.Quantity += quantity
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

func (h *inventoryItemHandler) ApplyInventoryItem(ctx context.Context, matchID uuid.UUID, inventoryItemID int, teamID uuid.UUID, duration time.Duration) error {
	newEffect := models.MatchInventoryItemEffect{
		ID:              uuid.New(),
		MatchID:         matchID, // Assuming matchID is defined in the context where this function is called
		TeamID:          teamID,
		InventoryItemID: inventoryItemID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(duration),
	}
	return h.db.Create(&newEffect).Error
}
