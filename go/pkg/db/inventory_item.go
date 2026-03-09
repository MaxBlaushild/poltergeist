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
	item.Quantity -= 1
	if err := h.db.Save(&item).Error; err != nil {
		return err
	}
	if item.Quantity <= 0 {
		_ = h.db.WithContext(ctx).
			Where("owned_inventory_item_id = ?", item.ID).
			Delete(&models.UserEquipment{}).Error
	}
	return nil
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

func (h *inventoryItemHandler) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	// Delete using both direct comparison and IS NOT NULL check to ensure we get all user items
	return h.db.WithContext(ctx).Where("user_id = ? AND user_id IS NOT NULL", userID).Delete(&models.OwnedInventoryItem{}).Error
}

// CRUD methods for inventory items
func (h *inventoryItemHandler) CreateInventoryItem(ctx context.Context, item *models.InventoryItem) error {
	if item != nil {
		if item.ItemLevel <= 0 {
			item.ItemLevel = 1
		}
		if item.ConsumeStatusesToRemove == nil {
			item.ConsumeStatusesToRemove = models.StringArray{}
		}
		if item.ConsumeSpellIDs == nil {
			item.ConsumeSpellIDs = models.StringArray{}
		}
		if item.InternalTags == nil {
			item.InternalTags = models.StringArray{}
		}
	}
	return h.db.WithContext(ctx).Create(item).Error
}

func (h *inventoryItemHandler) FindInventoryItemByID(ctx context.Context, id int) (*models.InventoryItem, error) {
	var item models.InventoryItem
	result := h.db.WithContext(ctx).Where("id = ?", id).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}

func (h *inventoryItemHandler) FindAllInventoryItems(ctx context.Context) ([]models.InventoryItem, error) {
	var items []models.InventoryItem
	result := h.db.WithContext(ctx).Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (h *inventoryItemHandler) UpdateInventoryItem(ctx context.Context, id int, updates map[string]interface{}) error {
	if updates == nil {
		return nil
	}
	if value, exists := updates["consume_statuses_to_remove"]; exists && value == nil {
		updates["consume_statuses_to_remove"] = models.StringArray{}
	}
	if value, exists := updates["consume_spell_ids"]; exists && value == nil {
		updates["consume_spell_ids"] = models.StringArray{}
	}
	if value, exists := updates["internal_tags"]; exists && value == nil {
		updates["internal_tags"] = models.StringArray{}
	}
	updates["updated_at"] = time.Now()
	return h.db.WithContext(ctx).Model(&models.InventoryItem{}).Where("id = ?", id).Updates(updates).Error
}

func (h *inventoryItemHandler) DeleteInventoryItem(ctx context.Context, id int) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := h.clearInventoryItemReferences(tx, id); err != nil {
			return err
		}
		return tx.Delete(&models.InventoryItem{}, id).Error
	})
}

func (h *inventoryItemHandler) DecrementUserInventoryItem(ctx context.Context, userID uuid.UUID, inventoryItemID int, quantity int) error {
	var item models.OwnedInventoryItem
	result := h.db.WithContext(ctx).Where("user_id = ? AND inventory_item_id = ?", userID, inventoryItemID).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("user does not own this item")
		}
		return result.Error
	}

	if item.Quantity < quantity {
		return errors.New("insufficient quantity")
	}

	item.Quantity -= quantity
	if item.Quantity <= 0 {
		// Delete the item if quantity reaches 0 or below
		return h.db.WithContext(ctx).Delete(&item).Error
	}
	return h.db.WithContext(ctx).Save(&item).Error
}

func (h *inventoryItemHandler) clearInventoryItemReferences(tx *gorm.DB, inventoryItemID int) error {
	rewardAndJoinTables := []string{
		"monster_item_rewards",
		"quest_item_rewards",
		"quest_archetype_item_rewards",
		"scenario_item_rewards",
		"scenario_option_item_rewards",
		"treasure_chest_items",
		"owned_inventory_items",
		"match_inventory_item_effects",
		"team_inventory_items",
		"outfit_profile_generations",
	}
	for _, table := range rewardAndJoinTables {
		if err := deleteByInventoryItemIDIfTableExists(tx, table, inventoryItemID); err != nil {
			return err
		}
	}

	nullableItemColumns := []struct {
		table  string
		column string
	}{
		{table: "point_of_interest_groups", column: "inventory_item_id"},
		{table: "quest_node_challenges", column: "inventory_item_id"},
		{table: "quest_archetype_challenges", column: "inventory_item_id"},
		{table: "monsters", column: "dominant_hand_inventory_item_id"},
		{table: "monsters", column: "off_hand_inventory_item_id"},
		{table: "monsters", column: "weapon_inventory_item_id"},
	}
	for _, entry := range nullableItemColumns {
		if err := nullifyColumnByInventoryItemIDIfTableExists(tx, entry.table, entry.column, inventoryItemID); err != nil {
			return err
		}
	}

	// This column is historically non-nullable and uses 0 as "no reward item".
	if tx.Migrator().HasTable("point_of_interest_challenges") {
		if err := tx.Exec(
			"UPDATE point_of_interest_challenges SET inventory_item_id = 0 WHERE inventory_item_id = ?",
			inventoryItemID,
		).Error; err != nil {
			return err
		}
	}

	return removeInventoryItemFromStarterConfigsIfPresent(tx, inventoryItemID)
}

func deleteByInventoryItemIDIfTableExists(tx *gorm.DB, tableName string, inventoryItemID int) error {
	if !tx.Migrator().HasTable(tableName) {
		return nil
	}
	return tx.Exec("DELETE FROM "+tableName+" WHERE inventory_item_id = ?", inventoryItemID).Error
}

func nullifyColumnByInventoryItemIDIfTableExists(tx *gorm.DB, tableName string, columnName string, inventoryItemID int) error {
	if !tx.Migrator().HasTable(tableName) {
		return nil
	}
	return tx.Exec(
		"UPDATE "+tableName+" SET "+columnName+" = NULL WHERE "+columnName+" = ?",
		inventoryItemID,
	).Error
}

func removeInventoryItemFromStarterConfigsIfPresent(tx *gorm.DB, inventoryItemID int) error {
	if !tx.Migrator().HasTable("new_user_starter_configs") {
		return nil
	}

	var configs []models.NewUserStarterConfig
	if err := tx.Find(&configs).Error; err != nil {
		return err
	}

	for i := range configs {
		filtered := make([]models.NewUserStarterItem, 0, len(configs[i].Items))
		changed := false
		for _, item := range configs[i].Items {
			if item.InventoryItemID == inventoryItemID {
				changed = true
				continue
			}
			filtered = append(filtered, item)
		}
		if !changed {
			continue
		}
		configs[i].Items = filtered
		configs[i].UpdatedAt = time.Now()
		if err := tx.Save(&configs[i]).Error; err != nil {
			return err
		}
	}

	return nil
}
