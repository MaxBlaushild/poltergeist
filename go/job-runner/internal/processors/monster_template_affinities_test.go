package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestDeriveMonsterTemplateAffinitiesHeuristicallyFireMonster(t *testing.T) {
	template := &models.MonsterTemplate{
		Name:             "Cinder Hound",
		Description:      "A flame-wreathed hunter that lunges through ash and ember.",
		BaseStrength:     14,
		BaseDexterity:    12,
		BaseConstitution: 13,
		BaseIntelligence: 6,
		BaseWisdom:       8,
		BaseCharisma:     7,
	}

	bonuses := deriveMonsterTemplateAffinitiesHeuristically(template)

	if bonuses.FireDamageBonusPercent <= 0 {
		t.Fatalf("expected fire damage bonus, got %+v", bonuses)
	}
	if bonuses.FireResistancePercent <= 0 {
		t.Fatalf("expected fire resistance, got %+v", bonuses)
	}
	if bonuses.IceResistancePercent >= 0 {
		t.Fatalf("expected ice vulnerability, got %+v", bonuses)
	}
}

func TestSanitizeMonsterTemplateAffinityPayloadClampsAndNormalizes(t *testing.T) {
	bonuses := sanitizeMonsterTemplateAffinityPayload(generatedMonsterTemplateAffinityPayload{
		AffinityDamageBonuses: map[string]int{
			"fire": 61,
		},
		AffinityResistances: map[string]int{
			"shadow": -53,
			"holy":   58,
		},
	})

	if bonuses.FireDamageBonusPercent != 60 {
		t.Fatalf("expected fire damage bonus to clamp to 60, got %d", bonuses.FireDamageBonusPercent)
	}
	if bonuses.ShadowResistancePercent != -50 {
		t.Fatalf("expected shadow resistance to clamp to -50, got %d", bonuses.ShadowResistancePercent)
	}
	if bonuses.HolyResistancePercent != 60 {
		t.Fatalf("expected holy resistance to round to 60, got %d", bonuses.HolyResistancePercent)
	}
}
