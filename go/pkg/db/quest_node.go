package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questNodeHandle struct {
	db *gorm.DB
}

func (h *questNodeHandle) Create(ctx context.Context, node *models.QuestNode) error {
	db := h.db.WithContext(ctx)
	if node.Polygon == "" {
		db = db.Omit("polygon")
	}
	return db.Create(node).Error
}

func (h *questNodeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestNode, error) {
	var node models.QuestNode
	if err := h.db.WithContext(ctx).
		Preload("Challenges").
		Preload("Children").
		First(&node, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &node, nil
}

func (h *questNodeHandle) FindByQuestID(ctx context.Context, questID uuid.UUID) ([]models.QuestNode, error) {
	var nodes []models.QuestNode
	if err := h.db.WithContext(ctx).
		Preload("Challenges").
		Preload("Children").
		Where("quest_id = ?", questID).
		Order("order_index ASC").
		Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func (h *questNodeHandle) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.QuestNode{}, "id = ?", id).Error
}
