package db

import (
  "context"
  "time"

  "github.com/MaxBlaushild/poltergeist/pkg/models"
  "github.com/google/uuid"
  "gorm.io/gorm"
)

type questAcceptanceV2Handle struct {
  db *gorm.DB
}

func (h *questAcceptanceV2Handle) Create(ctx context.Context, acceptance *models.QuestAcceptanceV2) error {
  return h.db.WithContext(ctx).Create(acceptance).Error
}

func (h *questAcceptanceV2Handle) FindByUserAndQuest(ctx context.Context, userID uuid.UUID, questID uuid.UUID) (*models.QuestAcceptanceV2, error) {
  var acceptance models.QuestAcceptanceV2
  if err := h.db.WithContext(ctx).
    Where("user_id = ? AND quest_id = ?", userID, questID).
    First(&acceptance).Error; err != nil {
    if err == gorm.ErrRecordNotFound {
      return nil, nil
    }
    return nil, err
  }
  return &acceptance, nil
}

func (h *questAcceptanceV2Handle) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.QuestAcceptanceV2, error) {
  var acceptances []models.QuestAcceptanceV2
  if err := h.db.WithContext(ctx).
    Where("user_id = ?", userID).
    Find(&acceptances).Error; err != nil {
    return nil, err
  }
  return acceptances, nil
}

func (h *questAcceptanceV2Handle) MarkTurnedIn(ctx context.Context, id uuid.UUID) error {
  now := time.Now()
  return h.db.WithContext(ctx).
    Model(&models.QuestAcceptanceV2{}).
    Where("id = ?", id).
    Updates(map[string]interface{}{"	turned_in_at": now, "updated_at": now}).Error
}
