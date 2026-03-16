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
	if battle.State == "" {
		battle.State = string(models.MonsterBattleStateActive)
	}
	if battle.TurnIndex < 0 {
		battle.TurnIndex = 0
	}
	if battle.MonsterManaDeficit < 0 {
		battle.MonsterManaDeficit = 0
	}
	if battle.MonsterAbilityCooldowns == nil {
		battle.MonsterAbilityCooldowns = models.MonsterBattleAbilityCooldowns{}
	}
	if battle.LastActionSequence < 0 {
		battle.LastActionSequence = 0
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

func (h *monsterBattleHandler) HasAnyActiveForUser(
	ctx context.Context,
	userID uuid.UUID,
) (bool, error) {
	var count int64
	now := time.Now()
	if err := h.db.WithContext(ctx).
		Model(&models.MonsterBattle{}).
		Where(
			"(user_id = ? OR id IN (SELECT battle_id FROM monster_battle_participants WHERE user_id = ?)) AND ended_at IS NULL AND last_activity_at >= ?",
			userID,
			userID,
			now.Add(-monsterBattleInactivityTimeout),
		).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (h *monsterBattleHandler) FindByID(
	ctx context.Context,
	battleID uuid.UUID,
) (*models.MonsterBattle, error) {
	var battle models.MonsterBattle
	if err := h.db.WithContext(ctx).Where("id = ?", battleID).First(&battle).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &battle, nil
}

func (h *monsterBattleHandler) FindActiveByParticipantAndMonster(
	ctx context.Context,
	userID uuid.UUID,
	monsterID uuid.UUID,
) (*models.MonsterBattle, error) {
	var battle models.MonsterBattle
	err := h.db.WithContext(ctx).
		Model(&models.MonsterBattle{}).
		Joins("JOIN monster_battle_participants mbp ON mbp.battle_id = monster_battles.id").
		Where("mbp.user_id = ? AND monster_battles.monster_id = ? AND monster_battles.ended_at IS NULL", userID, monsterID).
		Order("monster_battles.started_at DESC").
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

func (h *monsterBattleHandler) AdjustMonsterHealthDeficit(
	ctx context.Context,
	battleID uuid.UUID,
	delta int,
) error {
	if delta == 0 {
		return nil
	}
	now := time.Now()
	return h.db.WithContext(ctx).
		Model(&models.MonsterBattle{}).
		Where("id = ? AND ended_at IS NULL", battleID).
		Updates(map[string]interface{}{
			"monster_health_deficit": gorm.Expr("GREATEST(0, monster_health_deficit + ?)", delta),
			"last_activity_at":       now,
			"updated_at":             now,
		}).Error
}

func (h *monsterBattleHandler) UpdateMonsterCombatState(
	ctx context.Context,
	battleID uuid.UUID,
	manaDeficit int,
	cooldowns models.MonsterBattleAbilityCooldowns,
) error {
	if manaDeficit < 0 {
		manaDeficit = 0
	}
	if cooldowns == nil {
		cooldowns = models.MonsterBattleAbilityCooldowns{}
	}
	now := time.Now()
	return h.db.WithContext(ctx).
		Model(&models.MonsterBattle{}).
		Where("id = ? AND ended_at IS NULL", battleID).
		Updates(map[string]interface{}{
			"monster_mana_deficit":      manaDeficit,
			"monster_ability_cooldowns": cooldowns,
			"updated_at":                now,
		}).Error
}

func (h *monsterBattleHandler) RecordLastAction(
	ctx context.Context,
	battleID uuid.UUID,
	action models.MonsterBattleLastAction,
) error {
	now := time.Now()
	return h.db.WithContext(ctx).
		Model(&models.MonsterBattle{}).
		Where("id = ? AND ended_at IS NULL", battleID).
		Updates(map[string]interface{}{
			"last_action_sequence": gorm.Expr("last_action_sequence + 1"),
			"last_action":          action,
			"last_activity_at":     now,
			"updated_at":           now,
		}).Error
}

func (h *monsterBattleHandler) SetState(
	ctx context.Context,
	battleID uuid.UUID,
	state string,
) error {
	now := time.Now()
	return h.db.WithContext(ctx).
		Model(&models.MonsterBattle{}).
		Where("id = ? AND ended_at IS NULL", battleID).
		Updates(map[string]interface{}{
			"state":      state,
			"updated_at": now,
		}).Error
}

func (h *monsterBattleHandler) SetTurnIndex(
	ctx context.Context,
	battleID uuid.UUID,
	turnIndex int,
) error {
	if turnIndex < 0 {
		turnIndex = 0
	}
	now := time.Now()
	return h.db.WithContext(ctx).
		Model(&models.MonsterBattle{}).
		Where("id = ? AND ended_at IS NULL", battleID).
		Updates(map[string]interface{}{
			"turn_index": turnIndex,
			"updated_at": now,
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
