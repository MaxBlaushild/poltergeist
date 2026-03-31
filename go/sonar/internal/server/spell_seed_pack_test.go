package server

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestCombatSpellSeedPackRequestsParse(t *testing.T) {
	srv := &server{}
	seen := map[string]struct{}{}
	for _, request := range combatSpellSeedPackRequests() {
		spell, err := srv.parseSpellUpsertRequest(request, 1)
		if err != nil {
			t.Fatalf("expected spell seed %q to parse, got error: %v", request.Name, err)
		}
		key := strings.ToLower(strings.TrimSpace(spell.Name))
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate spell seed name %q", spell.Name)
		}
		seen[key] = struct{}{}
		if spell.AbilityType != models.SpellAbilityTypeSpell {
			t.Fatalf("expected %q to be a spell", spell.Name)
		}
	}
}

func TestCombatTechniqueSeedPackRequestsParse(t *testing.T) {
	srv := &server{}
	seen := map[string]struct{}{}
	for _, request := range combatTechniqueSeedPackRequests() {
		spell, err := srv.parseSpellUpsertRequest(request, 1)
		if err != nil {
			t.Fatalf("expected technique seed %q to parse, got error: %v", request.Name, err)
		}
		key := strings.ToLower(strings.TrimSpace(spell.Name))
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate technique seed name %q", spell.Name)
		}
		seen[key] = struct{}{}
		if spell.AbilityType != models.SpellAbilityTypeTechnique {
			t.Fatalf("expected %q to be a technique", spell.Name)
		}
		if spell.ManaCost != 0 {
			t.Fatalf("expected technique %q to have zero mana cost", spell.Name)
		}
	}
}
