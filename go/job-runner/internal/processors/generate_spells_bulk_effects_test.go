package processors

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestInferGeneratedAbilityEffectsTechniqueHasDamage(t *testing.T) {
	spec := jobs.SpellCreationSpec{
		Name:        "Iron Counter",
		Description: "A precise counter attack.",
	}

	effects := inferGeneratedAbilityEffects(spec, models.SpellAbilityTypeTechnique, 0)
	if len(effects) == 0 {
		t.Fatalf("expected at least one effect")
	}
	if effects[0].Type != models.SpellEffectTypeDealDamage {
		t.Fatalf("expected first effect to be deal_damage, got %s", effects[0].Type)
	}
	if effects[0].Amount <= 0 {
		t.Fatalf("expected positive damage amount, got %d", effects[0].Amount)
	}
	if effects[0].DamageAffinity == nil || strings.TrimSpace(*effects[0].DamageAffinity) == "" {
		t.Fatalf("expected damage affinity to be set for damage effect")
	}
}

func TestInferGeneratedAbilityEffectsHealingSpell(t *testing.T) {
	spec := jobs.SpellCreationSpec{
		Name:       "Verdant Renewal",
		EffectText: "Restore vitality to an ally.",
	}

	effects := inferGeneratedAbilityEffects(spec, models.SpellAbilityTypeSpell, 12)
	if len(effects) == 0 {
		t.Fatalf("expected at least one effect")
	}
	if effects[0].Type != models.SpellEffectTypeRestoreLifePartyMember {
		t.Fatalf("expected restore_life_party_member, got %s", effects[0].Type)
	}
	if effects[0].Amount <= 0 {
		t.Fatalf("expected positive heal amount, got %d", effects[0].Amount)
	}
}

func TestInferGeneratedAbilityEffectsDefensiveSpell(t *testing.T) {
	spec := jobs.SpellCreationSpec{
		Name:       "Rune Barrier",
		EffectText: "Raise a protective shield.",
	}

	effects := inferGeneratedAbilityEffects(spec, models.SpellAbilityTypeSpell, 18)
	if len(effects) == 0 {
		t.Fatalf("expected at least one effect")
	}
	if effects[0].Type != models.SpellEffectTypeApplyBeneficialStatus {
		t.Fatalf("expected apply_beneficial_statuses, got %s", effects[0].Type)
	}
	if len(effects[0].StatusesToApply) == 0 {
		t.Fatalf("expected beneficial statuses to be populated")
	}
}

func TestInferGeneratedAbilityEffectsWithPreferenceProducesMatchingText(t *testing.T) {
	spec := jobs.SpellCreationSpec{
		Name:        "Sanctuary Chorus",
		Description: "A soothing resonance that stabilizes the whole party.",
	}

	effects := inferGeneratedAbilityEffectsWithPreference(
		spec,
		models.SpellAbilityTypeSpell,
		18,
		models.SpellEffectTypeRestoreLifeAllParty,
		nil,
	)
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	if effects[0].Type != models.SpellEffectTypeRestoreLifeAllParty {
		t.Fatalf("expected restore_life_all_party_members, got %s", effects[0].Type)
	}

	text := strings.ToLower(buildGeneratedAbilityEffectText(effects, models.SpellAbilityTypeSpell))
	if !strings.Contains(text, "all allies") {
		t.Fatalf("expected effect text to describe all-allies healing, got %q", text)
	}
}

func TestInferGeneratedAbilityEffectsScalesWithTargetLevel(t *testing.T) {
	spec := jobs.SpellCreationSpec{
		Name:        "Arc Light",
		Description: "A direct offensive blast.",
	}

	lowLevel := 10
	highLevel := 90
	lowEffects := inferGeneratedAbilityEffectsWithPreference(
		spec,
		models.SpellAbilityTypeSpell,
		15,
		models.SpellEffectTypeDealDamage,
		&lowLevel,
	)
	highEffects := inferGeneratedAbilityEffectsWithPreference(
		spec,
		models.SpellAbilityTypeSpell,
		15,
		models.SpellEffectTypeDealDamage,
		&highLevel,
	)
	if len(lowEffects) == 0 || len(highEffects) == 0 {
		t.Fatalf("expected damage effects for both target levels")
	}
	if highEffects[0].Amount <= lowEffects[0].Amount {
		t.Fatalf("expected high-level amount > low-level amount, got low=%d high=%d", lowEffects[0].Amount, highEffects[0].Amount)
	}
}

