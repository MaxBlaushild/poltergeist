package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questArchetypeNodeChallengeHandle struct {
	db *gorm.DB
}

func (h *questArchetypeNodeChallengeHandle) Create(ctx context.Context, questArchetypeNodeChallenge *models.QuestArchetypeNodeChallenge) error {
	return h.db.WithContext(ctx).Create(questArchetypeNodeChallenge).Error
}

func (h *questArchetypeNodeChallengeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeNodeChallenge, error) {
	var questArchetypeNodeChallenge models.QuestArchetypeNodeChallenge
	if err := h.db.WithContext(ctx).First(&questArchetypeNodeChallenge, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &questArchetypeNodeChallenge, nil
}

func (h *questArchetypeNodeChallengeHandle) FindAll(ctx context.Context) ([]*models.QuestArchetypeNodeChallenge, error) {
	var questArchetypeNodeChallenges []*models.QuestArchetypeNodeChallenge
	if err := h.db.WithContext(ctx).Find(&questArchetypeNodeChallenges).Error; err != nil {
		return nil, err
	}
	return questArchetypeNodeChallenges, nil
}

func (h *questArchetypeNodeChallengeHandle) Update(ctx context.Context, questArchetypeNodeChallenge *models.QuestArchetypeNodeChallenge) error {
	return h.db.WithContext(ctx).Save(questArchetypeNodeChallenge).Error
}

func (h *questArchetypeNodeChallengeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.QuestArchetypeNodeChallenge{}, "id = ?", id).Error
}
