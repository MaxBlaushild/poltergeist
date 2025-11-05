package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type treasureChestHandle struct {
	db *gorm.DB
}

func (h *treasureChestHandle) Create(ctx context.Context, treasureChest *models.TreasureChest) error {
	treasureChest.ID = uuid.New()
	treasureChest.CreatedAt = time.Now()
	treasureChest.UpdatedAt = time.Now()

	if err := treasureChest.SetGeometry(treasureChest.Latitude, treasureChest.Longitude); err != nil {
		return err
	}

	return h.db.WithContext(ctx).Create(treasureChest).Error
}

func (h *treasureChestHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.TreasureChest, error) {
	var treasureChest models.TreasureChest
	if err := h.db.WithContext(ctx).
		Preload("Zone").
		Preload("Items").
		Preload("Items.InventoryItem").
		First(&treasureChest, id).Error; err != nil {
		return nil, err
	}
	return &treasureChest, nil
}

func (h *treasureChestHandle) FindAll(ctx context.Context) ([]models.TreasureChest, error) {
	var treasureChests []models.TreasureChest
	if err := h.db.WithContext(ctx).
		Preload("Zone").
		Preload("Items").
		Preload("Items.InventoryItem").
		Find(&treasureChests).Error; err != nil {
		return nil, err
	}
	return treasureChests, nil
}

func (h *treasureChestHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.TreasureChest, error) {
	var treasureChests []models.TreasureChest
	if err := h.db.WithContext(ctx).
		Where("zone_id = ? AND invalidated = false", zoneID).
		Preload("Zone").
		Preload("Items").
		Preload("Items.InventoryItem").
		Find(&treasureChests).Error; err != nil {
		return nil, err
	}
	return treasureChests, nil
}

func (h *treasureChestHandle) Update(ctx context.Context, id uuid.UUID, updates *models.TreasureChest) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()

	if updates.Latitude != 0 && updates.Longitude != 0 {
		if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
			return err
		}
	}

	return h.db.WithContext(ctx).Model(&models.TreasureChest{}).Where("id = ?", id).Updates(updates).Error
}

func (h *treasureChestHandle) Delete(ctx context.Context, id uuid.UUID) error {
	// Items are cascade deleted, so we just need to delete the chest
	return h.db.WithContext(ctx).Delete(&models.TreasureChest{}, "id = ?", id).Error
}

func (h *treasureChestHandle) AddItem(ctx context.Context, treasureChestID uuid.UUID, inventoryItemID int, quantity int) error {
	// Check if item already exists
	var existingItem models.TreasureChestItem
	result := h.db.WithContext(ctx).
		Where("treasure_chest_id = ? AND inventory_item_id = ?", treasureChestID, inventoryItemID).
		First(&existingItem)

	if result.Error == nil {
		// Update quantity
		existingItem.Quantity += quantity
		existingItem.UpdatedAt = time.Now()
		return h.db.WithContext(ctx).Save(&existingItem).Error
	} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create new item
		item := models.TreasureChestItem{
			ID:              uuid.New(),
			TreasureChestID: treasureChestID,
			InventoryItemID: inventoryItemID,
			Quantity:        quantity,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		return h.db.WithContext(ctx).Create(&item).Error
	}
	return result.Error
}

func (h *treasureChestHandle) RemoveItem(ctx context.Context, treasureChestID uuid.UUID, inventoryItemID int) error {
	return h.db.WithContext(ctx).
		Where("treasure_chest_id = ? AND inventory_item_id = ?", treasureChestID, inventoryItemID).
		Delete(&models.TreasureChestItem{}).Error
}

func (h *treasureChestHandle) UpdateItemQuantity(ctx context.Context, treasureChestID uuid.UUID, inventoryItemID int, quantity int) error {
	return h.db.WithContext(ctx).
		Model(&models.TreasureChestItem{}).
		Where("treasure_chest_id = ? AND inventory_item_id = ?", treasureChestID, inventoryItemID).
		Updates(map[string]interface{}{
			"quantity":   quantity,
			"updated_at": time.Now(),
		}).Error
}

func (h *treasureChestHandle) InvalidateByZoneID(ctx context.Context, zoneID uuid.UUID) error {
	return h.db.WithContext(ctx).
		Model(&models.TreasureChest{}).
		Where("zone_id = ?", zoneID).
		Updates(map[string]interface{}{
			"invalidated": true,
			"updated_at":  time.Now(),
		}).Error
}
