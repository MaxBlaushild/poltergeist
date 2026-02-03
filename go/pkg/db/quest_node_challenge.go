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

func (h *questNodeChallengeHandle) FindByNodeID(ctx context.Context, nodeID uuid.UUID) ([]models.QuestNodeChallenge, error) {
  var challenges []models.QuestNodeChallenge
  if err := h.db.WithContext(ctx).Where("quest_node_id = ?", nodeID).Find(&challenges).Error; err != nil {
    return nil, err
  }
  return challenges, nil
}
