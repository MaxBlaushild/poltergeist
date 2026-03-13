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
			StrongAgainstAffinity: &fire,
			WeakAgainstAffinity:   &ice,
		},
	}

	t.Run("strong against halves matching damage", func(t *testing.T) {
		damage, affinity, modifier := applyMonsterAffinityDamage(monster, 15, &fire)
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

	t.Run("strong against keeps minimum one damage", func(t *testing.T) {
		damage, _, modifier := applyMonsterAffinityDamage(monster, 1, &fire)
		if damage != 1 {
			t.Fatalf("expected minimum damage of 1, got %d", damage)
		}
		if modifier != monsterAffinityModifierStrongAgainst {
			t.Fatalf("expected strong against modifier, got %q", modifier)
		}
	})

	t.Run("weak against doubles matching damage", func(t *testing.T) {
		damage, affinity, modifier := applyMonsterAffinityDamage(monster, 15, &ice)
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
		damage, affinity, modifier := applyMonsterAffinityDamage(monster, 15, nil)
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
		damage, affinity, modifier := applyMonsterAffinityDamage(monster, 15, &shadow)
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

func TestParseMonsterTemplateUpsertRequestAffinities(t *testing.T) {
	s := &server{}

	t.Run("normalizes valid affinities", func(t *testing.T) {
		template, spells, err := s.parseMonsterTemplateUpsertRequest(nil, monsterTemplateUpsertRequest{
			Name:                  "Ash Warden",
			BaseStrength:          10,
			BaseDexterity:         10,
			BaseConstitution:      10,
			BaseIntelligence:      10,
			BaseWisdom:            10,
			BaseCharisma:          10,
			StrongAgainstAffinity: " FIRE ",
			WeakAgainstAffinity:   "ice",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(spells) != 0 {
			t.Fatalf("expected no spells, got %d", len(spells))
		}
		if template.StrongAgainstAffinity == nil || *template.StrongAgainstAffinity != "fire" {
			t.Fatalf("expected normalized strong affinity fire, got %v", template.StrongAgainstAffinity)
		}
		if template.WeakAgainstAffinity == nil || *template.WeakAgainstAffinity != "ice" {
			t.Fatalf("expected normalized weak affinity ice, got %v", template.WeakAgainstAffinity)
		}
	})

	t.Run("rejects duplicate strong and weak affinity", func(t *testing.T) {
		_, _, err := s.parseMonsterTemplateUpsertRequest(nil, monsterTemplateUpsertRequest{
			Name:                  "Ash Warden",
			BaseStrength:          10,
			BaseDexterity:         10,
			BaseConstitution:      10,
			BaseIntelligence:      10,
			BaseWisdom:            10,
			BaseCharisma:          10,
			StrongAgainstAffinity: "fire",
			WeakAgainstAffinity:   "fire",
		})
		if err == nil {
			t.Fatal("expected duplicate affinity validation error")
		}
	})

	t.Run("rejects invalid affinity", func(t *testing.T) {
		_, _, err := s.parseMonsterTemplateUpsertRequest(nil, monsterTemplateUpsertRequest{
			Name:                  "Ash Warden",
			BaseStrength:          10,
			BaseDexterity:         10,
			BaseConstitution:      10,
			BaseIntelligence:      10,
			BaseWisdom:            10,
			BaseCharisma:          10,
			StrongAgainstAffinity: "water",
		})
		if err == nil {
			t.Fatal("expected invalid affinity validation error")
		}
	})
}
