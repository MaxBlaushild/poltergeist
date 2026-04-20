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

	chosen := chooseMonsterBattleAbility(
		monster,
		nil,
		20,
		100,
		999,
		monster.Level,
		time.Now(),
	)
	if chosen == nil || chosen.Name != "Recover" {
		t.Fatalf("expected low-health monster to prefer healing, got %+v", chosen)
	}
}

func TestChooseMonsterBattleAbilityBossDelaysHealingUntilEmergency(t *testing.T) {
	heal := models.Spell{
		ID:          uuid.New(),
		Name:        "Recover",
		AbilityType: models.SpellAbilityTypeSpell,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeRestoreLifePartyMember, Amount: 30},
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
			MonsterType: models.MonsterTemplateTypeBoss,
			Spells: []models.MonsterTemplateSpell{
				{Spell: heal},
				{Spell: damage},
			},
		},
	}

	chosen := chooseMonsterBattleAbility(
		monster,
		nil,
		25,
		100,
		999,
		monster.Level,
		time.Now(),
	)
	if chosen == nil || chosen.Name != "Bolt" {
		t.Fatalf("expected boss above emergency threshold to stay offensive, got %+v", chosen)
	}
}

func TestChooseMonsterBattleAbilityBossPrefersHealingAtEmergency(t *testing.T) {
	heal := models.Spell{
		ID:          uuid.New(),
		Name:        "Recover",
		AbilityType: models.SpellAbilityTypeSpell,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeRestoreLifePartyMember, Amount: 30},
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
			MonsterType: models.MonsterTemplateTypeBoss,
			Spells: []models.MonsterTemplateSpell{
				{Spell: heal},
				{Spell: damage},
			},
		},
	}

	chosen := chooseMonsterBattleAbility(
		monster,
		nil,
		20,
		100,
		999,
		monster.Level,
		time.Now(),
	)
	if chosen == nil || chosen.Name != "Recover" {
		t.Fatalf("expected boss at emergency health to heal, got %+v", chosen)
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

	chosen := chooseMonsterBattleAbility(
		monster,
		nil,
		90,
		100,
		10,
		monster.Level,
		time.Now(),
	)
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

	chosen := chooseMonsterBattleAbility(
		monster,
		battle,
		90,
		100,
		999,
		monster.Level,
		now,
	)
	if chosen == nil || chosen.Name != "Jab" {
		t.Fatalf("expected non-cooling-down ability to be chosen, got %+v", chosen)
	}
}

func TestAdjustedMonsterAbilityHealingForCombatCapsBossHealing(t *testing.T) {
	heal := &models.Spell{
		ID:          uuid.New(),
		Name:        "Recover",
		AbilityType: models.SpellAbilityTypeSpell,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeRestoreLifePartyMember, Amount: 120},
		},
	}
	monster := &models.Monster{
		Level: 20,
		Template: &models.MonsterTemplate{
			MonsterType: models.MonsterTemplateTypeBoss,
		},
	}
	battle := &models.MonsterBattle{
		MonsterHealthDeficit: 200,
	}

	got := adjustedMonsterAbilityHealingForCombat(monster, battle, heal, 300)
	if got != 54 {
		t.Fatalf("expected boss heal cap of 54, got %d", got)
	}
}

