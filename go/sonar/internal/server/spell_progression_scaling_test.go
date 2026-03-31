package server

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestSpellProgressionTargetAmountIncreasesWithBand(t *testing.T) {
	damage10 := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 10, models.SpellAbilityTypeSpell)
	damage25 := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 25, models.SpellAbilityTypeSpell)
	damage50 := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 50, models.SpellAbilityTypeSpell)
	damage70 := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 70, models.SpellAbilityTypeSpell)

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
		{level: 10, expected: 100},
		{level: 25, expected: 250},
		{level: 50, expected: 500},
		{level: 70, expected: 700},
	}

	for _, tc := range cases {
		actual := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, tc.level, models.SpellAbilityTypeSpell)
		if actual != tc.expected {
			t.Fatalf("expected damage target at level %d to be %d, got %d", tc.level, tc.expected, actual)
		}
	}
}

func TestSpellProgressionAllEnemiesDamageFollowsLevelBaseline(t *testing.T) {
	cases := []struct {
		level    int
		expected int
	}{
		{level: 10, expected: 60},
		{level: 25, expected: 150},
		{level: 50, expected: 300},
		{level: 70, expected: 420},
	}

	for _, tc := range cases {
		actual := spellProgressionTargetAmount(models.SpellEffectTypeDealDamageAllEnemies, tc.level, models.SpellAbilityTypeSpell)
		if actual != tc.expected {
			t.Fatalf("expected all-enemies damage target at level %d to be %d, got %d", tc.level, tc.expected, actual)
		}
	}
}

func TestSpellProgressionCombatAmountUsesMonsterHealthBaseline(t *testing.T) {
	level25 := scaleSpellProgressionCombatAmount(55, models.SpellEffectTypeDealDamage, 25, 25, models.SpellAbilityTypeSpell)
	level70 := scaleSpellProgressionCombatAmount(55, models.SpellEffectTypeDealDamage, 25, 70, models.SpellAbilityTypeSpell)

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
		models.SpellAbilityTypeSpell,
	)
	if level50FromSmallSeed < 400 {
		t.Fatalf(
			"expected level 50 damage to be anchored to the new combat baseline, got %d",
			level50FromSmallSeed,
		)
	}

	level70FromSmallSeed := scaleSpellProgressionCombatAmount(
		10,
		models.SpellEffectTypeDealDamage,
		25,
		70,
		models.SpellAbilityTypeSpell,
	)
	if level70FromSmallSeed < 650 {
		t.Fatalf(
			"expected level 70 damage to hit the more aggressive high-tier baseline, got %d",
			level70FromSmallSeed,
		)
	}
}

func TestSpellProgressionManaCostScalesByBand(t *testing.T) {
	level10 := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 10, models.SpellAbilityTypeSpell)
	level25 := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 25, models.SpellAbilityTypeSpell)
	level50 := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 50, models.SpellAbilityTypeSpell)
	level70 := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 70, models.SpellAbilityTypeSpell)

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
	if level70 < 400 {
		t.Fatalf("expected high-tier spell mana cost to be much steeper, got %d", level70)
	}
}

func TestSpellProgressionAllEnemiesManaCostExceedsSingleTarget(t *testing.T) {
	singleTarget := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 70, models.SpellAbilityTypeSpell)
	allEnemies := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamageAllEnemies, 25, 70, models.SpellAbilityTypeSpell)
	if allEnemies <= singleTarget {
		t.Fatalf(
			"expected all-enemies mana to exceed single-target mana at level 70, got aoe=%d single=%d",
			allEnemies,
			singleTarget,
		)
	}
}

func TestTechniqueProgressionUsesLowerDamageTargetsAndZeroMana(t *testing.T) {
	techniqueDamage := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 50, models.SpellAbilityTypeTechnique)
	spellDamage := spellProgressionTargetAmount(models.SpellEffectTypeDealDamage, 50, models.SpellAbilityTypeSpell)
	if techniqueDamage >= spellDamage {
		t.Fatalf("expected techniques to target lower damage than spells, got technique=%d spell=%d", techniqueDamage, spellDamage)
	}
	if techniqueDamage != 400 {
		t.Fatalf("expected techniques to use the new 8x baseline at level 50, got %d", techniqueDamage)
	}

	techniqueMana := scaleSpellProgressionManaCost(12, models.SpellEffectTypeDealDamage, 25, 70, models.SpellAbilityTypeTechnique)
	if techniqueMana != 0 {
		t.Fatalf("expected techniques to remain mana-free, got %d", techniqueMana)
	}
}

func TestSpellProgressionFlavorDescriptionStripsMetaReferences(t *testing.T) {
	seed := &models.Spell{
		Name:          "Inferno Blast",
		SchoolOfMagic: "Fire",
		Description:   "Level 50 evolution of Fire Wisp. Level 10 evolution of Inferno Blast. Unleash a searing wave of fire that engulfs your enemy.",
	}

	description := buildSpellProgressionFlavorDescription(seed, models.SpellEffectTypeDealDamage)
	lower := strings.ToLower(description)
	if strings.Contains(lower, "level ") || strings.Contains(lower, "evolution") || strings.Contains(lower, "progression") {
		t.Fatalf("expected progression meta references to be removed, got %q", description)
	}
	if description != "Unleash a searing wave of fire that engulfs your enemy." {
		t.Fatalf("expected only flavorful description to remain, got %q", description)
	}
}

