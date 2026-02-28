package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userStatusHandler struct {
	db *gorm.DB
}

func (h *userStatusHandler) Create(ctx context.Context, status *models.UserStatus) error {
	now := time.Now()
	if status.ID == uuid.Nil {
		status.ID = uuid.New()
	}
	if status.CreatedAt.IsZero() {
		status.CreatedAt = now
	}
	status.UpdatedAt = now
	if status.StartedAt.IsZero() {
		status.StartedAt = now
	}
	if status.EffectType == "" {
		status.EffectType = models.UserStatusEffectTypeStatModifier
	}
	return h.db.WithContext(ctx).Create(status).Error
}

func (h *userStatusHandler) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserStatus, error) {
	var statuses []models.UserStatus
	now := time.Now()
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND started_at <= ? AND expires_at > ?", userID, now, now).
		Order("expires_at ASC, created_at ASC").
		Find(&statuses).Error; err != nil {
		return nil, err
	}
	return statuses, nil
}

func (h *userStatusHandler) GetActiveStatBonuses(ctx context.Context, userID uuid.UUID) (models.CharacterStatBonuses, error) {
	var bonuses models.CharacterStatBonuses
	now := time.Now()
	result := h.db.WithContext(ctx).
		Table("user_statuses").
		Select(`
			COALESCE(SUM(strength_mod), 0) AS strength,
			COALESCE(SUM(dexterity_mod), 0) AS dexterity,
			COALESCE(SUM(constitution_mod), 0) AS constitution,
			COALESCE(SUM(intelligence_mod), 0) AS intelligence,
			COALESCE(SUM(wisdom_mod), 0) AS wisdom,
			COALESCE(SUM(charisma_mod), 0) AS charisma
		`).
		Where("user_id = ? AND effect_type = ? AND started_at <= ? AND expires_at > ?", userID, models.UserStatusEffectTypeStatModifier, now, now).
		Scan(&bonuses)
	if result.Error != nil {
		return models.CharacterStatBonuses{}, result.Error
	}
	return bonuses, nil
}

func (h *userStatusHandler) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserStatus{}).Error
}