func TestEstimateMonsterCurveForTargetLevelIncreasesWithLevel(t *testing.T) {
	low := 5
	high := 70
	lowCurve := estimateMonsterCurveForTargetLevel(&low)
	highCurve := estimateMonsterCurveForTargetLevel(&high)
	if lowCurve == nil || highCurve == nil {
		t.Fatalf("expected curve estimates for both levels")
	}
	if highCurve.EstimatedHealth <= lowCurve.EstimatedHealth {
		t.Fatalf(
			"expected higher-level estimated health > lower-level health, low=%d high=%d",
			lowCurve.EstimatedHealth,
			highCurve.EstimatedHealth,
		)
	}
	if highCurve.EstimatedDamagePerTurn <= lowCurve.EstimatedDamagePerTurn {
		t.Fatalf(
			"expected higher-level estimated dpt > lower-level dpt, low=%d high=%d",
			lowCurve.EstimatedDamagePerTurn,
			highCurve.EstimatedDamagePerTurn,
		)
	}
}

func TestInferGeneratedAbilityEffectsDamageIsCurveAppropriate(t *testing.T) {
	level := 50
	spec := jobs.SpellCreationSpec{
		Name:          "Storm Bolt",
		Description:   "A focused lightning blast.",
		SchoolOfMagic: "Tempest",
	}

	effects := inferGeneratedAbilityEffectsWithPreference(
		spec,
		models.SpellAbilityTypeSpell,
		18,
		models.SpellEffectTypeDealDamage,
		&level,
	)
	if len(effects) == 0 || effects[0].Type != models.SpellEffectTypeDealDamage {
		t.Fatalf("expected damage effect")
	}
	curve := estimateMonsterCurveForTargetLevel(&level)
	if curve == nil || curve.EstimatedHealth <= 0 {
		t.Fatalf("expected valid monster curve for level %d", level)
	}
	damageRatio := float64(effects[0].Amount) / float64(curve.EstimatedHealth)
	if damageRatio < 0.06 || damageRatio > 0.30 {
		t.Fatalf(
			"expected damage ratio to be level-appropriate, got ratio=%.3f damage=%d health=%d",
			damageRatio,
			effects[0].Amount,
			curve.EstimatedHealth,
		)
	}
}

func TestInferGeneratedAbilityEffectsHealingTracksMonsterThreatCurve(t *testing.T) {
	level := 50
	spec := jobs.SpellCreationSpec{
		Name:        "Renewing Cadence",
		Description: "A practiced restoration burst.",
	}

	effects := inferGeneratedAbilityEffectsWithPreference(
		spec,
		models.SpellAbilityTypeSpell,
		20,
		models.SpellEffectTypeRestoreLifePartyMember,
		&level,
	)
	if len(effects) == 0 || effects[0].Type != models.SpellEffectTypeRestoreLifePartyMember {
		t.Fatalf("expected single-target healing effect")
	}
	curve := estimateMonsterCurveForTargetLevel(&level)
	if curve == nil || curve.EstimatedDamagePerTurn <= 0 {
		t.Fatalf("expected valid monster curve dpt for level %d", level)
	}
	minExpected := int(float64(curve.EstimatedDamagePerTurn) * 0.5)
	if effects[0].Amount < minExpected {
		t.Fatalf(
			"expected heal amount to recover a meaningful share of threat curve: heal=%d min_expected=%d dpt=%d",
			effects[0].Amount,
			minExpected,
			curve.EstimatedDamagePerTurn,
		)
	}
}

func TestInferGeneratedAbilityEffectsInfersElementalAffinity(t *testing.T) {
	spec := jobs.SpellCreationSpec{
		Name:        "Inferno Spear",
		Description: "A spear of flame that pierces defenses.",
	}

	effects := inferGeneratedAbilityEffects(spec, models.SpellAbilityTypeSpell, 20)
	if len(effects) == 0 {
		t.Fatalf("expected at least one effect")
	}
	if effects[0].Type != models.SpellEffectTypeDealDamage {
		t.Fatalf("expected damage effect, got %s", effects[0].Type)
	}
	if effects[0].DamageAffinity == nil {
		t.Fatalf("expected damage affinity for damage effect")
	}
	if *effects[0].DamageAffinity != string(models.DamageAffinityFire) {
		t.Fatalf("expected fire affinity, got %q", *effects[0].DamageAffinity)
	}

	text := strings.ToLower(buildGeneratedAbilityEffectText(effects, models.SpellAbilityTypeSpell))
	if !strings.Contains(text, "fire damage") {
		t.Fatalf("expected effect text to include affinity, got %q", text)
	}
}

