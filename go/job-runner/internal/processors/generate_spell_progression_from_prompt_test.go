package processors

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestPromptSpellProgressionTargetAmountIncreasesWithBand(t *testing.T) {
	damage10 := promptSpellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 10, models.SpellAbilityTypeSpell)
	damage25 := promptSpellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 25, models.SpellAbilityTypeSpell)
	damage50 := promptSpellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 50, models.SpellAbilityTypeSpell)
	damage70 := promptSpellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 70, models.SpellAbilityTypeSpell)

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

func TestPromptSpellProgressionDamageFollowsLevelBaselines(t *testing.T) {
	spellCases := []struct {
		level    int
		expected int
	}{
		{level: 10, expected: 50},
		{level: 25, expected: 125},
		{level: 50, expected: 250},
		{level: 70, expected: 350},
	}
	for _, tc := range spellCases {
		actual := promptSpellProgressionTargetAmount(
			models.SpellEffectTypeDealDamage,
			tc.level,
			models.SpellAbilityTypeSpell,
		)
		if actual != tc.expected {
			t.Fatalf("expected spell damage target at level %d to be %d, got %d", tc.level, tc.expected, actual)
		}
	}

	techniqueCases := []struct {
		level    int
		expected int
	}{
		{level: 10, expected: 40},
		{level: 25, expected: 100},
		{level: 50, expected: 200},
		{level: 70, expected: 280},
	}
	for _, tc := range techniqueCases {
		actual := promptSpellProgressionTargetAmount(
			models.SpellEffectTypeDealDamage,
			tc.level,
			models.SpellAbilityTypeTechnique,
		)
		if actual != tc.expected {
			t.Fatalf("expected technique damage target at level %d to be %d, got %d", tc.level, tc.expected, actual)
		}
	}
}

func TestPromptSpellProgressionAllEnemiesDamageFollowsLevelBaselines(t *testing.T) {
	spellCases := []struct {
		level    int
		expected int
	}{
		{level: 10, expected: 40},
		{level: 25, expected: 100},
		{level: 50, expected: 200},
		{level: 70, expected: 280},
	}
	for _, tc := range spellCases {
		actual := promptSpellProgressionTargetAmount(
			models.SpellEffectTypeDealDamageAllEnemies,
			tc.level,
			models.SpellAbilityTypeSpell,
		)
		if actual != tc.expected {
			t.Fatalf("expected AoE spell damage target at level %d to be %d, got %d", tc.level, tc.expected, actual)
		}
	}

	techniqueCases := []struct {
		level    int
		expected int
	}{
		{level: 10, expected: 30},
		{level: 25, expected: 75},
		{level: 50, expected: 150},
		{level: 70, expected: 210},
	}
	for _, tc := range techniqueCases {
		actual := promptSpellProgressionTargetAmount(
			models.SpellEffectTypeDealDamageAllEnemies,
			tc.level,
			models.SpellAbilityTypeTechnique,
		)
		if actual != tc.expected {
			t.Fatalf("expected AoE technique damage target at level %d to be %d, got %d", tc.level, tc.expected, actual)
		}
	}
}

func TestPromptSpellProgressionCombatAmountUsesMonsterHealthBaseline(t *testing.T) {
	level25 := promptScaleSpellProgressionCombatAmount(55, models.SpellEffectTypeDealDamage, 25, 25, models.SpellAbilityTypeSpell)
	level70 := promptScaleSpellProgressionCombatAmount(55, models.SpellEffectTypeDealDamage, 25, 70, models.SpellAbilityTypeSpell)

	if level70 <= level25+80 {
		t.Fatalf(
			"expected level 70 to scale far beyond level 25 (health-aware), got level25=%d level70=%d",
			level25,
			level70,
		)
	}

	level50FromSmallSeed := promptScaleSpellProgressionCombatAmount(
		10,
		models.SpellEffectTypeDealDamage,
		25,
		50,
		models.SpellAbilityTypeSpell,
	)
	if level50FromSmallSeed < 180 {
		t.Fatalf(
			"expected level 50 damage to be anchored to monster HP baseline, got %d",
			level50FromSmallSeed,
		)
	}

	level70FromSmallSeed := promptScaleSpellProgressionCombatAmount(
		10,
		models.SpellEffectTypeDealDamage,
		25,
		70,
		models.SpellAbilityTypeSpell,
	)
	if level70FromSmallSeed < 320 {
		t.Fatalf(
			"expected level 70 damage to be high for high-health monsters, got %d",
			level70FromSmallSeed,
		)
	}
}

func TestPromptSpellProgressionManaCostScalesByBand(t *testing.T) {
	level10 := promptScaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 10, models.SpellAbilityTypeSpell)
	level25 := promptScaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 25, models.SpellAbilityTypeSpell)
	level50 := promptScaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 50, models.SpellAbilityTypeSpell)
	level70 := promptScaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 70, models.SpellAbilityTypeSpell)

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

