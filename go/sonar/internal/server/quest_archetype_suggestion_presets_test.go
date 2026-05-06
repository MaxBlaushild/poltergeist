package server

import (
	"context"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestSanitizeQuestArchetypeSuggestionPresetResponseNormalizesGeneratedValues(t *testing.T) {
	s := &server{}
	marketID := uuid.New()
	bridgeID := uuid.New()
	locationArchetypes := []*models.LocationArchetype{
		{ID: marketID, Name: "Night Market"},
		{ID: bridgeID, Name: "Bridge Approach"},
	}

	preset, err := s.sanitizeQuestArchetypeSuggestionPresetResponse(
		context.Background(),
		questArchetypeSuggestionPresetLLMResponse{
			Count:       4,
			ZoneKind:    " Harbor District ",
			ThemePrompt: " Contraband routes are buckling under inspection pressure. ",
			FamilyTags:  []string{"civic", "trade"},
			FamilyMixTargets: map[string]int{
				"delivery":      5,
				"investigation": 4,
			},
			CharacterTags:                  []string{"courier", "fixer"},
			InternalTags:                   []string{"smuggling", "route_pressure"},
			RequiredLocationArchetypeNames: []string{" Night Market ", "Unknown", "Bridge Approach"},
			RequiredLocationMetadataTags:   []string{"market", "checkpoint"},
		},
		questArchetypeSuggestionPresetPayload{},
		locationArchetypes,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if preset.Count != 4 {
		t.Fatalf("expected generated count to be preserved, got %d", preset.Count)
	}
	if preset.ZoneKind != "harbor-district" {
		t.Fatalf("expected normalized zone kind, got %q", preset.ZoneKind)
	}
	if preset.ThemePrompt != "Contraband routes are buckling under inspection pressure. Make it unmistakably suited to harbor district zones, with routes, landmarks, and local tensions that only that environment could support." {
		t.Fatalf("expected trimmed theme prompt, got %q", preset.ThemePrompt)
	}
	if len(preset.RequiredLocationArchetypeIDs) != 2 {
		t.Fatalf("expected 2 matched location archetypes, got %d", len(preset.RequiredLocationArchetypeIDs))
	}
	if preset.RequiredLocationArchetypeIDs[0] != marketID.String() {
		t.Fatalf("expected first mapped archetype ID %q, got %q", marketID.String(), preset.RequiredLocationArchetypeIDs[0])
	}
	if preset.RequiredLocationArchetypeIDs[1] != bridgeID.String() {
		t.Fatalf("expected second mapped archetype ID %q, got %q", bridgeID.String(), preset.RequiredLocationArchetypeIDs[1])
	}
	totalFamilyMix := 0
	for _, count := range preset.FamilyMixTargets {
		totalFamilyMix += count
	}
	if totalFamilyMix != 4 {
		t.Fatalf("expected family mix targets to be trimmed to fit count, got %#v", preset.FamilyMixTargets)
	}
	if len(preset.FamilyMixTargets) != 1 {
		t.Fatalf("expected overflow family mix entries to be trimmed away, got %#v", preset.FamilyMixTargets)
	}
}

func TestSanitizeQuestArchetypeSuggestionPresetResponsePreservesHintedSmallCount(t *testing.T) {
	s := &server{}

	preset, err := s.sanitizeQuestArchetypeSuggestionPresetResponse(
		context.Background(),
		questArchetypeSuggestionPresetLLMResponse{
			Count: 10,
			FamilyMixTargets: map[string]int{
				"investigation": 2,
				"negotiation":   1,
			},
		},
		questArchetypeSuggestionPresetPayload{
			Count: 2,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if preset.Count != 2 {
		t.Fatalf("expected hinted small count to be preserved, got %d", preset.Count)
	}
	if preset.FamilyMixTargets["investigation"] != 2 {
		t.Fatalf("expected first family mix target to survive trim, got %#v", preset.FamilyMixTargets)
	}
	if preset.FamilyMixTargets["negotiation"] != 0 {
		t.Fatalf("expected overflow family mix target to be trimmed, got %#v", preset.FamilyMixTargets)
	}
}

func TestSanitizeQuestArchetypeSuggestionPresetResponseFallsBackToHints(t *testing.T) {
	s := &server{}
	courtyardID := uuid.New()
	locationArchetypes := []*models.LocationArchetype{
		{ID: courtyardID, Name: "Courtyard"},
	}

	preset, err := s.sanitizeQuestArchetypeSuggestionPresetResponse(
		context.Background(),
		questArchetypeSuggestionPresetLLMResponse{},
		questArchetypeSuggestionPresetPayload{
			Count:                        8,
			ZoneKind:                     "Temple Grounds",
			ThemePrompt:                  "Temple grounds processions and omen trails are pushing investigators toward a public ritual disruption.",
			FamilyTags:                   []string{"omens", "scholarly"},
			FamilyMixTargets:             map[string]int{"omen_chasing": 2, "ritual_interruption": 1},
			CharacterTags:                []string{"astrologer", "scribe"},
			InternalTags:                 []string{"star_charts", "prophecy"},
			RequiredLocationArchetypeIDs: []string{courtyardID.String()},
			RequiredLocationMetadataTags: []string{"courtyard", "monument"},
		},
		locationArchetypes,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if preset.Count != 8 {
		t.Fatalf("expected hinted count to survive, got %d", preset.Count)
	}
	if preset.ZoneKind != "temple-grounds" {
		t.Fatalf("expected normalized hinted zone kind, got %q", preset.ZoneKind)
	}
	if preset.ThemePrompt != "Temple grounds processions and omen trails are pushing investigators toward a public ritual disruption." {
		t.Fatalf("expected hinted theme prompt, got %q", preset.ThemePrompt)
	}
	if len(preset.FamilyMixTargets) != 2 || preset.FamilyMixTargets["omen_chasing"] != 2 {
		t.Fatalf("expected hinted family mix targets, got %#v", preset.FamilyMixTargets)
	}
	if len(preset.RequiredLocationArchetypeIDs) != 1 || preset.RequiredLocationArchetypeIDs[0] != courtyardID.String() {
		t.Fatalf("expected hinted location archetype ID, got %#v", preset.RequiredLocationArchetypeIDs)
	}
}

func TestSanitizeQuestArchetypeSuggestionPresetResponseEnrichesGenericHintThemeForZoneKind(t *testing.T) {
	s := &server{}

	preset, err := s.sanitizeQuestArchetypeSuggestionPresetResponse(
		context.Background(),
		questArchetypeSuggestionPresetLLMResponse{},
		questArchetypeSuggestionPresetPayload{
			Count:       2,
			ZoneKind:    "City",
			ThemePrompt: "A magical surge is pushing rival crews toward open conflict.",
		},
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if preset.ThemePrompt != "A magical surge is pushing rival crews toward open conflict. Make it unmistakably suited to city zones, with routes, landmarks, and local tensions that only that environment could support." {
		t.Fatalf("expected generic hint theme to be enriched with zone-kind flavor, got %q", preset.ThemePrompt)
	}
}

func TestSanitizeQuestArchetypeSuggestionPresetResponseAllowsNoRequiredLocationArchetypes(t *testing.T) {
	s := &server{}
	marketID := uuid.New()
	locationArchetypes := []*models.LocationArchetype{
		{ID: marketID, Name: "Night Market"},
	}

	preset, err := s.sanitizeQuestArchetypeSuggestionPresetResponse(
		context.Background(),
		questArchetypeSuggestionPresetLLMResponse{
			Count:                          2,
			ThemePrompt:                    "A flexible district mystery built around public rumors and shifting pressure.",
			RequiredLocationArchetypeNames: []string{},
		},
		questArchetypeSuggestionPresetPayload{},
		locationArchetypes,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(preset.RequiredLocationArchetypeIDs) != 0 {
		t.Fatalf("expected no required location archetypes, got %#v", preset.RequiredLocationArchetypeIDs)
	}
}