func TestBuildConfiguredAbilityEffectPlanRespectsCounts(t *testing.T) {
	plan := buildConfiguredAbilityEffectPlan(5, &jobs.SpellBulkEffectCounts{
		DealDamage:             3,
		RestoreLifePartyMember: 2,
	})
	if len(plan) != 5 {
		t.Fatalf("expected 5 planned effect types, got %d", len(plan))
	}

	damageCount := 0
	healCount := 0
	for _, effectType := range plan {
		switch effectType {
		case models.SpellEffectTypeDealDamage:
			damageCount++
		case models.SpellEffectTypeRestoreLifePartyMember:
			healCount++
		default:
			t.Fatalf("unexpected effect type in plan: %s", effectType)
		}
	}
	if damageCount != 3 {
		t.Fatalf("expected 3 damage slots, got %d", damageCount)
	}
	if healCount != 2 {
		t.Fatalf("expected 2 heal slots, got %d", healCount)
	}
}

func TestHarmonizeGeneratedAbilityNameWithEffectsKeepsMatchingName(t *testing.T) {
	name := harmonizeGeneratedAbilityNameWithEffects(
		"Ember Lance",
		models.SpellAbilityTypeSpell,
		models.SpellEffects{{
			Type:           models.SpellEffectTypeDealDamage,
			Amount:         18,
			DamageAffinity: testStringPtr(string(models.DamageAffinityFire)),
		}},
	)
	if name != "Ember Lance" {
		t.Fatalf("expected matching name to remain unchanged, got %q", name)
	}
}

func TestHarmonizeGeneratedAbilityNameWithEffectsRenamesForEffectAndAffinity(t *testing.T) {
	name := harmonizeGeneratedAbilityNameWithEffects(
		"Verdant Renewal",
		models.SpellAbilityTypeSpell,
		models.SpellEffects{{
			Type:           models.SpellEffectTypeDealDamage,
			Amount:         18,
			DamageAffinity: testStringPtr(string(models.DamageAffinityLightning)),
		}},
	)
	lower := strings.ToLower(name)
	if !containsAnyKeyword(lower, affinityKeywordsForName(string(models.DamageAffinityLightning))) {
		t.Fatalf("expected lightning affinity keyword in name, got %q", name)
	}
	if !containsAnyKeyword(lower, effectKeywordsForName(models.SpellEffectTypeDealDamage)) {
		t.Fatalf("expected lightning damage name alignment, got %q", name)
	}
}

func TestHarmonizeGeneratedAbilityNameWithEffectsRenamesForHealing(t *testing.T) {
	name := harmonizeGeneratedAbilityNameWithEffects(
		"Nightfall Rend",
		models.SpellAbilityTypeSpell,
		models.SpellEffects{{
			Type:   models.SpellEffectTypeRestoreLifePartyMember,
			Amount: 14,
		}},
	)
	lower := strings.ToLower(name)
	if !containsAnyKeyword(lower, effectKeywordsForName(models.SpellEffectTypeRestoreLifePartyMember)) {
		t.Fatalf("expected healing-aligned name, got %q", name)
	}
}

func TestReserveGeneratedAbilityNameAvoidsExistingAndInRunDupes(t *testing.T) {
	seen := map[string]struct{}{
		"ember lance": {},
	}
	first := reserveGeneratedAbilityName("Ember Lance", string(models.SpellAbilityTypeSpell), 1, seen)
	second := reserveGeneratedAbilityName("Ember Lance", string(models.SpellAbilityTypeSpell), 2, seen)

	if first != "Ember Lance 2" {
		t.Fatalf("expected first duplicate to be suffixed, got %q", first)
	}
	if second != "Ember Lance 3" {
		t.Fatalf("expected second duplicate to be suffixed, got %q", second)
	}
}

