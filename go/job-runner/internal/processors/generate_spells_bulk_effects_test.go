package processors

import (
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
