package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userStoryFlagHandle struct {
	db *gorm.DB
}

func (h *userStoryFlagHandle) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserStoryFlag, error) {
	var flags []models.UserStoryFlag
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&flags).Error; err != nil {
		return nil, err
	}
	return flags, nil
}

func (h *userStoryFlagHandle) Upsert(ctx context.Context, userID uuid.UUID, flagKey string, value bool) error {
	now := time.Now()
	record := models.UserStoryFlag{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    userID,
		FlagKey:   flagKey,
		Value:     value,
	}
	return h.db.WithContext(ctx).
		Where("user_id = ? AND flag_key = ?", userID, flagKey).
		Assign(map[string]interface{}{
			"value":      value,
			"updated_at": now,
		}).
		FirstOrCreate(&record).Error
}

func (h *userStoryFlagHandle) DeleteByUserAndKeys(ctx context.Context, userID uuid.UUID, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	return h.db.WithContext(ctx).
		Where("user_id = ? AND flag_key IN ?", userID, keys).
		Delete(&models.UserStoryFlag{}).Error
}
