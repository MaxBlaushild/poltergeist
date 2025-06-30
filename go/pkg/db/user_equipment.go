package db

import (
	"context"
	"errors"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userEquipmentHandler struct {
	db *gorm.DB
}

func (h *userEquipmentHandler) GetUserEquipment(ctx context.Context, userID uuid.UUID) ([]models.UserEquipment, error) {
	var equipment []models.UserEquipment
	result := h.db.WithContext(ctx).
		Preload("OwnedInventoryItem").
		Where("user_id = ?", userID).
		Find(&equipment)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return equipment, nil
}

func (h *userEquipmentHandler) EquipItem(ctx context.Context, userID uuid.UUID, ownedInventoryItemID uuid.UUID, equipmentSlot string) (*models.UserEquipment, error) {
	// First check if there's already an item equipped in this slot
	var existingEquipment models.UserEquipment
	result := h.db.WithContext(ctx).Where("user_id = ? AND equipment_slot = ?", userID, equipmentSlot).First(&existingEquipment)
	
	// If an item is already equipped in this slot, unequip it first
	if result.Error == nil {
		if err := h.UnequipItem(ctx, userID, equipmentSlot); err != nil {
			return nil, err
		}
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	
	// Create new equipment entry
	newEquipment := models.UserEquipment{
		ID:                   uuid.New(),
		UserID:               userID,
		EquipmentSlot:        equipmentSlot,
		OwnedInventoryItemID: ownedInventoryItemID,
	}
	
	result = h.db.WithContext(ctx).Create(&newEquipment)
	if result.Error != nil {
		return nil, result.Error
	}
	
	// Load the relationship data
	result = h.db.WithContext(ctx).
		Preload("OwnedInventoryItem").
		Where("id = ?", newEquipment.ID).
		First(&newEquipment)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return &newEquipment, nil
}

func (h *userEquipmentHandler) UnequipItem(ctx context.Context, userID uuid.UUID, equipmentSlot string) error {
	result := h.db.WithContext(ctx).Where("user_id = ? AND equipment_slot = ?", userID, equipmentSlot).Delete(&models.UserEquipment{})
	return result.Error
}

func (h *userEquipmentHandler) UnequipItemByOwnedInventoryItemID(ctx context.Context, userID uuid.UUID, ownedInventoryItemID uuid.UUID) error {
	result := h.db.WithContext(ctx).Where("user_id = ? AND owned_inventory_item_id = ?", userID, ownedInventoryItemID).Delete(&models.UserEquipment{})
	return result.Error
}

func (h *userEquipmentHandler) GetEquippedItemInSlot(ctx context.Context, userID uuid.UUID, equipmentSlot string) (*models.UserEquipment, error) {
	var equipment models.UserEquipment
	result := h.db.WithContext(ctx).
		Preload("OwnedInventoryItem").
		Where("user_id = ? AND equipment_slot = ?", userID, equipmentSlot).
		First(&equipment)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	
	return &equipment, nil
}