package server

import (
	"context"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const combatTurnDuration = 150 * time.Second

func cooldownTurnsRemaining(userSpell models.UserSpell, now time.Time) int {
	if userSpell.CooldownExpiresAt == nil {
		return 0
	}
	remaining := userSpell.CooldownExpiresAt.Sub(now)
	if remaining <= 0 {
		return 0
	}
	return int((remaining + combatTurnDuration - time.Nanosecond) / combatTurnDuration)
}

func cooldownSecondsRemaining(userSpell models.UserSpell, now time.Time) int {
	if userSpell.CooldownExpiresAt == nil {
		return 0
	}
	remaining := userSpell.CooldownExpiresAt.Sub(now)
	if remaining <= 0 {
		return 0
	}
	return int((remaining + time.Second - time.Nanosecond) / time.Second)
}

func cooldownExpiresAtFromTurns(turns int, now time.Time) *time.Time {
	if turns <= 0 {
		return nil
	}
	expiresAt := now.Add(time.Duration(turns) * combatTurnDuration)
	return &expiresAt
}

func monsterCooldownTurnsRemaining(
	cooldowns models.MonsterBattleAbilityCooldowns,
	abilityID string,
	now time.Time,
) int {
	if len(cooldowns) == 0 {
		return 0
	}
	expiresAt, exists := cooldowns[strings.TrimSpace(abilityID)]
	if !exists || expiresAt.IsZero() {
		return 0
	}
	remaining := expiresAt.Sub(now)
	if remaining <= 0 {
		return 0
	}
	return int((remaining + combatTurnDuration - time.Nanosecond) / combatTurnDuration)
}

func normalizeMonsterAbilityCooldowns(
	cooldowns models.MonsterBattleAbilityCooldowns,
	now time.Time,
) models.MonsterBattleAbilityCooldowns {
	if len(cooldowns) == 0 {
		return models.MonsterBattleAbilityCooldowns{}
	}
	normalized := make(models.MonsterBattleAbilityCooldowns, len(cooldowns))
	for rawID, expiresAt := range cooldowns {
		abilityID := strings.TrimSpace(rawID)
		if abilityID == "" || expiresAt.IsZero() || !expiresAt.After(now) {
			continue
		}
		normalized[abilityID] = expiresAt
	}
	return normalized
}

func (s *server) advanceUserCooldownsForCombatTurn(
	ctx context.Context,
	userID uuid.UUID,
	excludeSpellID *uuid.UUID,
	now time.Time,
) error {
	userSpells, err := s.dbClient.UserSpell().FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, userSpell := range userSpells {
		if excludeSpellID != nil && userSpell.SpellID == *excludeSpellID {
			continue
		}
		if cooldownTurnsRemaining(userSpell, now) <= 0 || userSpell.CooldownExpiresAt == nil {
			continue
		}
		nextExpiresAt := userSpell.CooldownExpiresAt.Add(-combatTurnDuration)
		var value *time.Time
		if nextExpiresAt.After(now) {
			value = &nextExpiresAt
		}
		if err := s.dbClient.UserSpell().UpdateCooldownExpiresAt(ctx, userID, userSpell.SpellID, value); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) advanceMonsterCooldownsForCombatTurn(
	ctx context.Context,
	battle *models.MonsterBattle,
	excludeAbilityID *uuid.UUID,
	now time.Time,
) error {
	if battle == nil {
		return nil
	}
	next := make(models.MonsterBattleAbilityCooldowns)
	current := normalizeMonsterAbilityCooldowns(battle.MonsterAbilityCooldowns, now)
	excludeID := ""
	if excludeAbilityID != nil {
		excludeID = excludeAbilityID.String()
	}
	changed := len(current) != len(battle.MonsterAbilityCooldowns)
	for abilityID, expiresAt := range current {
		if excludeID != "" && abilityID == excludeID {
			next[abilityID] = expiresAt
			continue
		}
		nextExpiresAt := expiresAt.Add(-combatTurnDuration)
		if nextExpiresAt.After(now) {
			next[abilityID] = nextExpiresAt
			if !nextExpiresAt.Equal(expiresAt) {
				changed = true
			}
			continue
		}
		changed = true
	}
	if !changed {
		battle.MonsterAbilityCooldowns = current
		return nil
	}
	battle.MonsterAbilityCooldowns = next
	return s.dbClient.MonsterBattle().UpdateMonsterCombatState(
		ctx,
		battle.ID,
		battle.MonsterManaDeficit,
		battle.MonsterAbilityCooldowns,
	)
}

func (s *server) advanceBattleStatusDurations(
	ctx context.Context,
	userID uuid.UUID,
	battleID uuid.UUID,
) error {
	shift := -combatTurnDuration
	if err := s.dbClient.UserStatus().ShiftActiveExpirations(ctx, userID, shift); err != nil {
		return err
	}
	if err := s.dbClient.MonsterStatus().ShiftActiveExpirations(ctx, battleID, shift); err != nil {
		return err
	}
	return nil
}
