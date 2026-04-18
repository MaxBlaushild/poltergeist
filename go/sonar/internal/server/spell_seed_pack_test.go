package server

import (
	"context"
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestCombatSpellSeedPackRequestsParse(t *testing.T) {
	srv := &server{}
	seen := map[string]struct{}{}
	genreID := uuid.NewString()
	for _, request := range combatSpellSeedPackRequests(genreID) {
		spell, err := srv.parseSpellUpsertRequest(context.Background(), request, 1, nil)
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
		if strings.TrimSpace(request.GenreID) != genreID {
			t.Fatalf("expected spell seed %q to carry genre %q, got %q", request.Name, genreID, request.GenreID)
		}
	}
}

func TestCombatTechniqueSeedPackRequestsParse(t *testing.T) {
	srv := &server{}
	seen := map[string]struct{}{}
	genreID := uuid.NewString()
	for _, request := range combatTechniqueSeedPackRequests(genreID) {
		spell, err := srv.parseSpellUpsertRequest(context.Background(), request, 1, nil)
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
		if strings.TrimSpace(request.GenreID) != genreID {
			t.Fatalf("expected technique seed %q to carry genre %q, got %q", request.Name, genreID, request.GenreID)
		}
	}
}
