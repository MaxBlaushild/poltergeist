package server

import (
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestNextMonsterBattleTurnIndex(t *testing.T) {
	tests := []struct {
		name         string
		currentIndex int
		entryCount   int
		alive        []bool
		want         int
	}{
		{
			name:         "advances to next alive entry",
			currentIndex: 0,
			entryCount:   3,
			alive:        []bool{true, false, true},
			want:         2,
		},
		{
			name:         "wraps around to first alive entry",
			currentIndex: 2,
			entryCount:   3,
			alive:        []bool{true, false, true},
			want:         0,
		},
		{
			name:         "keeps current when nobody is alive",
			currentIndex: 1,
			entryCount:   3,
			alive:        []bool{false, false, false},
			want:         1,
		},
		{
			name:         "normalizes invalid current index",
			currentIndex: 99,
			entryCount:   3,
			alive:        []bool{false, true, true},
			want:         1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextMonsterBattleTurnIndex(
				tt.currentIndex,
				tt.entryCount,
				func(index int) bool {
					if index < 0 || index >= len(tt.alive) {
						return false
					}
					return tt.alive[index]
				},
			)
			if got != tt.want {
				t.Fatalf("nextMonsterBattleTurnIndex() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestChooseMonsterBattleAbilityPrefersSupportWhenLowHealth(t *testing.T) {
	heal := models.Spell{
		ID:          uuid.New(),
		Name:        "Recover",
		AbilityType: models.SpellAbilityTypeSpell,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeRestoreLifePartyMember, Amount: 12},
		},
	}
	damage := models.Spell{
		ID:          uuid.New(),
		Name:        "Bolt",
		AbilityType: models.SpellAbilityTypeSpell,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeDealDamage, Amount: 8, Hits: 1},
		},
	}
	monster := &models.Monster{
		Level: 12,
		Template: &models.MonsterTemplate{
			Spells: []models.MonsterTemplateSpell{
				{Spell: heal},
				{Spell: damage},
			},
		},
	}

	chosen := chooseMonsterBattleAbility(monster, nil, 20, 100, 999, time.Now())
	if chosen == nil || chosen.Name != "Recover" {
		t.Fatalf("expected low-health monster to prefer healing, got %+v", chosen)
	}
}

func TestChooseMonsterBattleAbilitySkipsUnaffordableSpell(t *testing.T) {
	spell := models.Spell{
		ID:          uuid.New(),
		Name:        "Meteor",
		AbilityType: models.SpellAbilityTypeSpell,
		ManaCost:    40,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeDealDamage, Amount: 18, Hits: 1},
		},
	}
	technique := models.Spell{
		ID:            uuid.New(),
		Name:          "Claw",
		AbilityType:   models.SpellAbilityTypeTechnique,
		CooldownTurns: 2,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeDealDamage, Amount: 8, Hits: 1},
		},
	}
	monster := &models.Monster{
		Level: 10,
		Template: &models.MonsterTemplate{
			Spells: []models.MonsterTemplateSpell{
				{Spell: spell},
				{Spell: technique},
			},
		},
	}

	chosen := chooseMonsterBattleAbility(monster, nil, 90, 100, 10, time.Now())
	if chosen == nil || chosen.Name != "Claw" {
		t.Fatalf("expected affordable ability to be chosen, got %+v", chosen)
	}
}

func TestChooseMonsterBattleAbilitySkipsCoolingDownAbility(t *testing.T) {
	now := time.Now()
	coolingDown := models.Spell{
		ID:            uuid.New(),
		Name:          "Smash",
		AbilityType:   models.SpellAbilityTypeTechnique,
		CooldownTurns: 2,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeDealDamage, Amount: 14, Hits: 1},
		},
	}
	backup := models.Spell{
		ID:          uuid.New(),
		Name:        "Jab",
		AbilityType: models.SpellAbilityTypeTechnique,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeDealDamage, Amount: 5, Hits: 1},
		},
	}
	monster := &models.Monster{
		Level: 10,
		Template: &models.MonsterTemplate{
			Spells: []models.MonsterTemplateSpell{
				{Spell: coolingDown},
				{Spell: backup},
			},
		},
	}
	battle := &models.MonsterBattle{
		MonsterAbilityCooldowns: models.MonsterBattleAbilityCooldowns{
			coolingDown.ID.String(): now.Add(2 * combatTurnDuration),
		},
	}

	chosen := chooseMonsterBattleAbility(monster, battle, 90, 100, 999, now)
	if chosen == nil || chosen.Name != "Jab" {
		t.Fatalf("expected non-cooling-down ability to be chosen, got %+v", chosen)
	}
}

func TestMonsterAbilityDamageForCombatTechniqueAddsStrengthBonus(t *testing.T) {
	monster := &models.Monster{
		Level: 15,
		Template: &models.MonsterTemplate{
			BaseStrength: 18,
		},
	}
	ability := &models.Spell{
		ID:          uuid.New(),
		AbilityType: models.SpellAbilityTypeTechnique,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeDealDamage, Amount: 10, Hits: 2},
		},
	}

	got := monsterAbilityDamageForCombat(monster, ability)
	if got <= 20 {
		t.Fatalf("expected technique damage to include level/strength bonuses, got %d", got)
	}
}
