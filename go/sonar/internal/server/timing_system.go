package server

import (
	"context"
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

func cooldownExpiresAtFromTurns(turns int, now time.Time) *time.Time {
	if turns <= 0 {
		return nil
	}
	expiresAt := now.Add(time.Duration(turns) * combatTurnDuration)
	return &expiresAt
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
