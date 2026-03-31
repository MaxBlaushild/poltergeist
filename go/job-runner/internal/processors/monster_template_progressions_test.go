package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestChooseProgressionsForMonsterTemplatePrefersTechniqueForPhysicalTemplate(t *testing.T) {
	template := &models.MonsterTemplate{
		MonsterType:                models.MonsterTemplateTypeMonster,
		Name:                       "Iron Mauler",
		BaseStrength:               18,
		BaseDexterity:              14,
		BaseConstitution:           15,
		BaseIntelligence:           6,
		BaseWisdom:                 7,
		BaseCharisma:               5,
		SlashingDamageBonusPercent: 25,
	}

	techniqueID := uuid.New()
	spellID := uuid.New()
	selected := chooseProgressionsForMonsterTemplate(
		template,
		[]monsterTemplateProgressionCandidate{
			{
				ProgressionID:  techniqueID,
				Name:           "Butcher's Sequence",
				AbilityType:    models.SpellAbilityTypeTechnique,
				MemberCount:    4,
				DirectDamage:   true,
				AffinityCounts: map[models.DamageAffinity]int{models.DamageAffinitySlashing: 2},
			},
			{
				ProgressionID:  spellID,
				Name:           "Ash Canticle",
				AbilityType:    models.SpellAbilityTypeSpell,
				MemberCount:    4,
				DirectDamage:   true,
				AffinityCounts: map[models.DamageAffinity]int{models.DamageAffinityFire: 2},
			},
		},
	)

	if len(selected) == 0 {
		t.Fatalf("expected at least one selected progression")
	}
	if selected[0].ProgressionID != techniqueID {
		t.Fatalf("expected first progression to be technique %s, got %s", techniqueID, selected[0].ProgressionID)
	}
}

func TestChooseProgressionsForMonsterTemplatePrefersMatchingAffinityForCaster(t *testing.T) {
	template := &models.MonsterTemplate{
		MonsterType:              models.MonsterTemplateTypeBoss,
		Name:                     "Cinder Oracle",
		BaseStrength:             6,
		BaseDexterity:            8,
		BaseConstitution:         9,
		BaseIntelligence:         18,
		BaseWisdom:               16,
		BaseCharisma:             14,
		FireDamageBonusPercent:   35,
		FireResistancePercent:    20,
		ArcaneDamageBonusPercent: 8,
	}

	fireID := uuid.New()
	shadowID := uuid.New()
	selected := chooseProgressionsForMonsterTemplate(
		template,
		[]monsterTemplateProgressionCandidate{
			{
				ProgressionID:  shadowID,
				Name:           "Night Gospel",
				AbilityType:    models.SpellAbilityTypeSpell,
				MemberCount:    4,
				DirectDamage:   true,
				AffinityCounts: map[models.DamageAffinity]int{models.DamageAffinityShadow: 3},
			},
			{
				ProgressionID:  fireID,
				Name:           "Cinder Hymnal",
				AbilityType:    models.SpellAbilityTypeSpell,
				MemberCount:    4,
				DirectDamage:   true,
				AreaDamage:     true,
				AffinityCounts: map[models.DamageAffinity]int{models.DamageAffinityFire: 3},
			},
		},
	)

	if len(selected) == 0 {
		t.Fatalf("expected at least one selected progression")
	}
	if selected[0].ProgressionID != fireID {
		t.Fatalf("expected first progression to be fire-aligned %s, got %s", fireID, selected[0].ProgressionID)
	}
}