func TestHarmonizeGeneratedAbilityDescriptionWithEffectsKeepsMatchingDescription(t *testing.T) {
	description := harmonizeGeneratedAbilityDescriptionWithEffects(
		"Unleashes fire energy to damage a single enemy.",
		models.SpellAbilityTypeSpell,
		models.SpellEffects{{
			Type:           models.SpellEffectTypeDealDamage,
			Amount:         18,
			DamageAffinity: testStringPtr(string(models.DamageAffinityFire)),
		}},
	)
	if description != "Unleashes fire energy to damage a single enemy." {
		t.Fatalf("expected matching description to remain unchanged, got %q", description)
	}
}

func TestHarmonizeGeneratedAbilityDescriptionWithEffectsRenamesForHealing(t *testing.T) {
	description := harmonizeGeneratedAbilityDescriptionWithEffects(
		"A thunderous blast that shatters armor.",
		models.SpellAbilityTypeSpell,
		models.SpellEffects{{
			Type:   models.SpellEffectTypeRestoreLifePartyMember,
			Amount: 14,
		}},
	)
	lower := strings.ToLower(description)
	if !strings.Contains(lower, "heal") && !strings.Contains(lower, "restor") {
		t.Fatalf("expected healing-aligned description, got %q", description)
	}
}

func TestHarmonizeGeneratedAbilityDescriptionWithEffectsRenamesForAffinity(t *testing.T) {
	description := harmonizeGeneratedAbilityDescriptionWithEffects(
		"A gentle restorative pulse.",
		models.SpellAbilityTypeSpell,
		models.SpellEffects{{
			Type:           models.SpellEffectTypeDealDamage,
			Amount:         18,
			DamageAffinity: testStringPtr(string(models.DamageAffinityLightning)),
		}},
	)
	lower := strings.ToLower(description)
	if !strings.Contains(lower, "lightning") {
		t.Fatalf("expected affinity-aligned description, got %q", description)
	}
	if !strings.Contains(lower, "damage") {
		t.Fatalf("expected damage-aligned description, got %q", description)
	}
}

func TestInferGeneratedAbilityEffectsAllEnemiesDamage(t *testing.T) {
	level := 50
	spec := jobs.SpellCreationSpec{
		Name:          "Inferno Tempest",
		Description:   "A blazing wave that scorches all enemies at once.",
		SchoolOfMagic: "Pyromancy",
	}

	effects := inferGeneratedAbilityEffectsWithPreference(
		spec,
		models.SpellAbilityTypeSpell,
		28,
		models.SpellEffectTypeDealDamageAllEnemies,
		&level,
	)
	if len(effects) != 1 {
		t.Fatalf("expected one effect, got %d", len(effects))
	}
	if effects[0].Type != models.SpellEffectTypeDealDamageAllEnemies {
		t.Fatalf("expected deal_damage_all_enemies, got %s", effects[0].Type)
	}
	if effects[0].Amount <= 0 {
		t.Fatalf("expected positive AoE damage amount, got %d", effects[0].Amount)
	}
	if effects[0].DamageAffinity == nil || strings.TrimSpace(*effects[0].DamageAffinity) == "" {
		t.Fatalf("expected damage affinity for AoE damage")
	}

	text := strings.ToLower(buildGeneratedAbilityEffectText(effects, models.SpellAbilityTypeSpell))
	if !strings.Contains(text, "all enemies") {
		t.Fatalf("expected all-enemies effect text, got %q", text)
	}
}

func TestBuildConfiguredAbilityEffectPlanIncludesAllEnemyDamage(t *testing.T) {
	plan := buildConfiguredAbilityEffectPlan(4, &jobs.SpellBulkEffectCounts{
		DealDamage:           2,
		DealDamageAllEnemies: 2,
	})
	if len(plan) != 4 {
		t.Fatalf("expected 4 planned effect types, got %d", len(plan))
	}

	singleTargetCount := 0
	allEnemiesCount := 0
	for _, effectType := range plan {
		switch effectType {
		case models.SpellEffectTypeDealDamage:
			singleTargetCount++
		case models.SpellEffectTypeDealDamageAllEnemies:
			allEnemiesCount++
		default:
			t.Fatalf("unexpected effect type in plan: %s", effectType)
		}
	}
	if singleTargetCount != 2 {
		t.Fatalf("expected 2 single-target damage slots, got %d", singleTargetCount)
	}
	if allEnemiesCount != 2 {
		t.Fatalf("expected 2 all-enemies damage slots, got %d", allEnemiesCount)
	}
}

func testStringPtr(value string) *string {
	return &value
}