func TestBuildSpellProgressionVariantUsesBandTargetLevel(t *testing.T) {
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

	variant := buildSpellProgressionVariant(seed, 25, 42, map[string]struct{}{}, nil, models.SpellAbilityTypeSpell)
	if variant.AbilityLevel != 50 {
		t.Fatalf("expected variant level to match the target band level, got %d", variant.AbilityLevel)
	}
}

func TestBuildSpellProgressionVariantUsesBandSpecificDescription(t *testing.T) {
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

	lowVariant := buildSpellProgressionVariant(seed, 25, 10, map[string]struct{}{}, nil, models.SpellAbilityTypeSpell)
	highVariant := buildSpellProgressionVariant(seed, 25, 70, map[string]struct{}{}, nil, models.SpellAbilityTypeSpell)

	if lowVariant.Description == highVariant.Description {
		t.Fatalf("expected progression descriptions to vary by band, got %q", lowVariant.Description)
	}
	if !strings.Contains(strings.ToLower(lowVariant.Description), "quick") {
		t.Fatalf("expected low-band description to read as lighter intensity, got %q", lowVariant.Description)
	}
	if !strings.Contains(strings.ToLower(highVariant.Description), "cataclysmic") {
		t.Fatalf("expected high-band description to read as higher intensity, got %q", highVariant.Description)
	}
}

func TestBuildTechniqueProgressionVariantUsesTechniqueRules(t *testing.T) {
	seed := &models.Spell{
		Name:          "Iron Current",
		AbilityType:   models.SpellAbilityTypeTechnique,
		SchoolOfMagic: "Martial",
		ManaCost:      0,
		CooldownTurns: 3,
		Effects: models.SpellEffects{
			{
				Type:   models.SpellEffectTypeDealDamage,
				Amount: 45,
			},
		},
	}

	variant := buildSpellProgressionVariant(seed, 25, 70, map[string]struct{}{}, nil, models.SpellAbilityTypeTechnique)
	if variant.AbilityType != models.SpellAbilityTypeTechnique {
		t.Fatalf("expected technique variant ability type, got %s", variant.AbilityType)
	}
	if variant.ManaCost != 0 {
		t.Fatalf("expected technique progression mana cost to remain 0, got %d", variant.ManaCost)
	}
	if variant.CooldownTurns != 3 {
		t.Fatalf("expected technique progression cooldown to match seed cooldown, got %d", variant.CooldownTurns)
	}
	if !strings.Contains(strings.ToLower(variant.Description), "cataclysmic") {
		t.Fatalf("expected technique description to retain band intensity, got %q", variant.Description)
	}
	if strings.Contains(strings.ToLower(variant.Name), "nova") {
		t.Fatalf("expected technique naming, got %q", variant.Name)
	}
}

func TestParseGeneratedSpellProgressionVariantFlavors(t *testing.T) {
	raw := "```json\n{\"variants\":[{\"levelBand\":10,\"name\":\"Kindled Arc\",\"description\":\"A quick ribbon of flame snaps across a foe.\"},{\"abilityLevel\":70,\"name\":\"Inferno Crown\",\"description\":\"A cataclysmic storm of fire breaks over the battlefield.\"}]}\n```"

	parsed, err := parseGeneratedSpellProgressionVariantFlavors(raw)
	if err != nil {
		t.Fatalf("expected parser to succeed, got %v", err)
	}
	if parsed[10].Name != "Kindled Arc" {
		t.Fatalf("expected level 10 name to parse, got %q", parsed[10].Name)
	}
	if parsed[70].Name != "Inferno Crown" {
		t.Fatalf("expected abilityLevel to normalize to level band 70, got %q", parsed[70].Name)
	}
	if strings.Contains(strings.ToLower(parsed[70].Description), "level") {
		t.Fatalf("expected description to be stripped of meta references, got %q", parsed[70].Description)
	}
}

func TestBuildSpellProgressionVariantUsesFlavorOverride(t *testing.T) {
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

	override := &generatedSpellProgressionVariantFlavor{
		LevelBand:   50,
		Name:        "Ashen Halo",
		Description: "A surging ring of fire crashes inward on a single foe.",
	}
	variant := buildSpellProgressionVariant(seed, 25, 50, map[string]struct{}{}, override, models.SpellAbilityTypeSpell)
	if variant.Name != "Ashen Halo" {
		t.Fatalf("expected override name to be used, got %q", variant.Name)
	}
	if variant.Description != "A surging ring of fire crashes inward on a single foe." {
		t.Fatalf("expected override description to be used, got %q", variant.Description)
	}
}
