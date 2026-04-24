package locationseeder

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestMakePointOfInterestThemingPromptIncludesZoneKindDirection(t *testing.T) {
	client := &client{}
	place := googlemaps.Place{
		DisplayName:      googlemaps.LocalizedText{Text: "Lantern Mill"},
		EditorialSummary: googlemaps.LocalizedText{Text: "A weathered riverside cafe with a creaking wheel."},
		Types:            []string{"cafe", "tourist_attraction"},
		PriceLevel:       "PRICE_LEVEL_MODERATE",
		PrimaryType:      "cafe",
		PrimaryTypeDisplayName: googlemaps.LocalizedText{
			Text: "Cafe",
		},
	}
	zone := &models.Zone{
		Name:        "Whispergrove",
		Description: "A shrine-laced woodland district with mossy paths and old roots.",
		Kind:        "forest",
	}
	genre := &models.ZoneGenre{
		Name:       models.DefaultZoneGenreNameFantasy,
		PromptSeed: models.DefaultFantasyZoneGenrePromptSeed(),
	}
	zoneKind := &models.ZoneKind{
		Slug:        "forest",
		Name:        "Forest",
		Description: "Wild, herbal, canopy-heavy places with old shrine roots and beast-haunted trails.",
	}

	prompt := client.makePointOfInterestThemingPrompt(place, zone, genre, zoneKind)

	for _, needle := range []string{
		"Zone kind direction:",
		"- zone kind: Forest",
		"- slug: forest",
		"Wild, herbal, canopy-heavy places with old shrine roots and beast-haunted trails.",
		"Make the location feel natively at home in forest spaces.",
	} {
		if !strings.Contains(prompt, needle) {
			t.Fatalf("expected prompt to include %q, got:\n%s", needle, prompt)
		}
	}
}

func TestMakePointOfInterestImagePromptIncludesZoneKindDirection(t *testing.T) {
	client := &client{}
	place := googlemaps.Place{
		DisplayName:      googlemaps.LocalizedText{Text: "Ash Market"},
		EditorialSummary: googlemaps.LocalizedText{Text: "A cramped bazaar of salvage stalls and soot-black awnings."},
		Types:            []string{"market", "store"},
		PriceLevel:       "PRICE_LEVEL_INEXPENSIVE",
	}
	zone := &models.Zone{
		Name:        "Cinder Ward",
		Description: "An iron-choked quarter full of furnaces, slag, and working crews.",
		Kind:        "industrial",
	}
	genre := &models.ZoneGenre{
		Name:       models.DefaultZoneGenreNameFantasy,
		PromptSeed: models.DefaultFantasyZoneGenrePromptSeed(),
	}
	zoneKind := &models.ZoneKind{
		Slug:        "industrial",
		Name:        "Industrial",
		Description: "Harsh, mechanical, forge-lit places with soot, rivets, pipes, and laboring crews.",
	}

	prompt := client.makePointOfInterestImagePromptPrompt(place, zone, genre, zoneKind)

	for _, needle := range []string{
		"Zone kind direction:",
		"- zone kind: Industrial",
		"- slug: industrial",
		"Harsh, mechanical, forge-lit places with soot, rivets, pipes, and laboring crews.",
		"Let the fantasy name, lore, architecture, props, and image cues reflect that zone kind",
	} {
		if !strings.Contains(prompt, needle) {
			t.Fatalf("expected prompt to include %q, got:\n%s", needle, prompt)
		}
	}
}