func TestPromptSpellProgressionAllEnemiesManaCostScalesHigherThanSingleTarget(t *testing.T) {
	singleTarget := promptScaleSpellProgressionManaCost(
		12,
		models.SpellEffectTypeDealDamage,
		25,
		70,
		models.SpellAbilityTypeSpell,
	)
	allEnemies := promptScaleSpellProgressionManaCost(
		12,
		models.SpellEffectTypeDealDamageAllEnemies,
		25,
		70,
		models.SpellAbilityTypeSpell,
	)
	if allEnemies <= singleTarget {
		t.Fatalf(
			"expected all-enemies mana to exceed single-target mana at level 70, got aoe=%d single=%d",
			allEnemies,
			singleTarget,
		)
	}
}

func TestPromptSpellProgressionTechniqueScalingIsLowerAndManaFree(t *testing.T) {
	spellDamage := promptScaleSpellProgressionCombatAmount(
		55,
		models.SpellEffectTypeDealDamage,
		25,
		70,
		models.SpellAbilityTypeSpell,
	)
	techniqueDamage := promptScaleSpellProgressionCombatAmount(
		55,
		models.SpellEffectTypeDealDamage,
		25,
		70,
		models.SpellAbilityTypeTechnique,
	)
	if techniqueDamage >= spellDamage {
		t.Fatalf(
			"expected technique damage to be lower than spell damage, got technique=%d spell=%d",
			techniqueDamage,
			spellDamage,
		)
	}

	techniqueMana := promptScaleSpellProgressionManaCost(
		12,
		models.SpellEffectTypeDealDamage,
		25,
		70,
		models.SpellAbilityTypeTechnique,
	)
	if techniqueMana != 0 {
		t.Fatalf("expected technique mana cost to stay at 0, got %d", techniqueMana)
	}
}

func TestPromptSpellProgressionFlavorDescriptionStripsMetaReferences(t *testing.T) {
	seed := &models.Spell{
		Name:          "Inferno Blast",
		SchoolOfMagic: "Fire",
		Description:   "Level 50 evolution of Fire Wisp. Level 10 evolution of Inferno Blast. Unleash a searing wave of fire that engulfs your enemy.",
	}

	description := buildPromptSpellProgressionFlavorDescription(
		seed,
		models.SpellEffectTypeDealDamage,
		models.SpellAbilityTypeSpell,
	)
	lower := strings.ToLower(description)
	if strings.Contains(lower, "level ") || strings.Contains(lower, "evolution") || strings.Contains(lower, "progression") {
		t.Fatalf("expected progression meta references to be removed, got %q", description)
	}
	if description != "Unleash a searing wave of fire that engulfs your enemy." {
		t.Fatalf("expected only flavorful description to remain, got %q", description)
	}
}

func TestBuildPromptSpellProgressionVariantUsesBandTargetLevel(t *testing.T) {
	seed := &models.Spell{
		Name:          "Inferno Blast",
		SchoolOfMagic: "Fire",
		ManaCost:      12,
		Effects: models.SpellEffects{
			{
				Type:   models.SpellEffectTypeDealDamage,
				Amount: 60,
			},
		},
	}

	variant := buildPromptSpellProgressionVariant(
		seed,
		25,
		42,
		map[string]struct{}{},
		models.SpellAbilityTypeSpell,
	)
	if variant.AbilityLevel != 50 {
		t.Fatalf("expected variant level to match the target band level, got %d", variant.AbilityLevel)
	}
}

func TestBuildPromptSpellProgressionVariantUsesBandSpecificDescription(t *testing.T) {
	seed := &models.Spell{
		Name:          "Inferno Blast",
		SchoolOfMagic: "Fire",
		ManaCost:      12,
		Effects: models.SpellEffects{
			{
				Type:   models.SpellEffectTypeDealDamage,
				Amount: 60,
			},
		},
	}

	lowVariant := buildPromptSpellProgressionVariant(
		seed,
		25,
		10,
		map[string]struct{}{},
		models.SpellAbilityTypeSpell,
	)
	highVariant := buildPromptSpellProgressionVariant(
		seed,
		25,
		70,
		map[string]struct{}{},
		models.SpellAbilityTypeSpell,
	)

	if lowVariant.Description == highVariant.Description {
		t.Fatalf("expected prompt progression descriptions to vary by band, got %q", lowVariant.Description)
	}
	if !strings.Contains(strings.ToLower(lowVariant.Description), "quick") {
		t.Fatalf("expected low-band description to read as lighter intensity, got %q", lowVariant.Description)
	}
	if !strings.Contains(strings.ToLower(highVariant.Description), "cataclysmic") {
		t.Fatalf("expected high-band description to read as higher intensity, got %q", highVariant.Description)
	}
}
