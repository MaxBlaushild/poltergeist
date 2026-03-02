package db

import (
	"context"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterStatusHandler struct {
	db *gorm.DB
}

func (h *monsterStatusHandler) Create(ctx context.Context, status *models.MonsterStatus) error {
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
		status.EffectType = models.MonsterStatusEffectTypeStatModifier
	}
	if status.DamagePerTick < 0 {
		status.DamagePerTick = 0
	}
	return h.db.WithContext(ctx).Create(status).Error
}

func (h *monsterStatusHandler) FindActiveByBattleID(
	ctx context.Context,
	battleID uuid.UUID,
) ([]models.MonsterStatus, error) {
	var statuses []models.MonsterStatus
	now := time.Now()
	if err := h.db.WithContext(ctx).
		Where("battle_id = ? AND started_at <= ? AND expires_at > ?", battleID, now, now).
		Order("expires_at ASC, created_at ASC").
		Find(&statuses).Error; err != nil {
		return nil, err
	}
	return statuses, nil
}

func (h *monsterStatusHandler) GetActiveStatBonuses(
	ctx context.Context,
	battleID uuid.UUID,
) (models.CharacterStatBonuses, error) {
	var bonuses models.CharacterStatBonuses
	now := time.Now()
	result := h.db.WithContext(ctx).
		Table("monster_statuses").
		Select(`
			COALESCE(SUM(strength_mod), 0) AS strength,
			COALESCE(SUM(dexterity_mod), 0) AS dexterity,
			COALESCE(SUM(constitution_mod), 0) AS constitution,
			COALESCE(SUM(intelligence_mod), 0) AS intelligence,
			COALESCE(SUM(wisdom_mod), 0) AS wisdom,
			COALESCE(SUM(charisma_mod), 0) AS charisma
		`).
		Where(
			"battle_id = ? AND effect_type = ? AND started_at <= ? AND expires_at > ?",
			battleID,
			models.MonsterStatusEffectTypeStatModifier,
			now,
			now,
		).
		Scan(&bonuses)
	if result.Error != nil {
		return models.CharacterStatBonuses{}, result.Error
	}
	return bonuses, nil
}

func (h *monsterStatusHandler) UpdateLastTickAt(
	ctx context.Context,
	statusID uuid.UUID,
	lastTickAt time.Time,
) error {
	return h.db.WithContext(ctx).
		Model(&models.MonsterStatus{}).
		Where("id = ?", statusID).
		Updates(map[string]interface{}{
			"last_tick_at": lastTickAt,
			"updated_at":   time.Now(),
		}).
		Error
}

func (h *monsterStatusHandler) DeleteActiveByBattleIDAndNames(
	ctx context.Context,
	battleID uuid.UUID,
	names []string,
) error {
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
		Where(
			"battle_id = ? AND started_at <= ? AND expires_at > ? AND lower(name) IN ?",
			battleID,
			now,
			now,
			normalized,
		).
		Delete(&models.MonsterStatus{}).
		Error
}

func (h *monsterStatusHandler) DeleteAllForBattleID(
	ctx context.Context,
	battleID uuid.UUID,
) error {
	return h.db.WithContext(ctx).
		Where("battle_id = ?", battleID).
		Delete(&models.MonsterStatus{}).
		Error
}