func TestApplyMonsterHealingLockoutAddsSharedCooldownToBossHeals(t *testing.T) {
	now := time.Now()
	healOne := models.Spell{
		ID:          uuid.New(),
		Name:        "Recover",
		AbilityType: models.SpellAbilityTypeSpell,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeRestoreLifePartyMember, Amount: 20},
		},
	}
	healTwo := models.Spell{
		ID:          uuid.New(),
		Name:        "Renew",
		AbilityType: models.SpellAbilityTypeSpell,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeRestoreLifeAllParty, Amount: 12},
		},
	}
	damage := models.Spell{
		ID:          uuid.New(),
		Name:        "Bolt",
		AbilityType: models.SpellAbilityTypeSpell,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeDealDamage, Amount: 10, Hits: 1},
		},
	}
	monster := &models.Monster{
		Level: 15,
		Template: &models.MonsterTemplate{
			MonsterType: models.MonsterTemplateTypeBoss,
			Spells: []models.MonsterTemplateSpell{
				{Spell: healOne},
				{Spell: healTwo},
				{Spell: damage},
			},
		},
	}

	cooldowns := applyMonsterHealingLockout(monster, models.MonsterBattleAbilityCooldowns{}, now)
	if turns := monsterCooldownTurnsRemaining(cooldowns, healOne.ID.String(), now); turns != monsterBattleBossHealLockTurns {
		t.Fatalf("expected first healing spell to be locked for %d turns, got %d", monsterBattleBossHealLockTurns, turns)
	}
	if turns := monsterCooldownTurnsRemaining(cooldowns, healTwo.ID.String(), now); turns != monsterBattleBossHealLockTurns {
		t.Fatalf("expected second healing spell to be locked for %d turns, got %d", monsterBattleBossHealLockTurns, turns)
	}
	if turns := monsterCooldownTurnsRemaining(cooldowns, damage.ID.String(), now); turns != 0 {
		t.Fatalf("expected offensive spell to remain usable, got cooldown=%d", turns)
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

	got := monsterAbilityDamageForCombat(monster, ability, monster.Level)
	if got <= 20 {
		t.Fatalf("expected technique damage to include level/strength bonuses, got %d", got)
	}
}

func TestMonsterCombatAbilitiesCapsAbilityTierToUserLevelPlusOne(t *testing.T) {
	early := models.Spell{
		ID:           uuid.New(),
		Name:         "Ember Snap",
		AbilityLevel: 5,
		AbilityType:  models.SpellAbilityTypeSpell,
	}
	late := models.Spell{
		ID:           uuid.New(),
		Name:         "Ember Nova",
		AbilityLevel: 9,
		AbilityType:  models.SpellAbilityTypeSpell,
	}
	monster := &models.Monster{
		Level: 12,
		Template: &models.MonsterTemplate{
			Spells: []models.MonsterTemplateSpell{
				{Spell: early},
				{Spell: late},
			},
		},
	}

	abilities := monsterCombatAbilitiesForUserLevel(monster, 4)
	if len(abilities) != 1 {
		t.Fatalf("expected exactly one ability after early-game cap, got %d", len(abilities))
	}
	if abilities[0].ID != early.ID {
		t.Fatalf("expected user level 4 to only expose %q, got %q", early.Name, abilities[0].Name)
	}
}

func TestMonsterCombatAbilitiesSkipsLockedDirectTemplateSpells(t *testing.T) {
	late := models.Spell{
		ID:           uuid.New(),
		Name:         "Shadow Burst",
		AbilityLevel: 10,
		AbilityType:  models.SpellAbilityTypeSpell,
	}
	monster := &models.Monster{
		Level: 3,
		Template: &models.MonsterTemplate{
			Spells: []models.MonsterTemplateSpell{
				{Spell: late},
			},
		},
	}

	abilities := monsterCombatAbilities(monster)
	if len(abilities) != 0 {
		t.Fatalf("expected locked direct template spell to stay unavailable, got %d abilities", len(abilities))
	}
}

func TestMonsterAbilityDamageForCombatAppliesEarlyGameDampener(t *testing.T) {
	monster := &models.Monster{
		Level: 12,
	}
	ability := &models.Spell{
		ID:          uuid.New(),
		AbilityType: models.SpellAbilityTypeSpell,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeDealDamage, Amount: 50, Hits: 1},
		},
	}

	got := monsterAbilityDamageForCombat(monster, ability, 4)
	if got != 30 {
		t.Fatalf("expected early-game dampener to reduce damage to 30, got %d", got)
	}
}

