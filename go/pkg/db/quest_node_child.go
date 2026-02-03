package db

import (
  "context"

  "github.com/MaxBlaushild/poltergeist/pkg/models"
  "github.com/google/uuid"
)

type questNodeChildHandle struct {
  db *gorm.DB
}

func (h *questNodeChildHandle) Create(ctx context.Context, child *models.QuestNodeChild) error {
  return h.db.WithContext(ctx).Create(child).Error
}

func (h *questNodeChildHandle) FindByNodeID(ctx context.Context, nodeID uuid.UUID) ([]models.QuestNodeChild, error) {
  var children []models.QuestNodeChild
  if err := h.db.WithContext(ctx).Where("quest_node_id = ?", nodeID).Find(&children).Error; err != nil {
    return nil, err
  }
  return children, nil
}

func (h *questNodeChildHandle) DeleteByQuestID(ctx context.Context, questID uuid.UUID) error {
  return h.db.WithContext(ctx).
    Where("quest_node_id IN (SELECT id FROM quest_nodes WHERE quest_id = ?)", questID).
    Delete(&models.QuestNodeChild{}).Error
}

func (h *questNodeChildHandle) DeleteByNodeID(ctx context.Context, nodeID uuid.UUID) error {
  return h.db.WithContext(ctx).Where("quest_node_id = ?", nodeID).Delete(&models.QuestNodeChild{}).Error
}
