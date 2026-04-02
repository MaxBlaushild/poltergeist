package db

import (
	"context"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questHandle struct {
	db *gorm.DB
}

func (h *questHandle) preloadDetail(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("SpellRewards").
		Preload("SpellRewards.Spell").
		Preload("Nodes").
		Preload("Nodes.Children").
		Preload("Nodes.Challenge").
		Preload("Nodes.Challenge.PointOfInterest").
		Preload("Nodes.Challenge.PointOfInterest.Tags").
		Preload("Nodes.Scenario").
		Preload("Nodes.Scenario.PointOfInterest").
		Preload("Nodes.Scenario.PointOfInterest.Tags").
		Preload("Nodes.Scenario.Options").
		Preload("Nodes.Monster").
		Preload("Nodes.MonsterEncounter").
		Preload("Nodes.MonsterEncounter.PointOfInterest").
		Preload("Nodes.MonsterEncounter.PointOfInterest.Tags").
		Preload("Nodes.MonsterEncounter.Members", func(db *gorm.DB) *gorm.DB {
			return db.Order("slot ASC").Order("created_at ASC")
		}).
		Preload("Nodes.MonsterEncounter.Members.Monster")
}

func (h *questHandle) Create(ctx context.Context, quest *models.Quest) error {
	if quest != nil {
		quest.Category = models.NormalizeQuestCategory(quest.Category)
		if !models.IsMainStoryQuestCategory(quest.Category) {
			quest.MainStoryPreviousQuestID = nil
			quest.MainStoryNextQuestID = nil
		}
		quest.DifficultyMode = models.NormalizeQuestDifficultyMode(string(quest.DifficultyMode))
		quest.Difficulty = models.NormalizeQuestDifficulty(quest.Difficulty)
		quest.MonsterEncounterTargetLevel = models.NormalizeMonsterEncounterTargetLevel(quest.MonsterEncounterTargetLevel)
		if strings.TrimSpace(string(quest.RewardMode)) == "" {
			if quest.Gold > 0 || quest.RewardExperience > 0 {
				quest.RewardMode = models.RewardModeExplicit
			} else {
				quest.RewardMode = models.RewardModeRandom
			}
		}
		quest.RewardMode = models.NormalizeRewardMode(string(quest.RewardMode))
		quest.RandomRewardSize = models.NormalizeRandomRewardSize(string(quest.RandomRewardSize))
		if quest.RewardExperience < 0 {
			quest.RewardExperience = 0
		}
	}
	return h.db.WithContext(ctx).Create(quest).Error
}

func (h *questHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Quest) error {
	if updates == nil {
		return nil
	}
	updates.DifficultyMode = models.NormalizeQuestDifficultyMode(string(updates.DifficultyMode))
	updates.Category = models.NormalizeQuestCategory(updates.Category)
	if !models.IsMainStoryQuestCategory(updates.Category) {
		updates.MainStoryPreviousQuestID = nil
		updates.MainStoryNextQuestID = nil
	}
	updates.Difficulty = models.NormalizeQuestDifficulty(updates.Difficulty)
	updates.MonsterEncounterTargetLevel = models.NormalizeMonsterEncounterTargetLevel(updates.MonsterEncounterTargetLevel)
	updates.RewardMode = models.NormalizeRewardMode(string(updates.RewardMode))
	updates.RandomRewardSize = models.NormalizeRandomRewardSize(string(updates.RandomRewardSize))
	if updates.RewardExperience < 0 {
		updates.RewardExperience = 0
	}
	payload := map[string]interface{}{
		"name":                           updates.Name,
		"description":                    updates.Description,
		"category":                       updates.Category,
		"acceptance_dialogue":            updates.AcceptanceDialogue,
		"image_url":                      updates.ImageURL,
		"zone_id":                        updates.ZoneID,
		"quest_archetype_id":             updates.QuestArchetypeID,
		"quest_giver_character_id":       updates.QuestGiverCharacterID,
		"main_story_previous_quest_id":   updates.MainStoryPreviousQuestID,
		"main_story_next_quest_id":       updates.MainStoryNextQuestID,
		"recurring_quest_id":             updates.RecurringQuestID,
		"recurrence_frequency":           updates.RecurrenceFrequency,
		"next_recurrence_at":             updates.NextRecurrenceAt,
		"difficulty_mode":                updates.DifficultyMode,
		"difficulty":                     updates.Difficulty,
		"monster_encounter_target_level": updates.MonsterEncounterTargetLevel,
		"reward_mode":                    updates.RewardMode,
		"random_reward_size":             updates.RandomRewardSize,
		"reward_experience":              updates.RewardExperience,
		"gold":                           updates.Gold,
		"updated_at":                     updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.Quest{}).Where("id = ?", id).Updates(payload).Error
}

func (h *questHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Quest, error) {
	var quest models.Quest
	if err := h.preloadDetail(ctx).First(&quest, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &quest, nil
}

func (h *questHandle) FindAllSummaries(ctx context.Context) ([]models.Quest, error) {
	var quests []models.Quest
	nodeCounts := h.db.WithContext(ctx).
		Model(&models.QuestNode{}).
		Select("quest_id, COUNT(*) AS node_count").
		Group("quest_id")

	if err := h.db.WithContext(ctx).
		Model(&models.Quest{}).
		Select("quests.*, COALESCE(node_counts.node_count, 0) AS node_count").
		Joins("LEFT JOIN (?) AS node_counts ON node_counts.quest_id = quests.id", nodeCounts).
		Order("quests.updated_at DESC").
		Order("quests.created_at DESC").
		Find(&quests).Error; err != nil {
		return nil, err
	}
	return quests, nil
}

func (h *questHandle) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Quest, error) {
	var quests []models.Quest
	if err := h.preloadDetail(ctx).Where("id IN ?", ids).Find(&quests).Error; err != nil {
		return nil, err
	}
	return quests, nil
}

func (h *questHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Quest, error) {
	var quests []models.Quest
	if err := h.db.WithContext(ctx).
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("SpellRewards").
		Preload("SpellRewards.Spell").
		Preload("Nodes").
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
		Preload("SpellRewards").
		Preload("SpellRewards.Spell").
		Preload("Nodes").
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
		Preload("SpellRewards").
		Preload("SpellRewards.Spell").
		Preload("Nodes").
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
