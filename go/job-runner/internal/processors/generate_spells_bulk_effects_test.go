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
