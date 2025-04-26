package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questArchetypeChallengeHandle struct {
	db *gorm.DB
}

func (h *questArchetypeChallengeHandle) Create(ctx context.Context, questArchetypeChallenge *models.QuestArchetypeChallenge) error {
	return h.db.WithContext(ctx).Create(questArchetypeChallenge).Error
}

func (h *questArchetypeChallengeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeChallenge, error) {
	var questArchetypeChallenge models.QuestArchetypeChallenge
	if err := h.db.WithContext(ctx).Preload("UnlockedNode").First(&questArchetypeChallenge, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &questArchetypeChallenge, nil
}

func (h *questArchetypeChallengeHandle) FindAll(ctx context.Context) ([]*models.QuestArchetypeChallenge, error) {
	var questArchetypeChallenges []*models.QuestArchetypeChallenge
	if err := h.db.WithContext(ctx).Find(&questArchetypeChallenges).Error; err != nil {
		return nil, err
	}
	return questArchetypeChallenges, nil
}

func (h *questArchetypeChallengeHandle) Update(ctx context.Context, questArchetypeChallenge *models.QuestArchetypeChallenge) error {
	return h.db.WithContext(ctx).Save(questArchetypeChallenge).Error
}

func (h *questArchetypeChallengeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.QuestArchetypeChallenge{}, "id = ?", id).Error
}

func (h *questArchetypeChallengeHandle) FindAllByNodeID(ctx context.Context, nodeID uuid.UUID) ([]*models.QuestArchetypeChallenge, error) {
	var questArchetypeNodeChallenges []*models.QuestArchetypeNodeChallenge
	if err := h.db.WithContext(ctx).Preload("QuestArchetypeChallenge.UnlockedNode").Find(&questArchetypeNodeChallenges, "quest_archetype_node_id = ?", nodeID).Error; err != nil {
		return nil, err
	}

	if len(questArchetypeNodeChallenges) == 0 {
		return nil, nil
	}

	questArchetypeChallenges := make([]*models.QuestArchetypeChallenge, len(questArchetypeNodeChallenges))
	for i, nodeChallenge := range questArchetypeNodeChallenges {
		questArchetypeChallenges[i] = &nodeChallenge.QuestArchetypeChallenge
	}

	return questArchetypeChallenges, nil
}
