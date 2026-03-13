package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestParseSpellEffectsCoercesDetrimentalStatusesNegative(t *testing.T) {
	s := &server{}
	positive := true

	for _, effectType := range []models.SpellEffectType{
		models.SpellEffectTypeApplyDetrimentalStatus,
		models.SpellEffectTypeApplyDetrimentalAll,
	} {
		effects, err := s.parseSpellEffects([]spellEffectPayload{
			{
				Type: string(effectType),
				StatusesToApply: []scenarioFailureStatusPayload{
					{
						Name:            "Weakened",
						Description:     "Reduced output.",
						Effect:          "Lowers strength.",
						Positive:        &positive,
						DurationSeconds: 30,
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("parseSpellEffects(%s) returned error: %v", effectType, err)
		}
		if len(effects) != 1 || len(effects[0].StatusesToApply) != 1 {
			t.Fatalf("parseSpellEffects(%s) returned unexpected effect payload: %+v", effectType, effects)
		}
		if effects[0].StatusesToApply[0].Positive {
			t.Fatalf("expected %s to coerce applied status to detrimental", effectType)
		}
	}
}

func TestSpellHasCastableEffectRecognizesDetrimentalStatusEffects(t *testing.T) {
	for _, effectType := range []models.SpellEffectType{
		models.SpellEffectTypeApplyDetrimentalStatus,
		models.SpellEffectTypeApplyDetrimentalAll,
	} {
		spell := &models.Spell{
			Effects: models.SpellEffects{
				{Type: effectType},
			},
		}
		if !spellHasCastableEffect(spell) {
			t.Fatalf("expected %s to be castable", effectType)
		}
	}
}
