package db

import (
  "context"

  "github.com/MaxBlaushild/poltergeist/pkg/models"
  "github.com/google/uuid"
  "gorm.io/gorm"
)

type trackedQuestHandle struct {
  db *gorm.DB
}

func (h *trackedQuestHandle) Create(ctx context.Context, questID uuid.UUID, userID uuid.UUID) error {
  var existing models.TrackedQuest
  if err := h.db.WithContext(ctx).
    Where("quest_id = ? AND user_id = ?", questID, userID).
    First(&existing).Error; err == nil {
    return nil
  }

  return h.db.WithContext(ctx).Create(&models.TrackedQuest{
    QuestID: questID,
    UserID:  userID,
  }).Error
}

func (h *trackedQuestHandle) Delete(ctx context.Context, questID uuid.UUID) error {
  return h.db.Unscoped().Delete(&models.TrackedQuest{}, "quest_id = ?", questID).Error
}

func (h *trackedQuestHandle) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.TrackedQuest, error) {
  var tracked []models.TrackedQuest
  if err := h.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tracked).Error; err != nil {
    return nil, err
  }
  return tracked, nil
}

func (h *trackedQuestHandle) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
  return h.db.Unscoped().Where("user_id = ?", userID).Delete(&models.TrackedQuest{}).Error
}
