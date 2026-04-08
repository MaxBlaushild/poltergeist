package db

import (
	"context"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type questArchetypeHandle struct {
	db *gorm.DB
}

func (h *questArchetypeHandle) Create(ctx context.Context, questArchetype *models.QuestArchetype) error {
	if questArchetype != nil {
		questArchetype.Category = models.NormalizeQuestCategory(questArchetype.Category)
		questArchetype.RequiredStoryFlags = normalizeJSONStringArray(questArchetype.RequiredStoryFlags)
		questArchetype.SetStoryFlags = normalizeJSONStringArray(questArchetype.SetStoryFlags)
		questArchetype.ClearStoryFlags = normalizeJSONStringArray(questArchetype.ClearStoryFlags)
		questArchetype.QuestGiverRelationshipEffects = normalizeCharacterRelationshipState(questArchetype.QuestGiverRelationshipEffects)
		questArchetype.DifficultyMode = models.NormalizeQuestDifficultyMode(string(questArchetype.DifficultyMode))
		questArchetype.Difficulty = models.NormalizeQuestDifficulty(questArchetype.Difficulty)
		questArchetype.MonsterEncounterTargetLevel = models.NormalizeMonsterEncounterTargetLevel(questArchetype.MonsterEncounterTargetLevel)
		if questArchetype.Root.ID != uuid.Nil {
			log.Printf(
				"[main-story-convert][quest-archetype][create] archetype=%s includes root struct %s; omitting associations on save",
				questArchetype.ID.String(),
				questArchetype.Root.ID.String(),
			)
		}
	}
	return h.db.WithContext(ctx).
		Omit(clause.Associations).
		Create(questArchetype).Error
}

// func loadChallengesWithUnlockedNodes(db *gorm.DB) *gorm.DB {
// 	return db.Preload("Challenges", func(db *gorm.DB) *gorm.DB {
// 		return db.Preload("UnlockedNode", func(db *gorm.DB) *gorm.DB {
// 			return db.Preload("Challenges", func(db *gorm.DB) *gorm.DB {
// 				return db.Preload("UnlockedNode", func(db *gorm.DB) *gorm.DB {
// 					return db.Preload("Challenges", func(db *gorm.DB) *gorm.DB {
// 						return db.Preload("UnlockedNode")
// 					})
// 				})
// 			})
// 		})
// 	})
// }

func (h *questArchetypeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetype, error) {
	var questArchetype models.QuestArchetype
	if err := h.db.WithContext(ctx).
		Preload("QuestGiverCharacter").
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("SpellRewards").
		Preload("SpellRewards.Spell").
		Preload("Root").
		Preload("Root.ChallengeTemplate").
		Preload("Root.Challenges").
		Preload("Root.Challenges.ChallengeTemplate").
		Preload("Root.LocationArchetype").
		Preload("Root.ScenarioTemplate").
		First(&questArchetype, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &questArchetype, nil
}

func (h *questArchetypeHandle) Update(ctx context.Context, questArchetype *models.QuestArchetype) error {
	if questArchetype != nil {
		questArchetype.Category = models.NormalizeQuestCategory(questArchetype.Category)
		questArchetype.RequiredStoryFlags = normalizeJSONStringArray(questArchetype.RequiredStoryFlags)
		questArchetype.SetStoryFlags = normalizeJSONStringArray(questArchetype.SetStoryFlags)
		questArchetype.ClearStoryFlags = normalizeJSONStringArray(questArchetype.ClearStoryFlags)
		questArchetype.QuestGiverRelationshipEffects = normalizeCharacterRelationshipState(questArchetype.QuestGiverRelationshipEffects)
		questArchetype.DifficultyMode = models.NormalizeQuestDifficultyMode(string(questArchetype.DifficultyMode))
		questArchetype.Difficulty = models.NormalizeQuestDifficulty(questArchetype.Difficulty)
		questArchetype.MonsterEncounterTargetLevel = models.NormalizeMonsterEncounterTargetLevel(questArchetype.MonsterEncounterTargetLevel)
	}
	return h.db.WithContext(ctx).
		Omit(clause.Associations).
		Save(questArchetype).Error
}

func (h *questArchetypeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.QuestArchetype{}, "id = ?", id).Error
}

func (h *questArchetypeHandle) DeletePermanent(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Unscoped().Delete(&models.QuestArchetype{}, "id = ?", id).Error
}

func (h *questArchetypeHandle) ClearQuestGiverCharacterIDByCharacterID(
	ctx context.Context,
	characterID uuid.UUID,
) error {
	return h.db.WithContext(ctx).
		Unscoped().
		Model(&models.QuestArchetype{}).
		Where("quest_giver_character_id = ?", characterID).
		Updates(map[string]interface{}{
			"quest_giver_character_id": nil,
			"updated_at":               time.Now(),
		}).Error
}

func (h *questArchetypeHandle) FindAll(ctx context.Context) ([]*models.QuestArchetype, error) {
	var questArchetypes []*models.QuestArchetype
	if err := h.db.WithContext(ctx).
		Preload("QuestGiverCharacter").
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("SpellRewards").
		Preload("SpellRewards.Spell").
		Preload("Root").
		Preload("Root.ChallengeTemplate").
		Preload("Root.Challenges").
		Preload("Root.Challenges.ChallengeTemplate").
		Preload("Root.LocationArchetype").
		Preload("Root.ScenarioTemplate").
		Find(&questArchetypes).Error; err != nil {
		return nil, err
	}
	return questArchetypes, nil
}
