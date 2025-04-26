package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questArchetypeNodeHandle struct {
	db *gorm.DB
}

func (h *questArchetypeNodeHandle) Create(ctx context.Context, questArchetypeNode *models.QuestArchetypeNode) error {
	return h.db.WithContext(ctx).Create(questArchetypeNode).Error
}

func (h *questArchetypeNodeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeNode, error) {
	var questArchetypeNode models.QuestArchetypeNode
	if err := h.db.WithContext(ctx).Preload("Challenges").Preload("LocationArchetype").First(&questArchetypeNode, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &questArchetypeNode, nil
}

func (h *questArchetypeNodeHandle) FindAll(ctx context.Context) ([]*models.QuestArchetypeNode, error) {
	var questArchetypeNodes []*models.QuestArchetypeNode
	if err := h.db.WithContext(ctx).Preload("Challenges").Preload("LocationArchetype").Find(&questArchetypeNodes).Error; err != nil {
		return nil, err
	}
	return questArchetypeNodes, nil
}

func (h *questArchetypeNodeHandle) Update(ctx context.Context, questArchetypeNode *models.QuestArchetypeNode) error {
	return h.db.WithContext(ctx).Save(questArchetypeNode).Error
}

func (h *questArchetypeNodeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.QuestArchetypeNode{}, "id = ?", id).Error
}
