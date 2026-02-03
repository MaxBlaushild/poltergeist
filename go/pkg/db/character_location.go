package db

import (
  "context"
  "time"

  "github.com/MaxBlaushild/poltergeist/pkg/models"
  "github.com/google/uuid"
  "gorm.io/gorm"
)

type characterLocationHandle struct {
  db *gorm.DB
}

func (h *characterLocationHandle) FindByCharacterID(ctx context.Context, characterID uuid.UUID) ([]*models.CharacterLocation, error) {
  var locations []*models.CharacterLocation
  if err := h.db.WithContext(ctx).
    Where("character_id = ?", characterID).
    Find(&locations).Error; err != nil {
    return nil, err
  }
  return locations, nil
}

func (h *characterLocationHandle) ReplaceForCharacter(ctx context.Context, characterID uuid.UUID, locations []models.CharacterLocation) error {
  tx := h.db.WithContext(ctx).Begin()
  if err := tx.Error; err != nil {
    return err
  }

  if err := tx.Where("character_id = ?", characterID).Delete(&models.CharacterLocation{}).Error; err != nil {
    tx.Rollback()
    return err
  }

  now := time.Now()
  for i := range locations {
    locations[i].ID = uuid.New()
    locations[i].CharacterID = characterID
    locations[i].CreatedAt = now
    locations[i].UpdatedAt = now
  }

  if len(locations) > 0 {
    if err := tx.Create(&locations).Error; err != nil {
      tx.Rollback()
      return err
    }
  }

  return tx.Commit().Error
}
