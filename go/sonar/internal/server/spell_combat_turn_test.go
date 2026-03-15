package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestSpellDealsMonsterDamage(t *testing.T) {
	t.Run("returns true for direct damage effects", func(t *testing.T) {
		spell := &models.Spell{
			Effects: models.SpellEffects{
				{
					Type:   models.SpellEffectTypeDealDamage,
					Amount: 12,
				},
				{
					Type: models.SpellEffectTypeApplyDetrimentalStatus,
				},
			},
		}

		if !spellDealsMonsterDamage(spell) {
			t.Fatal("expected mixed damage spell to be treated as monster-damaging")
		}
	})

	t.Run("returns false for status-only effects", func(t *testing.T) {
		spell := &models.Spell{
			Effects: models.SpellEffects{
				{
					Type: models.SpellEffectTypeApplyDetrimentalStatus,
				},
			},
		}

		if spellDealsMonsterDamage(spell) {
			t.Fatal("expected status-only spell to avoid monster damage path")
		}
	})
}
