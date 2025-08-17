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

func (h *userEquipmentHandler) GetUserEquipment(ctx context.Context, userID uuid.UUID) (*models.UserEquipment, error) {
	var equipment models.UserEquipment
	result := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&equipment)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create new equipment record if none exists
			equipment = models.UserEquipment{
				ID:     uuid.New(),
				UserID: userID,
			}
			result = h.db.WithContext(ctx).Create(&equipment)
			if result.Error != nil {
				return nil, result.Error
			}
		} else {
			return nil, result.Error
		}
	}

	return &equipment, nil
}

func (h *userEquipmentHandler) EquipItem(ctx context.Context, userID uuid.UUID, inventoryItemID uuid.UUID, slot string) error {
	// Validate slot name matches one of the equipment slots
	switch slot {
	case "helm", "chest", "leftHand", "rightHand", "feet", "gloves", "neck", "leftRing", "rightRing", "leg":
		// Valid slot
	default:
		return errors.New("invalid equipment slot")
	}

	// Get or create user equipment
	equipment, err := h.GetUserEquipment(ctx, userID)
	if err != nil {
		return err
	}

	// Build update map based on slot
	updates := map[string]interface{}{
		slot + "_inventory_item_id": inventoryItemID,
		"updated_at":                gorm.Expr("NOW()"),
	}

	result := h.db.WithContext(ctx).
		Model(&models.UserEquipment{}).
		Where("id = ?", equipment.ID).
		Updates(updates)

	return result.Error
}

func (h *userEquipmentHandler) UnequipItem(ctx context.Context, userID uuid.UUID, slot string) error {
	// Validate slot name matches one of the equipment slots
	switch slot {
	case "helm", "chest", "leftHand", "rightHand", "feet", "gloves", "neck", "leftRing", "rightRing", "leg":
		// Valid slot
	default:
		return errors.New("invalid equipment slot")
	}

	// Get user equipment
	equipment, err := h.GetUserEquipment(ctx, userID)
	if err != nil {
		return err
	}

	// Build update map based on slot
	updates := map[string]interface{}{
		slot + "_inventory_item_id": nil,
		"updated_at":                gorm.Expr("NOW()"),
	}

	result := h.db.WithContext(ctx).
		Model(&models.UserEquipment{}).
		Where("id = ?", equipment.ID).
		Updates(updates)

	return result.Error
}

func (h *userEquipmentHandler) UnequipItemByOwnedInventoryItemID(ctx context.Context, userID uuid.UUID, inventoryItemID uuid.UUID) error {
	equipment, err := h.GetUserEquipment(ctx, userID)
	if err != nil {
		return err
	}

	// Build update map checking all slots
	updates := map[string]interface{}{
		"updated_at": gorm.Expr("NOW()"),
	}

	// Check each slot and unequip if matching
	if equipment.HelmInventoryItemID != nil && *equipment.HelmInventoryItemID == inventoryItemID {
		updates["helm_inventory_item_id"] = nil
	}
	if equipment.ChestInventoryItemID != nil && *equipment.ChestInventoryItemID == inventoryItemID {
		updates["chest_inventory_item_id"] = nil
	}
	if equipment.LeftHandInventoryItemID != nil && *equipment.LeftHandInventoryItemID == inventoryItemID {
		updates["left_hand_inventory_item_id"] = nil
	}
	if equipment.RightHandInventoryItemID != nil && *equipment.RightHandInventoryItemID == inventoryItemID {
		updates["right_hand_inventory_item_id"] = nil
	}
	if equipment.FeetInventoryItemID != nil && *equipment.FeetInventoryItemID == inventoryItemID {
		updates["feet_inventory_item_id"] = nil
	}
	if equipment.GlovesInventoryItemID != nil && *equipment.GlovesInventoryItemID == inventoryItemID {
		updates["gloves_inventory_item_id"] = nil
	}
	if equipment.NeckInventoryItemID != nil && *equipment.NeckInventoryItemID == inventoryItemID {
		updates["neck_inventory_item_id"] = nil
	}
	if equipment.LeftRingInventoryItemID != nil && *equipment.LeftRingInventoryItemID == inventoryItemID {
		updates["left_ring_inventory_item_id"] = nil
	}
	if equipment.RightRingInventoryItemID != nil && *equipment.RightRingInventoryItemID == inventoryItemID {
		updates["right_ring_inventory_item_id"] = nil
	}
	if equipment.LegInventoryItemID != nil && *equipment.LegInventoryItemID == inventoryItemID {
		updates["leg_inventory_item_id"] = nil
	}

	result := h.db.WithContext(ctx).
		Model(&models.UserEquipment{}).
		Where("id = ?", equipment.ID).
		Updates(updates)

	return result.Error
}
