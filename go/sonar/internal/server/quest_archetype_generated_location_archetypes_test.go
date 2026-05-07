package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestSanitizeQuestArchetypeGeneratedLocationArchetypePayloadNormalizesTypes(t *testing.T) {
	spec, ok := sanitizeQuestArchetypeGeneratedLocationArchetypePayload(
		generatedQuestArchetypeLocationArchetypePayload{
			Name:          " Cozy Book Nook ",
			IncludedTypes: []string{"book_store", "cafe", "book store"},
			ExcludedTypes: []string{"night club", "cafe"},
		},
	)
	if !ok {
		t.Fatalf("expected generated location archetype payload to sanitize")
	}
	if spec.Name != "Cozy Book Nook" {
		t.Fatalf("expected normalized name, got %q", spec.Name)
	}
	if len(spec.IncludedTypes) != 2 ||
		spec.IncludedTypes[0] != googlemaps.TypeBookStore ||
		spec.IncludedTypes[1] != googlemaps.TypeCafe {
		t.Fatalf("expected canonical included types, got %+v", spec.IncludedTypes)
	}
	if len(spec.ExcludedTypes) != 1 || spec.ExcludedTypes[0] != googlemaps.TypeNightClub {
		t.Fatalf("expected excluded overlap with included types to be removed, got %+v", spec.ExcludedTypes)
	}
}

func TestFindQuestArchetypeSuggestionExistingLocationArchetypeMatchReusesNameOrSignature(t *testing.T) {
	bookNookID := uuid.New()
	existing := []*models.LocationArchetype{
		{
			ID:            bookNookID,
			Name:          "Cozy Book Nook",
			IncludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeBookStore, googlemaps.TypeCafe},
			ExcludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeNightClub},
		},
	}

	byName := findQuestArchetypeSuggestionExistingLocationArchetypeMatch(
		sanitizedQuestArchetypeGeneratedLocationArchetype{
			Name:          "cozy book nook",
			IncludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeBookStore},
		},
		existing,
	)
	if byName == nil || byName.ID != bookNookID {
		t.Fatalf("expected name-based archetype reuse, got %+v", byName)
	}

	bySignature := findQuestArchetypeSuggestionExistingLocationArchetypeMatch(
		sanitizedQuestArchetypeGeneratedLocationArchetype{
			Name:          "Reading Hideaway",
			IncludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeBookStore, googlemaps.TypeCafe},
			ExcludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeNightClub},
		},
		existing,
	)
	if bySignature == nil || bySignature.ID != bookNookID {
		t.Fatalf("expected signature-based archetype reuse, got %+v", bySignature)
	}
}
