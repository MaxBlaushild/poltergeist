package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestRebalanceSpellDamageEffectsToLegacyBaseline(t *testing.T) {
	fire := "fire"
	spell := models.Spell{
		AbilityType:  models.SpellAbilityTypeSpell,
		AbilityLevel: 70,
		Effects: models.SpellEffects{
			{
				Type:           models.SpellEffectTypeDealDamage,
				Amount:         700,
				Hits:           1,
				DamageAffinity: &fire,
			},
			{
				Type: models.SpellEffectTypeApplyDetrimentalStatus,
				StatusesToApply: models.ScenarioFailureStatusTemplates{
					{
						Name:          "Burning",
						DamagePerTick: 140,
					},
				},
			},
		},
	}

	effects, changed := rebalanceSpellDamageEffectsToLegacyBaseline(spell)
	if !changed {
		t.Fatal("expected rebalance helper to report changes")
	}
	if got := effects[0].Amount; got != 350 {
		t.Fatalf("expected single-target damage to reset to 350, got %d", got)
	}
	if got := effects[1].StatusesToApply[0].DamagePerTick; got != 70 {
		t.Fatalf("expected damage-per-tick to reset to 70, got %d", got)
	}
}

func TestRebalanceSpellDamageEffectsToLegacyBaselineHandlesAoe(t *testing.T) {
	lightning := "lightning"
	spell := models.Spell{
		AbilityType:  models.SpellAbilityTypeSpell,
		AbilityLevel: 50,
		Effects: models.SpellEffects{
			{
				Type:           models.SpellEffectTypeDealDamageAllEnemies,
				Amount:         300,
				Hits:           1,
				DamageAffinity: &lightning,
			},
		},
	}

	effects, changed := rebalanceSpellDamageEffectsToLegacyBaseline(spell)
	if !changed {
		t.Fatal("expected aoe spell to be rebalanced")
	}
	if got := effects[0].Amount; got != 200 {
		t.Fatalf("expected aoe damage to reset to 200, got %d", got)
	}
}

func TestRebalanceSpellDamageEffectsToLegacyBaselineLeavesLegacyValuesAlone(t *testing.T) {
	arcane := "arcane"
	spell := models.Spell{
		AbilityType:  models.SpellAbilityTypeSpell,
		AbilityLevel: 25,
		Effects: models.SpellEffects{
			{
				Type:           models.SpellEffectTypeDealDamage,
				Amount:         125,
				Hits:           1,
				DamageAffinity: &arcane,
			},
		},
	}

	effects, changed := rebalanceSpellDamageEffectsToLegacyBaseline(spell)
	if changed {
		t.Fatal("expected legacy-baseline spell to remain unchanged")
	}
	if got := effects[0].Amount; got != 125 {
		t.Fatalf("expected amount to stay at 125, got %d", got)
	}
}
