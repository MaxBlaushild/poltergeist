package server

import (
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestCooldownTurnsRemaining(t *testing.T) {
	now := time.Date(2026, time.March, 12, 12, 0, 0, 0, time.UTC)

	if got := cooldownTurnsRemaining(models.UserSpell{}, now); got != 0 {
		t.Fatalf("expected no cooldown, got %d", got)
	}

	oneSecond := now.Add(time.Second)
	if got := cooldownTurnsRemaining(models.UserSpell{CooldownExpiresAt: &oneSecond}, now); got != 1 {
		t.Fatalf("expected 1 turn remaining for 1 second, got %d", got)
	}

	fullTurn := now.Add(combatTurnDuration)
	if got := cooldownTurnsRemaining(models.UserSpell{CooldownExpiresAt: &fullTurn}, now); got != 1 {
		t.Fatalf("expected 1 turn remaining for full turn, got %d", got)
	}

	twoTurnsAndABit := now.Add((2 * combatTurnDuration) + time.Second)
	if got := cooldownTurnsRemaining(models.UserSpell{CooldownExpiresAt: &twoTurnsAndABit}, now); got != 3 {
		t.Fatalf("expected 3 turns remaining when cooldown spills into next turn, got %d", got)
	}
}
