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

func (h *inventoryItemHandler) GetItems(ctx context.Context, userOrTeam models.OwnedInventoryItem) ([]models.OwnedInventoryItem, error) {
	var items []models.OwnedInventoryItem

	if userOrTeam.TeamID != nil {
		result := h.db.Where("team_id = ?", userOrTeam.TeamID).Find(&items)
		if result.Error != nil {
			return nil, result.Error
		}
	} else {
		result := h.db.Where("user_id = ?", userOrTeam.UserID).Find(&items)
		if result.Error != nil {
			return nil, result.Error
		}
	}
	return items, nil
}

func (h *inventoryItemHandler) GetUsersItems(ctx context.Context, userID uuid.UUID) ([]models.OwnedInventoryItem, error) {
	var items []models.OwnedInventoryItem
	result := h.db.Where("user_id = ?", userID).Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (h *inventoryItemHandler) StealItems(ctx context.Context, thiefTeamID uuid.UUID, victimTeamID uuid.UUID) error {
	items, err := h.GetItems(ctx, models.OwnedInventoryItem{TeamID: &victimTeamID})
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := h.CreateOrIncrementInventoryItem(ctx, &thiefTeamID, nil, item.InventoryItemID, item.Quantity); err != nil {
			return err
		}
		item.Quantity = 0
		h.db.Save(&item)
	}
	return nil
}

func (h *inventoryItemHandler) StealItem(ctx context.Context, thiefTeamID uuid.UUID, victimTeamID uuid.UUID, inventoryItemID int) error {
	items, err := h.GetItems(ctx, models.OwnedInventoryItem{TeamID: &victimTeamID})
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.InventoryItemID == inventoryItemID {
			if err := h.CreateOrIncrementInventoryItem(ctx, &thiefTeamID, nil, item.InventoryItemID, item.Quantity); err != nil {
				return err
			}
			item.Quantity = 0
			h.db.Save(&item)
		}
	}
	return nil
}

func (h *inventoryItemHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.OwnedInventoryItem, error) {
	var item models.OwnedInventoryItem
	result := h.db.Where("id = ?", id).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}

func (h *inventoryItemHandler) CreateOrIncrementInventoryItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, inventoryItemID int, quantity int) error {
	var item models.OwnedInventoryItem
	var query string
	var queryID *uuid.UUID

	if teamID != nil {
		query = "team_id = ? AND inventory_item_id = ?"
		queryID = teamID
	} else {
		query = "user_id = ? AND inventory_item_id = ?"
		queryID = userID
	}

	result := h.db.Where(query, queryID, inventoryItemID).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			newItem := models.OwnedInventoryItem{
				ID:              uuid.New(),
				TeamID:          teamID,
				UserID:          userID,
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

func (h *inventoryItemHandler) UseInventoryItem(ctx context.Context, ownedInventoryItemID uuid.UUID) error {
	var item models.OwnedInventoryItem
	result := h.db.Where("id = ?", ownedInventoryItemID).First(&item)
	if result.Error != nil {
		return result.Error
	}
	
	// If this item is equipped and we're consuming it, unequip it first
	if item.UserID != nil {
		h.db.Where("owned_inventory_item_id = ?", ownedInventoryItemID).Delete(&models.UserEquipment{})
	}
	
	item.Quantity -= 1
	return h.db.Save(&item).Error
}

func (h *inventoryItemHandler) ApplyInventoryItem(ctx context.Context, matchID uuid.UUID, inventoryItemID int, teamID uuid.UUID, duration time.Duration) error {
	newEffect := models.MatchInventoryItemEffect{
		ID:              uuid.New(),
		MatchID:         matchID,
		TeamID:          teamID,
		InventoryItemID: inventoryItemID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(duration),
	}
	return h.db.Create(&newEffect).Error
}
