package db

import (
  "context"

  "github.com/MaxBlaushild/poltergeist/pkg/models"
  "github.com/google/uuid"
  "gorm.io/gorm"
)

type questHandle struct {
  db *gorm.DB
}

func (h *questHandle) Create(ctx context.Context, quest *models.Quest) error {
  return h.db.WithContext(ctx).Create(quest).Error
}

func (h *questHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Quest) error {
  return h.db.WithContext(ctx).Model(&models.Quest{}).Where("id = ?", id).Updates(updates).Error
}

func (h *questHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Quest, error) {
  var quest models.Quest
  if err := h.db.WithContext(ctx).
    Preload("Nodes").
    Preload("Nodes.Challenges").
    Preload("Nodes.Children").
    First(&quest, "id = ?", id).Error; err != nil {
    if err == gorm.ErrRecordNotFound {
      return nil, nil
    }
    return nil, err
  }
  return &quest, nil
}

func (h *questHandle) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Quest, error) {
  var quests []models.Quest
  if err := h.db.WithContext(ctx).
    Preload("Nodes").
    Preload("Nodes.Challenges").
    Preload("Nodes.Children").
    Where("id IN ?", ids).
    Find(&quests).Error; err != nil {
    return nil, err
  }
  return quests, nil
}

func (h *questHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Quest, error) {
  var quests []models.Quest
  if err := h.db.WithContext(ctx).
    Preload("Nodes").
    Preload("Nodes.Challenges").
    Preload("Nodes.Children").
    Where("zone_id = ?", zoneID).
    Find(&quests).Error; err != nil {
    return nil, err
  }
  return quests, nil
}

func (h *questHandle) FindAll(ctx context.Context) ([]models.Quest, error) {
  var quests []models.Quest
  if err := h.db.WithContext(ctx).
    Preload("Nodes").
    Preload("Nodes.Challenges").
    Preload("Nodes.Children").
    Find(&quests).Error; err != nil {
    return nil, err
  }
  return quests, nil
}
