package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestSpellProgressionTargetAmountIncreasesWithBand(t *testing.T) {
	damage10 := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 10)
	damage25 := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 25)
	damage50 := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 50)
	damage70 := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 70)

	if !(damage10 < damage25 && damage25 < damage50 && damage50 < damage70) {
		t.Fatalf(
			"expected strictly increasing damage targets by band, got 10=%d 25=%d 50=%d 70=%d",
			damage10,
			damage25,
			damage50,
			damage70,
		)
	}
}

func TestSpellProgressionDamageFollowsLevelBaseline(t *testing.T) {
	cases := []struct {
		level    int
		expected int
	}{
		{level: 10, expected: 50},
		{level: 25, expected: 125},
		{level: 50, expected: 250},
		{level: 70, expected: 350},
	}

	for _, tc := range cases {
		actual := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, tc.level)
		if actual != tc.expected {
			t.Fatalf("expected damage target at level %d to be %d, got %d", tc.level, tc.expected, actual)
		}
	}
}

func TestSpellProgressionCombatAmountUsesMonsterHealthBaseline(t *testing.T) {
	level25 := scaleSpellProgressionCombatAmount(55, models.SpellEffectTypeDealDamage, 25, 25)
	level70 := scaleSpellProgressionCombatAmount(55, models.SpellEffectTypeDealDamage, 25, 70)

	if level70 <= level25+80 {
		t.Fatalf(
			"expected level 70 to scale far beyond level 25 (health-aware), got level25=%d level70=%d",
			level25,
			level70,
		)
	}

	level50FromSmallSeed := scaleSpellProgressionCombatAmount(
		10,
		models.SpellEffectTypeDealDamage,
		25,
		50,
	)
	if level50FromSmallSeed < 180 {
		t.Fatalf(
			"expected level 50 damage to be anchored to monster HP baseline, got %d",
			level50FromSmallSeed,
		)
	}

	level70FromSmallSeed := scaleSpellProgressionCombatAmount(
		10,
		models.SpellEffectTypeDealDamage,
		25,
		70,
	)
	if level70FromSmallSeed < 320 {
		t.Fatalf(
			"expected level 70 damage to be high for high-health monsters, got %d",
			level70FromSmallSeed,
		)
	}
}

func TestSpellProgressionManaCostScalesByBand(t *testing.T) {
	level10 := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 10)
	level25 := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 25)
	level50 := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 50)
	level70 := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 70)

	if !(level10 < level25 && level25 < level50 && level50 < level70) {
		t.Fatalf(
			"expected mana cost to increase with band, got 10=%d 25=%d 50=%d 70=%d",
			level10,
			level25,
			level50,
			level70,
		)
	}
	if level70 < level25+30 {
		t.Fatalf(
			"expected level 70 mana to be materially higher than level 25, got 25=%d 70=%d",
			level25,
			level70,
		)
	}
	if level70 < 170 {
		t.Fatalf("expected high-tier spell mana cost to be substantial, got %d", level70)
	}
}
