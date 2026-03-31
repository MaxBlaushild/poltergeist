package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestApplyMonsterAffinityDamage(t *testing.T) {
	fire := "fire"
	ice := "ice"
	monster := &models.Monster{
		Template: &models.MonsterTemplate{
			FireResistancePercent: 50,
			IceResistancePercent:  -100,
		},
	}

	t.Run("strong against halves matching damage", func(t *testing.T) {
		damage, affinity, modifier := applyMonsterAffinityDamage(monster, 15, &fire, models.CharacterStatBonuses{})
		if damage != 7 {
			t.Fatalf("expected halved damage of 7, got %d", damage)
		}
		if affinity == nil || *affinity != fire {
			t.Fatalf("expected normalized fire affinity, got %v", affinity)
		}
		if modifier != monsterAffinityModifierStrongAgainst {
			t.Fatalf("expected strong against modifier, got %q", modifier)
		}
	})

	t.Run("resistance can reduce tiny hits to zero", func(t *testing.T) {
		damage, _, modifier := applyMonsterAffinityDamage(monster, 1, &fire, models.CharacterStatBonuses{})
		if damage != 0 {
			t.Fatalf("expected damage of 0, got %d", damage)
		}
		if modifier != monsterAffinityModifierStrongAgainst {
			t.Fatalf("expected strong against modifier, got %q", modifier)
		}
	})

	t.Run("weak against doubles matching damage", func(t *testing.T) {
		damage, affinity, modifier := applyMonsterAffinityDamage(monster, 15, &ice, models.CharacterStatBonuses{})
		if damage != 30 {
			t.Fatalf("expected doubled damage of 30, got %d", damage)
		}
		if affinity == nil || *affinity != ice {
			t.Fatalf("expected normalized ice affinity, got %v", affinity)
		}
		if modifier != monsterAffinityModifierWeakAgainst {
			t.Fatalf("expected weak against modifier, got %q", modifier)
		}
	})

	t.Run("missing affinity leaves damage unchanged", func(t *testing.T) {
		damage, affinity, modifier := applyMonsterAffinityDamage(monster, 15, nil, models.CharacterStatBonuses{})
		if damage != 15 {
			t.Fatalf("expected unchanged damage of 15, got %d", damage)
		}
		if affinity != nil {
			t.Fatalf("expected nil affinity, got %v", affinity)
		}
		if modifier != monsterAffinityModifierNone {
			t.Fatalf("expected no modifier, got %q", modifier)
		}
	})

	t.Run("non matching affinity leaves damage unchanged", func(t *testing.T) {
		shadow := "shadow"
		damage, affinity, modifier := applyMonsterAffinityDamage(monster, 15, &shadow, models.CharacterStatBonuses{})
		if damage != 15 {
			t.Fatalf("expected unchanged damage of 15, got %d", damage)
		}
		if affinity == nil || *affinity != shadow {
			t.Fatalf("expected normalized shadow affinity, got %v", affinity)
		}
		if modifier != monsterAffinityModifierNone {
			t.Fatalf("expected no modifier, got %q", modifier)
		}
	})
}

func TestApplyCharacterAffinityResistance(t *testing.T) {
	fire := "fire"

	t.Run("positive resistance reduces damage", func(t *testing.T) {
		damage, affinity, resistance := applyCharacterAffinityResistance(
			100,
			&fire,
			models.CharacterStatBonuses{FireResistancePercent: 25},
		)
		if damage != 75 {
			t.Fatalf("expected 75 damage after resistance, got %d", damage)
		}
		if affinity == nil || *affinity != fire {
			t.Fatalf("expected normalized affinity fire, got %v", affinity)
		}
		if resistance != 25 {
			t.Fatalf("expected resistance 25, got %d", resistance)
		}
	})

	t.Run("negative resistance increases damage", func(t *testing.T) {
		damage, _, resistance := applyCharacterAffinityResistance(
			80,
			&fire,
			models.CharacterStatBonuses{FireResistancePercent: -25},
		)
		if damage != 100 {
			t.Fatalf("expected 100 damage after vulnerability, got %d", damage)
		}
		if resistance != -25 {
			t.Fatalf("expected resistance -25, got %d", resistance)
		}
	})
}

func TestApplyAffinityDamageBonus(t *testing.T) {
	fire := "fire"

	t.Run("positive bonus increases damage", func(t *testing.T) {
		damage, affinity, bonus := applyAffinityDamageBonus(
			100,
			&fire,
			models.CharacterStatBonuses{FireDamageBonusPercent: 25},
		)
		if damage != 125 {
			t.Fatalf("expected 125 damage after bonus, got %d", damage)
		}
		if affinity == nil || *affinity != fire {
			t.Fatalf("expected normalized affinity fire, got %v", affinity)
		}
		if bonus != 25 {
			t.Fatalf("expected bonus 25, got %d", bonus)
		}
	})

	t.Run("negative bonus reduces damage", func(t *testing.T) {
		damage, _, bonus := applyAffinityDamageBonus(
			80,
			&fire,
			models.CharacterStatBonuses{FireDamageBonusPercent: -25},
		)
		if damage != 60 {
			t.Fatalf("expected 60 damage after penalty, got %d", damage)
		}
		if bonus != -25 {
			t.Fatalf("expected bonus -25, got %d", bonus)
		}
	})
}

func TestParseMonsterTemplateUpsertRequestAffinities(t *testing.T) {
	s := &server{}

	t.Run("normalizes valid affinities", func(t *testing.T) {
		template, progressions, spells, err := s.parseMonsterTemplateUpsertRequest(nil, monsterTemplateUpsertRequest{
			Name:                        "Ash Warden",
			BaseStrength:                10,
			BaseDexterity:               10,
			BaseConstitution:            10,
			BaseIntelligence:            10,
			BaseWisdom:                  10,
			BaseCharisma:                10,
			FireResistancePercent:       50,
			IceDamageBonusPercent:       25,
			ShadowResistancePercent:     -100,
			LightningDamageBonusPercent: -20,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(spells) != 0 {
			t.Fatalf("expected no spells, got %d", len(spells))
		}
		if len(progressions) != 0 {
			t.Fatalf("expected no progressions, got %d", len(progressions))
		}
		if template.FireResistancePercent != 50 {
			t.Fatalf("expected fire resistance 50, got %d", template.FireResistancePercent)
		}
		if template.IceDamageBonusPercent != 25 {
			t.Fatalf("expected ice damage bonus 25, got %d", template.IceDamageBonusPercent)
		}
		if template.ShadowResistancePercent != -100 {
			t.Fatalf("expected shadow resistance -100, got %d", template.ShadowResistancePercent)
		}
		if template.LightningDamageBonusPercent != -20 {
			t.Fatalf("expected lightning damage bonus -20, got %d", template.LightningDamageBonusPercent)
		}
	})
}

func TestParseMonsterTemplateUpsertRequestNormalizesMonsterType(t *testing.T) {
	s := &server{}

	template, _, _, err := s.parseMonsterTemplateUpsertRequest(nil, monsterTemplateUpsertRequest{
		MonsterType:      " RAID ",
		Name:             "Catacomb Engine",
		BaseStrength:     10,
		BaseDexterity:    10,
		BaseConstitution: 10,
		BaseIntelligence: 10,
		BaseWisdom:       10,
		BaseCharisma:     10,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if template.MonsterType != models.MonsterTemplateTypeRaid {
		t.Fatalf("expected normalized raid monster type, got %q", template.MonsterType)
	}
}
