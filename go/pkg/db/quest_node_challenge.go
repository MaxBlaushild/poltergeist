package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questNodeChallengeHandle struct {
	db *gorm.DB
}

func (h *questNodeChallengeHandle) Create(ctx context.Context, challenge *models.QuestNodeChallenge) error {
	return h.db.WithContext(ctx).Create(challenge).Error
}

func (h *questNodeChallengeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestNodeChallenge, error) {
	var challenge models.QuestNodeChallenge
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (h *questNodeChallengeHandle) FindByNodeID(ctx context.Context, nodeID uuid.UUID) ([]models.QuestNodeChallenge, error) {
	var challenges []models.QuestNodeChallenge
	if err := h.db.WithContext(ctx).Where("quest_node_id = ?", nodeID).Find(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (h *questNodeChallengeHandle) Update(ctx context.Context, id uuid.UUID, updates *models.QuestNodeChallenge) (*models.QuestNodeChallenge, error) {
	updatePayload := map[string]interface{}{
		"tier":              updates.Tier,
		"question":          updates.Question,
		"reward":            updates.Reward,
		"inventory_item_id": updates.InventoryItemID,
		"difficulty":        updates.Difficulty,
		"stat_tags":         updates.StatTags,
		"proficiency":       updates.Proficiency,
		"updated_at":        updates.UpdatedAt,
	}
	if err := h.db.WithContext(ctx).Model(&models.QuestNodeChallenge{}).Where("id = ?", id).Updates(updatePayload).Error; err != nil {
		return nil, err
	}
	return h.FindByID(ctx, id)
}

func (h *questNodeChallengeHandle) DeleteByNodeID(ctx context.Context, nodeID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("quest_node_id = ?", nodeID).Delete(&models.QuestNodeChallenge{}).Error
}
