package processors

import (
	"reflect"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
)

func TestSanitizeGeneratedLocationArchetypesNormalizesTypesAndDedupes(t *testing.T) {
	allowedIndex := buildLocationArchetypePlaceTypeIndex([]googlemaps.PlaceType{
		googlemaps.TypeBookStore,
		googlemaps.TypeCafe,
		googlemaps.TypeLibrary,
		googlemaps.TypeMuseum,
	})

	raw := []generatedLocationArchetypePayload{
		{
			Name:          " Cozy Reading Nook ",
			IncludedTypes: []string{"book_store", "cafe", "BOOK STORE"},
			ExcludedTypes: []string{"museum", "book store", "not_a_real_type"},
		},
		{
			Name:          "Cozy Reading Nook",
			IncludedTypes: []string{"library"},
		},
		{
			Name:          "Quiet Curiosity Hall",
			IncludedTypes: []string{"museum", "library"},
			ExcludedTypes: []string{"cafe"},
		},
		{
			Name:          "",
			IncludedTypes: []string{"cafe"},
		},
	}

	sanitized := sanitizeGeneratedLocationArchetypes(raw, allowedIndex, 10)
	if len(sanitized) != 2 {
		t.Fatalf("expected 2 sanitized archetypes, got %d", len(sanitized))
	}

	if sanitized[0].Name != "Cozy Reading Nook" {
		t.Fatalf("expected cleaned name, got %q", sanitized[0].Name)
	}
	if !reflect.DeepEqual(
		[]googlemaps.PlaceType(sanitized[0].IncludedTypes),
		[]googlemaps.PlaceType{googlemaps.TypeBookStore, googlemaps.TypeCafe},
	) {
		t.Fatalf("unexpected included types: %#v", sanitized[0].IncludedTypes)
	}
	if !reflect.DeepEqual(
		[]googlemaps.PlaceType(sanitized[0].ExcludedTypes),
		[]googlemaps.PlaceType{googlemaps.TypeMuseum},
	) {
		t.Fatalf("unexpected excluded types: %#v", sanitized[0].ExcludedTypes)
	}

	if sanitized[1].Name != "Quiet Curiosity Hall" {
		t.Fatalf("expected second name to survive, got %q", sanitized[1].Name)
	}
}

func TestBuildLocationArchetypeSignatureDedupesEquivalentTypeMixes(t *testing.T) {
	first := buildLocationArchetypeSignature(
		googlemaps.PlaceTypeSlice{googlemaps.TypeCafe, googlemaps.TypeBookStore},
		googlemaps.PlaceTypeSlice{googlemaps.TypeMuseum},
	)
	second := buildLocationArchetypeSignature(
		googlemaps.PlaceTypeSlice{googlemaps.TypeBookStore, googlemaps.TypeCafe},
		googlemaps.PlaceTypeSlice{googlemaps.TypeMuseum},
	)

	if first == "" || second == "" {
		t.Fatalf("expected non-empty signatures")
	}
	if first != second {
		t.Fatalf("expected equivalent place type sets to share a signature, got %q and %q", first, second)
	}
}

func TestNormalizeLocationArchetypeNameKeyCollapsesPunctuation(t *testing.T) {
	actual := normalizeLocationArchetypeNameKey("Sunset, Sips & Stories!")
	if actual != "sunset sips stories" {
		t.Fatalf("expected normalized name key, got %q", actual)
	}
}
