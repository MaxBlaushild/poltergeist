package server

import (
	"context"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const userOutOfBattleDamageOverTimeTickInterval = 5 * time.Minute

func normalizeUserStatusEffectType(raw string) models.UserStatusEffectType {
	switch models.UserStatusEffectType(strings.TrimSpace(strings.ToLower(raw))) {
	case models.UserStatusEffectTypeDamageOverTime:
		return models.UserStatusEffectTypeDamageOverTime
	case models.UserStatusEffectTypeHealthOverTime:
		return models.UserStatusEffectTypeHealthOverTime
	case models.UserStatusEffectTypeManaOverTime:
		return models.UserStatusEffectTypeManaOverTime
	default:
		return models.UserStatusEffectTypeStatModifier
	}
}

func normalizeMonsterStatusEffectType(raw string) models.MonsterStatusEffectType {
	switch models.MonsterStatusEffectType(strings.TrimSpace(strings.ToLower(raw))) {
	case models.MonsterStatusEffectTypeDamageOverTime:
		return models.MonsterStatusEffectTypeDamageOverTime
	case models.MonsterStatusEffectTypeHealthOverTime:
		return models.MonsterStatusEffectTypeHealthOverTime
	default:
		return models.MonsterStatusEffectTypeStatModifier
	}
}

func userStatusTickDeltas(status models.UserStatus) (healthDelta int, manaDelta int, applies bool) {
	switch normalizeUserStatusEffectType(string(status.EffectType)) {
	case models.UserStatusEffectTypeDamageOverTime:
		if status.DamagePerTick <= 0 {
			return 0, 0, false
		}
		return -status.DamagePerTick, 0, true
	case models.UserStatusEffectTypeHealthOverTime:
		if status.HealthPerTick == 0 {
			return 0, 0, false
		}
		return status.HealthPerTick, 0, true
	case models.UserStatusEffectTypeManaOverTime:
		if status.ManaPerTick == 0 {
			return 0, 0, false
		}
		return 0, status.ManaPerTick, true
	default:
		return 0, 0, false
	}
}

func battleStatusTickReady(startedAt time.Time, lastTickAt *time.Time, now time.Time) bool {
	lastTickAnchor := startedAt
	if lastTickAt != nil && lastTickAt.After(lastTickAnchor) {
		lastTickAnchor = *lastTickAt
	}
	return lastTickAnchor.Before(now)
}

func monsterStatusHealthTickDelta(status models.MonsterStatus) (healthDelta int, applies bool) {
	switch normalizeMonsterStatusEffectType(string(status.EffectType)) {
	case models.MonsterStatusEffectTypeDamageOverTime:
		if status.DamagePerTick <= 0 {
			return 0, false
		}
		return -status.DamagePerTick, true
	case models.MonsterStatusEffectTypeHealthOverTime:
		if status.HealthPerTick == 0 {
			return 0, false
		}
		return status.HealthPerTick, true
	default:
		return 0, false
	}
}

func (s *server) applyOutOfBattleUserDamageOverTime(ctx context.Context, userID uuid.UUID) error {
	hasActiveBattle, err := s.dbClient.MonsterBattle().HasAnyActiveForUser(ctx, userID)
	if err != nil {
		return err
	}
	if hasActiveBattle {
		return nil
	}

	statuses, err := s.dbClient.UserStatus().FindActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if len(statuses) == 0 {
		return nil
	}

	now := time.Now()
	totalHealthDelta := 0
	totalManaDelta := 0
	for _, status := range statuses {
		healthDelta, manaDelta, applies := userStatusTickDeltas(status)
		if !applies {
			continue
		}

		lastTickAnchor := status.StartedAt
		if status.LastTickAt != nil && status.LastTickAt.After(lastTickAnchor) {
			lastTickAnchor = *status.LastTickAt
		}
		if !lastTickAnchor.Before(now) {
			continue
		}

		ticks := int(now.Sub(lastTickAnchor) / userOutOfBattleDamageOverTimeTickInterval)
		if ticks <= 0 {
			continue
		}

		totalHealthDelta += ticks * healthDelta
		totalManaDelta += ticks * manaDelta
		newLastTickAt := lastTickAnchor.Add(time.Duration(ticks) * userOutOfBattleDamageOverTimeTickInterval)
		if err := s.dbClient.UserStatus().UpdateLastTickAt(ctx, status.ID, newLastTickAt); err != nil {
			return err
		}
	}

	if totalHealthDelta != 0 || totalManaDelta != 0 {
		if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, userID, -totalHealthDelta, -totalManaDelta); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) applyBattleTurnDamageOverTime(
	ctx context.Context,
	userID uuid.UUID,
	battleID uuid.UUID,
) (userDamage int, monsterDamage int, err error) {
	now := time.Now()

	userStatuses, err := s.dbClient.UserStatus().FindActiveByUserID(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	for _, status := range userStatuses {
		healthDelta, manaDelta, applies := userStatusTickDeltas(status)
		if !applies {
			continue
		}
		if !battleStatusTickReady(status.StartedAt, status.LastTickAt, now) {
			continue
		}
		if healthDelta < 0 {
			userDamage += -healthDelta
		}
		if err := s.dbClient.UserStatus().UpdateLastTickAt(ctx, status.ID, now); err != nil {
			return 0, 0, err
		}
		if manaDelta != 0 || healthDelta != 0 {
			if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, userID, -healthDelta, -manaDelta); err != nil {
				return 0, 0, err
			}
		}
	}

	monsterStatuses, err := s.dbClient.MonsterStatus().FindActiveByBattleID(ctx, battleID)
	if err != nil {
		return 0, 0, err
	}
	for _, status := range monsterStatuses {
		healthDelta, applies := monsterStatusHealthTickDelta(status)
		if !applies {
			continue
		}
		if !battleStatusTickReady(status.StartedAt, status.LastTickAt, now) {
			continue
		}
		if healthDelta < 0 {
			monsterDamage += -healthDelta
		}
		if err := s.dbClient.MonsterStatus().UpdateLastTickAt(ctx, status.ID, now); err != nil {
			return 0, 0, err
		}
		if healthDelta != 0 {
			if err := s.dbClient.MonsterBattle().AdjustMonsterHealthDeficit(ctx, battleID, -healthDelta); err != nil {
				return 0, 0, err
			}
		}
	}
	return userDamage, monsterDamage, nil
}
