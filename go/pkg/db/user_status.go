package db

import (
	"context"
	"strings"
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
	if status.DamagePerTick < 0 {
		status.DamagePerTick = 0
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
			COALESCE(SUM(charisma_mod), 0) AS charisma,
			COALESCE(SUM(physical_damage_bonus_percent), 0) AS physical_damage_bonus_percent,
			COALESCE(SUM(piercing_damage_bonus_percent), 0) AS piercing_damage_bonus_percent,
			COALESCE(SUM(slashing_damage_bonus_percent), 0) AS slashing_damage_bonus_percent,
			COALESCE(SUM(bludgeoning_damage_bonus_percent), 0) AS bludgeoning_damage_bonus_percent,
			COALESCE(SUM(fire_damage_bonus_percent), 0) AS fire_damage_bonus_percent,
			COALESCE(SUM(ice_damage_bonus_percent), 0) AS ice_damage_bonus_percent,
			COALESCE(SUM(lightning_damage_bonus_percent), 0) AS lightning_damage_bonus_percent,
			COALESCE(SUM(poison_damage_bonus_percent), 0) AS poison_damage_bonus_percent,
			COALESCE(SUM(arcane_damage_bonus_percent), 0) AS arcane_damage_bonus_percent,
			COALESCE(SUM(holy_damage_bonus_percent), 0) AS holy_damage_bonus_percent,
			COALESCE(SUM(shadow_damage_bonus_percent), 0) AS shadow_damage_bonus_percent,
			COALESCE(SUM(physical_resistance_percent), 0) AS physical_resistance_percent,
			COALESCE(SUM(piercing_resistance_percent), 0) AS piercing_resistance_percent,
			COALESCE(SUM(slashing_resistance_percent), 0) AS slashing_resistance_percent,
			COALESCE(SUM(bludgeoning_resistance_percent), 0) AS bludgeoning_resistance_percent,
			COALESCE(SUM(fire_resistance_percent), 0) AS fire_resistance_percent,
			COALESCE(SUM(ice_resistance_percent), 0) AS ice_resistance_percent,
			COALESCE(SUM(lightning_resistance_percent), 0) AS lightning_resistance_percent,
			COALESCE(SUM(poison_resistance_percent), 0) AS poison_resistance_percent,
			COALESCE(SUM(arcane_resistance_percent), 0) AS arcane_resistance_percent,
			COALESCE(SUM(holy_resistance_percent), 0) AS holy_resistance_percent,
			COALESCE(SUM(shadow_resistance_percent), 0) AS shadow_resistance_percent
		`).
		Where("user_id = ? AND effect_type = ? AND started_at <= ? AND expires_at > ?", userID, models.UserStatusEffectTypeStatModifier, now, now).
		Scan(&bonuses)
	if result.Error != nil {
		return models.CharacterStatBonuses{}, result.Error
	}
	return bonuses, nil
}

func (h *userStatusHandler) UpdateLastTickAt(
	ctx context.Context,
	statusID uuid.UUID,
	lastTickAt time.Time,
) error {
	return h.db.WithContext(ctx).
		Model(&models.UserStatus{}).
		Where("id = ?", statusID).
		Updates(map[string]interface{}{
			"last_tick_at": lastTickAt,
			"updated_at":   time.Now(),
		}).
		Error
}

func (h *userStatusHandler) ShiftActiveExpirations(
	ctx context.Context,
	userID uuid.UUID,
	shift time.Duration,
) error {
	seconds := int(shift / time.Second)
	if seconds == 0 {
		return nil
	}
	now := time.Now()
	return h.db.WithContext(ctx).
		Model(&models.UserStatus{}).
		Where("user_id = ? AND started_at <= ? AND expires_at > ?", userID, now, now).
		Updates(map[string]interface{}{
			"expires_at": gorm.Expr("expires_at + (? * interval '1 second')", seconds),
			"updated_at": now,
		}).
		Error
}

func (h *userStatusHandler) DeleteActiveByUserIDAndNames(ctx context.Context, userID uuid.UUID, names []string) error {
	normalized := make([]string, 0, len(names))
	seen := map[string]struct{}{}
	for _, name := range names {
		clean := strings.ToLower(strings.TrimSpace(name))
		if clean == "" {
			continue
		}
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		normalized = append(normalized, clean)
	}
	if len(normalized) == 0 {
		return nil
	}

	now := time.Now()
	return h.db.WithContext(ctx).
		Where("user_id = ? AND started_at <= ? AND expires_at > ? AND lower(name) IN ?", userID, now, now, normalized).
		Delete(&models.UserStatus{}).
		Error
}

func (h *userStatusHandler) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserStatus{}).Error
}
