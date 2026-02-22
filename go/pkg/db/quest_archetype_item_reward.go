package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questArchetypeItemRewardHandle struct {
	db *gorm.DB
}

func (h *questArchetypeItemRewardHandle) ReplaceForQuestArchetype(ctx context.Context, questArchetypeID uuid.UUID, rewards []models.QuestArchetypeItemReward) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("quest_archetype_id = ?", questArchetypeID).Delete(&models.QuestArchetypeItemReward{}).Error; err != nil {
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

func (h *questArchetypeItemRewardHandle) FindByQuestArchetypeID(ctx context.Context, questArchetypeID uuid.UUID) ([]models.QuestArchetypeItemReward, error) {
	var rewards []models.QuestArchetypeItemReward
	if err := h.db.WithContext(ctx).
		Preload("InventoryItem").
		Where("quest_archetype_id = ?", questArchetypeID).
		Find(&rewards).Error; err != nil {
		return nil, err
	}
	return rewards, nil
}
