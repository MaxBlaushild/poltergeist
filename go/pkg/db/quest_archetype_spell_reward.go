package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questArchetypeSpellRewardHandle struct {
	db *gorm.DB
}

func (h *questArchetypeSpellRewardHandle) ReplaceForQuestArchetype(ctx context.Context, questArchetypeID uuid.UUID, rewards []models.QuestArchetypeSpellReward) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("quest_archetype_id = ?", questArchetypeID).Delete(&models.QuestArchetypeSpellReward{}).Error; err != nil {
			return err
		}
		if len(rewards) == 0 {
			return nil
		}
		now := time.Now()
		for i := range rewards {
			if rewards[i].ID == uuid.Nil {
				rewards[i].ID = uuid.New()
			}
			rewards[i].QuestArchetypeID = questArchetypeID
			if rewards[i].CreatedAt.IsZero() {
				rewards[i].CreatedAt = now
			}
			rewards[i].UpdatedAt = now
		}
		return tx.Create(&rewards).Error
	})
}

func (h *questArchetypeSpellRewardHandle) FindByQuestArchetypeID(ctx context.Context, questArchetypeID uuid.UUID) ([]models.QuestArchetypeSpellReward, error) {
	var rewards []models.QuestArchetypeSpellReward
	if err := h.db.WithContext(ctx).
		Preload("Spell").
		Where("quest_archetype_id = ?", questArchetypeID).
		Find(&rewards).Error; err != nil {
		return nil, err
	}
	return rewards, nil
}
