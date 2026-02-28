package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questSpellRewardHandle struct {
	db *gorm.DB
}

func (h *questSpellRewardHandle) ReplaceForQuest(ctx context.Context, questID uuid.UUID, rewards []models.QuestSpellReward) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("quest_id = ?", questID).Delete(&models.QuestSpellReward{}).Error; err != nil {
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
			rewards[i].QuestID = questID
			if rewards[i].CreatedAt.IsZero() {
				rewards[i].CreatedAt = now
			}
			rewards[i].UpdatedAt = now
		}
		return tx.Create(&rewards).Error
	})
}

func (h *questSpellRewardHandle) FindByQuestID(ctx context.Context, questID uuid.UUID) ([]models.QuestSpellReward, error) {
	var rewards []models.QuestSpellReward
	if err := h.db.WithContext(ctx).
		Preload("Spell").
		Where("quest_id = ?", questID).
		Find(&rewards).Error; err != nil {
		return nil, err
	}
	return rewards, nil
}
