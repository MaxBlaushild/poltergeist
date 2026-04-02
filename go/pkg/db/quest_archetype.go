package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questArchetypeHandle struct {
	db *gorm.DB
}

func (h *questArchetypeHandle) Create(ctx context.Context, questArchetype *models.QuestArchetype) error {
	if questArchetype != nil {
		questArchetype.Category = models.NormalizeQuestCategory(questArchetype.Category)
		questArchetype.DifficultyMode = models.NormalizeQuestDifficultyMode(string(questArchetype.DifficultyMode))
		questArchetype.Difficulty = models.NormalizeQuestDifficulty(questArchetype.Difficulty)
		questArchetype.MonsterEncounterTargetLevel = models.NormalizeMonsterEncounterTargetLevel(questArchetype.MonsterEncounterTargetLevel)
	}
	return h.db.WithContext(ctx).Create(questArchetype).Error
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
		Preload("Root.Challenges").
		Preload("Root.Challenges.ChallengeTemplate").
		Preload("Root.LocationArchetype").
		Preload("Root.ScenarioTemplate").
		First(&questArchetype, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &questArchetype, nil
}

func (h *questArchetypeHandle) Update(ctx context.Context, questArchetype *models.QuestArchetype) error {
	if questArchetype != nil {
		questArchetype.Category = models.NormalizeQuestCategory(questArchetype.Category)
		questArchetype.DifficultyMode = models.NormalizeQuestDifficultyMode(string(questArchetype.DifficultyMode))
		questArchetype.Difficulty = models.NormalizeQuestDifficulty(questArchetype.Difficulty)
		questArchetype.MonsterEncounterTargetLevel = models.NormalizeMonsterEncounterTargetLevel(questArchetype.MonsterEncounterTargetLevel)
	}
	return h.db.WithContext(ctx).Save(questArchetype).Error
}

func (h *questArchetypeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.QuestArchetype{}, "id = ?", id).Error
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
		Preload("Root.Challenges").
		Preload("Root.Challenges.ChallengeTemplate").
		Preload("Root.LocationArchetype").
		Preload("Root.ScenarioTemplate").
		Find(&questArchetypes).Error; err != nil {
		return nil, err
	}
	return questArchetypes, nil
}