func TestCapMonsterAbilityDamageAgainstHealthProtectsLowLevelPlayers(t *testing.T) {
	monster := &models.Monster{
		Level: 12,
		Template: &models.MonsterTemplate{
			MonsterType: models.MonsterTemplateTypeBoss,
		},
	}
	bossBurst := &models.Spell{
		ID:            uuid.New(),
		AbilityType:   models.SpellAbilityTypeSpell,
		CooldownTurns: 2,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeDealDamage, Amount: 60, Hits: 1},
		},
	}
	aoe := &models.Spell{
		ID:          uuid.New(),
		AbilityType: models.SpellAbilityTypeSpell,
		Effects: models.SpellEffects{
			{Type: models.SpellEffectTypeDealDamageAllEnemies, Amount: 60, Hits: 1},
		},
	}

	if got := capMonsterAbilityDamageAgainstHealth(70, monster, bossBurst, 3, 100); got != 35 {
		t.Fatalf("expected early boss burst to cap at 35, got %d", got)
	}
	if got := capMonsterAbilityDamageAgainstHealth(70, monster, aoe, 3, 100); got != 12 {
		t.Fatalf("expected early AOE to cap at 12, got %d", got)
	}
}

func TestMonsterCombatAbilitiesUsesHighestProgressionMemberAtOrBelowLevel(t *testing.T) {
	low := models.Spell{
		ID:           uuid.New(),
		Name:         "Cinder Snap",
		AbilityLevel: 10,
		AbilityType:  models.SpellAbilityTypeSpell,
	}
	mid := models.Spell{
		ID:           uuid.New(),
		Name:         "Cinder Lance",
		AbilityLevel: 25,
		AbilityType:  models.SpellAbilityTypeSpell,
	}
	high := models.Spell{
		ID:           uuid.New(),
		Name:         "Cinder Nova",
		AbilityLevel: 50,
		AbilityType:  models.SpellAbilityTypeSpell,
	}
	monster := &models.Monster{
		Level: 24,
		Template: &models.MonsterTemplate{
			Progressions: []models.MonsterTemplateProgression{
				{
					Progression: models.SpellProgression{
						ID: uuid.New(),
						Members: []models.SpellProgressionSpell{
							{LevelBand: 10, Spell: low},
							{LevelBand: 25, Spell: mid},
							{LevelBand: 50, Spell: high},
						},
					},
				},
			},
		},
	}

	abilities := monsterCombatAbilities(monster)
	if len(abilities) != 1 {
		t.Fatalf("expected exactly one resolved progression ability, got %d", len(abilities))
	}
	if abilities[0].ID != low.ID {
		t.Fatalf("expected highest eligible level ability %q, got %q", low.Name, abilities[0].Name)
	}

	monster.Level = 25
	abilities = monsterCombatAbilities(monster)
	if len(abilities) != 1 {
		t.Fatalf("expected exactly one resolved progression ability at level 25, got %d", len(abilities))
	}
	if abilities[0].ID != mid.ID {
		t.Fatalf("expected level-25 monster to resolve to %q, got %q", mid.Name, abilities[0].Name)
	}
}

func TestMonsterCombatAbilitiesSkipsProgressionWhenAllMembersAreAboveLevel(t *testing.T) {
	low := models.Spell{
		ID:           uuid.New(),
		Name:         "Cinder Snap",
		AbilityLevel: 10,
		AbilityType:  models.SpellAbilityTypeSpell,
	}
	mid := models.Spell{
		ID:           uuid.New(),
		Name:         "Cinder Lance",
		AbilityLevel: 25,
		AbilityType:  models.SpellAbilityTypeSpell,
	}
	monster := &models.Monster{
		Level: 5,
		Template: &models.MonsterTemplate{
			Progressions: []models.MonsterTemplateProgression{
				{
					Progression: models.SpellProgression{
						ID: uuid.New(),
						Members: []models.SpellProgressionSpell{
							{LevelBand: 10, Spell: low},
							{LevelBand: 25, Spell: mid},
						},
					},
				},
			},
		},
	}

	abilities := monsterCombatAbilities(monster)
	if len(abilities) != 0 {
		t.Fatalf("expected no resolved progression abilities below level threshold, got %d", len(abilities))
	}
}
