package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const monsterBattleInactivityTimeout = 30 * time.Minute

type monsterBattleHandler struct {
	db *gorm.DB
}

func (h *monsterBattleHandler) Create(ctx context.Context, battle *models.MonsterBattle) error {
	now := time.Now()
	if battle.ID == uuid.Nil {
		battle.ID = uuid.New()
	}
	if battle.CreatedAt.IsZero() {
		battle.CreatedAt = now
	}
	battle.UpdatedAt = now
	if battle.StartedAt.IsZero() {
		battle.StartedAt = now
	}
	if battle.LastActivityAt.IsZero() {
		battle.LastActivityAt = now
	}
	return h.db.WithContext(ctx).Create(battle).Error
}

func (h *monsterBattleHandler) FindActiveByUserAndMonster(
	ctx context.Context,
	userID uuid.UUID,
	monsterID uuid.UUID,
) (*models.MonsterBattle, error) {
	var battle models.MonsterBattle
	err := h.db.WithContext(ctx).
		Where("user_id = ? AND monster_id = ? AND ended_at IS NULL", userID, monsterID).
		Order("started_at DESC").
		First(&battle).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	now := time.Now()
	if battle.LastActivityAt.Before(now.Add(-monsterBattleInactivityTimeout)) {
		if err := h.End(ctx, battle.ID, now); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return &battle, nil
}

func (h *monsterBattleHandler) Touch(ctx context.Context, battleID uuid.UUID, at time.Time) error {
	return h.db.WithContext(ctx).
		Model(&models.MonsterBattle{}).
		Where("id = ? AND ended_at IS NULL", battleID).
		Updates(map[string]interface{}{
			"last_activity_at": at,
			"updated_at":       at,
		}).Error
}

func (h *monsterBattleHandler) End(ctx context.Context, battleID uuid.UUID, endedAt time.Time) error {
	return h.db.WithContext(ctx).
		Model(&models.MonsterBattle{}).
		Where("id = ? AND ended_at IS NULL", battleID).
		Updates(map[string]interface{}{
			"ended_at":         endedAt,
			"last_activity_at": endedAt,
			"updated_at":       endedAt,
		}).Error
}
