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
	default:
		return models.UserStatusEffectTypeStatModifier
	}
}

func normalizeMonsterStatusEffectType(raw string) models.MonsterStatusEffectType {
	switch models.MonsterStatusEffectType(strings.TrimSpace(strings.ToLower(raw))) {
	case models.MonsterStatusEffectTypeDamageOverTime:
		return models.MonsterStatusEffectTypeDamageOverTime
	default:
		return models.MonsterStatusEffectTypeStatModifier
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
	totalDamage := 0
	for _, status := range statuses {
		if normalizeUserStatusEffectType(string(status.EffectType)) != models.UserStatusEffectTypeDamageOverTime {
			continue
		}
		if status.DamagePerTick <= 0 {
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

		totalDamage += ticks * status.DamagePerTick
		newLastTickAt := lastTickAnchor.Add(time.Duration(ticks) * userOutOfBattleDamageOverTimeTickInterval)
		if err := s.dbClient.UserStatus().UpdateLastTickAt(ctx, status.ID, newLastTickAt); err != nil {
			return err
		}
	}

	if totalDamage > 0 {
		if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, userID, totalDamage, 0); err != nil {
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
		if normalizeUserStatusEffectType(string(status.EffectType)) != models.UserStatusEffectTypeDamageOverTime {
			continue
		}
		if status.DamagePerTick <= 0 {
			continue
		}
		userDamage += status.DamagePerTick
		if err := s.dbClient.UserStatus().UpdateLastTickAt(ctx, status.ID, now); err != nil {
			return 0, 0, err
		}
	}

	monsterStatuses, err := s.dbClient.MonsterStatus().FindActiveByBattleID(ctx, battleID)
	if err != nil {
		return 0, 0, err
	}
	for _, status := range monsterStatuses {
		if normalizeMonsterStatusEffectType(string(status.EffectType)) != models.MonsterStatusEffectTypeDamageOverTime {
			continue
		}
		if status.DamagePerTick <= 0 {
			continue
		}
		monsterDamage += status.DamagePerTick
		if err := s.dbClient.MonsterStatus().UpdateLastTickAt(ctx, status.ID, now); err != nil {
			return 0, 0, err
		}
	}

	if userDamage > 0 {
		if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, userID, userDamage, 0); err != nil {
			return 0, 0, err
		}
	}
	if monsterDamage > 0 {
		if err := s.dbClient.MonsterBattle().AdjustMonsterHealthDeficit(ctx, battleID, monsterDamage); err != nil {
			return 0, 0, err
		}
	}
	return userDamage, monsterDamage, nil
}
