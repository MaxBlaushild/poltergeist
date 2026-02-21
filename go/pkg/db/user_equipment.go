package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userEquipmentHandler struct {
	db *gorm.DB
}

func (h *userEquipmentHandler) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserEquipment, error) {
	var equipment []models.UserEquipment
	result := h.db.WithContext(ctx).Where("user_id = ?", userID).Find(&equipment)
	if result.Error != nil {
		return nil, result.Error
	}
	return equipment, nil
}

func (h *userEquipmentHandler) Equip(ctx context.Context, userID uuid.UUID, slot string, ownedInventoryItemID uuid.UUID) (*models.UserEquipment, error) {
	var result *models.UserEquipment
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND owned_inventory_item_id = ?", userID, ownedInventoryItemID).
			Delete(&models.UserEquipment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND slot = ?", userID, slot).
			Delete(&models.UserEquipment{}).Error; err != nil {
			return err
		}
		now := time.Now()
		equipment := &models.UserEquipment{
			ID:                  uuid.New(),
			UserID:              userID,
			Slot:                slot,
			OwnedInventoryItemID: ownedInventoryItemID,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
		if err := tx.Create(equipment).Error; err != nil {
			return err
		}
		result = equipment
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (h *userEquipmentHandler) UnequipSlot(ctx context.Context, userID uuid.UUID, slot string) error {
	return h.db.WithContext(ctx).Where("user_id = ? AND slot = ?", userID, slot).
		Delete(&models.UserEquipment{}).Error
}

func (h *userEquipmentHandler) UnequipOwnedItem(ctx context.Context, userID uuid.UUID, ownedInventoryItemID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("user_id = ? AND owned_inventory_item_id = ?", userID, ownedInventoryItemID).
		Delete(&models.UserEquipment{}).Error
}

func (h *userEquipmentHandler) GetStatBonuses(ctx context.Context, userID uuid.UUID) (models.CharacterStatBonuses, error) {
	var bonuses models.CharacterStatBonuses
	result := h.db.WithContext(ctx).
		Table("user_equipment").
		Select(`
			COALESCE(SUM(inventory_items.strength_mod), 0) AS strength,
			COALESCE(SUM(inventory_items.dexterity_mod), 0) AS dexterity,
			COALESCE(SUM(inventory_items.constitution_mod), 0) AS constitution,
			COALESCE(SUM(inventory_items.intelligence_mod), 0) AS intelligence,
			COALESCE(SUM(inventory_items.wisdom_mod), 0) AS wisdom,
			COALESCE(SUM(inventory_items.charisma_mod), 0) AS charisma
		`).
		Joins("JOIN owned_inventory_items ON owned_inventory_items.id = user_equipment.owned_inventory_item_id").
		Joins("JOIN inventory_items ON inventory_items.id = owned_inventory_items.inventory_item_id").
		Where("user_equipment.user_id = ? AND owned_inventory_items.quantity > 0", userID).
		Scan(&bonuses)
	if result.Error != nil {
		return models.CharacterStatBonuses{}, result.Error
	}
	return bonuses, nil
}
