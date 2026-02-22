package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questHandle struct {
	db *gorm.DB
}

func (h *questHandle) Create(ctx context.Context, quest *models.Quest) error {
	return h.db.WithContext(ctx).Create(quest).Error
}

func (h *questHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Quest) error {
	if updates == nil {
		return nil
	}
	payload := map[string]interface{}{
		"name":                     updates.Name,
		"description":              updates.Description,
		"acceptance_dialogue":      updates.AcceptanceDialogue,
		"image_url":                updates.ImageURL,
		"zone_id":                  updates.ZoneID,
		"quest_archetype_id":       updates.QuestArchetypeID,
		"quest_giver_character_id": updates.QuestGiverCharacterID,
		"recurring_quest_id":       updates.RecurringQuestID,
		"recurrence_frequency":     updates.RecurrenceFrequency,
		"next_recurrence_at":        updates.NextRecurrenceAt,
		"gold":                     updates.Gold,
		"updated_at":               updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.Quest{}).Where("id = ?", id).Updates(payload).Error
}

func (h *questHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Quest, error) {
	var quest models.Quest
	if err := h.db.WithContext(ctx).
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("Nodes", func(db *gorm.DB) *gorm.DB {
			return db.Select("quest_nodes.*, ST_AsText(quest_nodes.polygon) as polygon")
		}).
		Preload("Nodes.Challenges").
		Preload("Nodes.Children").
		First(&quest, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &quest, nil
}

func (h *questHandle) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Quest, error) {
	var quests []models.Quest
	if err := h.db.WithContext(ctx).
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("Nodes", func(db *gorm.DB) *gorm.DB {
			return db.Select("quest_nodes.*, ST_AsText(quest_nodes.polygon) as polygon")
		}).
		Preload("Nodes.Challenges").
		Preload("Nodes.Children").
		Where("id IN ?", ids).
		Find(&quests).Error; err != nil {
		return nil, err
	}
	return quests, nil
}

func (h *questHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Quest, error) {
	var quests []models.Quest
	if err := h.db.WithContext(ctx).
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("Nodes", func(db *gorm.DB) *gorm.DB {
			return db.Select("quest_nodes.*, ST_AsText(quest_nodes.polygon) as polygon")
		}).
		Preload("Nodes.Challenges").
		Preload("Nodes.Children").
		Where("zone_id = ?", zoneID).
		Find(&quests).Error; err != nil {
		return nil, err
	}
	return quests, nil
}

func (h *questHandle) FindByQuestGiverCharacterID(ctx context.Context, characterID uuid.UUID) ([]models.Quest, error) {
	var quests []models.Quest
	if err := h.db.WithContext(ctx).
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("Nodes", func(db *gorm.DB) *gorm.DB {
			return db.Select("quest_nodes.*, ST_AsText(quest_nodes.polygon) as polygon")
		}).
		Preload("Nodes.Challenges").
		Preload("Nodes.Children").
		Where("quest_giver_character_id = ?", characterID).
		Find(&quests).Error; err != nil {
		return nil, err
	}
	return quests, nil
}

func (h *questHandle) FindAll(ctx context.Context) ([]models.Quest, error) {
	var quests []models.Quest
	if err := h.db.WithContext(ctx).
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("Nodes", func(db *gorm.DB) *gorm.DB {
			return db.Select("quest_nodes.*, ST_AsText(quest_nodes.polygon) as polygon")
		}).
		Preload("Nodes.Challenges").
		Preload("Nodes.Children").
		Find(&quests).Error; err != nil {
		return nil, err
	}
	return quests, nil
}

func (h *questHandle) FindDueRecurring(ctx context.Context, asOf time.Time, limit int) ([]models.Quest, error) {
	var quests []models.Quest
	query := h.db.WithContext(ctx).
		Where("recurrence_frequency IS NOT NULL AND recurrence_frequency <> ''").
		Where("next_recurrence_at IS NOT NULL AND next_recurrence_at <= ?", asOf).
		Order("next_recurrence_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&quests).Error; err != nil {
		return nil, err
	}
	return quests, nil
}

func (h *questHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Quest{}, "id = ?", id).Error
}
