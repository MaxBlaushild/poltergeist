package processors

import (
	"reflect"
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestGenerateZoneSeedShopkeepersIncludesDialogue(t *testing.T) {
	t.Helper()

	zone := models.Zone{
		Name:      "Lantern Ward",
		Latitude:  40.7128,
		Longitude: -74.0060,
	}

	shopkeepers := generateZoneSeedShopkeepers(zone, []string{"potion_brewing", "ancient_scrolls"})
	if len(shopkeepers) != 2 {
		t.Fatalf("expected 2 shopkeepers, got %d", len(shopkeepers))
	}

	for _, shopkeeper := range shopkeepers {
		if len(shopkeeper.Dialogue) < 2 || len(shopkeeper.Dialogue) > 3 {
			t.Fatalf("expected shopkeeper %q to have 2-3 dialogue lines, got %d", shopkeeper.Name, len(shopkeeper.Dialogue))
		}
		if len(shopkeeper.ShopItemTags) != 1 {
			t.Fatalf("expected shopkeeper %q to have exactly 1 shop tag, got %d", shopkeeper.Name, len(shopkeeper.ShopItemTags))
		}

		tagLabel := humanizeShopkeeperTag(shopkeeper.ShopItemTags[0])
		hasTagOrZoneReference := false
		for _, line := range shopkeeper.Dialogue {
			if strings.TrimSpace(line) == "" {
				t.Fatalf("expected shopkeeper %q dialogue lines to be non-empty", shopkeeper.Name)
			}
			if len(line) > 180 {
				t.Fatalf("expected shopkeeper %q dialogue line to stay short, got %d chars", shopkeeper.Name, len(line))
			}

			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, tagLabel) || strings.Contains(lowerLine, strings.ToLower(zone.Name)) {
				hasTagOrZoneReference = true
			}
		}
		if !hasTagOrZoneReference {
			t.Fatalf("expected shopkeeper %q dialogue to reference either %q or %q", shopkeeper.Name, tagLabel, zone.Name)
		}
	}
}

func TestGenerateZoneSeedShopkeeperDialogueIsStable(t *testing.T) {
	t.Helper()

	first := generateZoneSeedShopkeeperDialogue("potion_brewing", "Lantern Ward", "Aria Amberfall")
	second := generateZoneSeedShopkeeperDialogue("potion_brewing", "Lantern Ward", "Aria Amberfall")

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected shopkeeper dialogue to be stable for the same input")
	}
	if len(first) == 0 {
		t.Fatalf("expected stable shopkeeper dialogue to be non-empty")
	}
}

func TestZoneSeedPointOfInterestShopkeeperTagForRollsHonorsSpawnChance(t *testing.T) {
	t.Helper()

	tag, ok := zoneSeedPointOfInterestShopkeeperTagForRolls(
		models.PointOfInterestMarkerCategoryMarket,
		9500,
		0,
	)
	if ok || tag != "" {
		t.Fatalf("expected market shopkeeper roll to miss when spawn roll exceeds the configured chance, got %q", tag)
	}
}

func TestZoneSeedPointOfInterestShopkeeperTagForRollsUsesWeightedHierarchy(t *testing.T) {
	t.Helper()

	first, ok := zoneSeedPointOfInterestShopkeeperTagForRolls(
		models.PointOfInterestMarkerCategoryArchive,
		0,
		0,
	)
	if !ok || first != "arcane" {
		t.Fatalf("expected first archive tag to be arcane, got ok=%v tag=%q", ok, first)
	}

	second, ok := zoneSeedPointOfInterestShopkeeperTagForRolls(
		models.PointOfInterestMarkerCategoryArchive,
		0,
		6,
	)
	if !ok || second != "guide" {
		t.Fatalf("expected second archive weight band to resolve to guide, got ok=%v tag=%q", ok, second)
	}

	third, ok := zoneSeedPointOfInterestShopkeeperTagForRolls(
		models.PointOfInterestMarkerCategoryArchive,
		0,
		11,
	)
	if !ok || third != "relic" {
		t.Fatalf("expected third archive weight band to resolve to relic, got ok=%v tag=%q", ok, third)
	}
}

func TestZoneSeedPointOfInterestShopkeeperTagForRollsUsesConfiguredProfiles(t *testing.T) {
	t.Helper()

	tag, ok := zoneSeedPointOfInterestShopkeeperTagForRollsWithProfiles(
		models.PointOfInterestMarkerCategoryGeneric,
		0,
		2,
		[]models.PointOfInterestShopkeeperSeedProfile{
			{
				Category:               models.PointOfInterestMarkerCategoryGeneric,
				SpawnChanceBasisPoints: 10_000,
				Candidates: []models.PointOfInterestShopkeeperSeedCandidate{
					{Tag: "guide", Weight: 2},
					{Tag: "arcane", Weight: 3},
				},
			},
		},
	)
	if !ok || tag != "arcane" {
		t.Fatalf("expected configured generic profile to resolve to arcane, got ok=%v tag=%q", ok, tag)
	}
}

func TestBuildZoneSeedPointOfInterestShopkeeperAttachesToPOI(t *testing.T) {
	t.Helper()

	zone := models.Zone{Name: "Lantern Ward"}
	poi := models.ZoneSeedPointOfInterestDraft{
		PlaceID:        "moonwake-cafe",
		Name:           "Moonwake Cafe",
		Latitude:       40.7128,
		Longitude:      -74.0060,
		MarkerCategory: models.PointOfInterestMarkerCategoryCoffeehouse,
	}

	shopkeeper := buildZoneSeedPointOfInterestShopkeeper(zone, poi, "social", map[string]struct{}{})

	if shopkeeper.PlaceID != poi.PlaceID {
		t.Fatalf("expected shopkeeper to attach to POI %q, got %q", poi.PlaceID, shopkeeper.PlaceID)
	}
	if shopkeeper.Latitude == nil || shopkeeper.Longitude == nil {
		t.Fatalf("expected shopkeeper coordinates to be set")
	}
	if *shopkeeper.Latitude != poi.Latitude || *shopkeeper.Longitude != poi.Longitude {
		t.Fatalf(
			"expected shopkeeper coordinates (%f,%f), got (%f,%f)",
			poi.Latitude,
			poi.Longitude,
			*shopkeeper.Latitude,
			*shopkeeper.Longitude,
		)
	}
	if !reflect.DeepEqual([]string(shopkeeper.ShopItemTags), []string{"social"}) {
		t.Fatalf("expected one social shop tag, got %v", shopkeeper.ShopItemTags)
	}
	if !strings.Contains(strings.ToLower(shopkeeper.Description), "moonwake cafe") {
		t.Fatalf("expected shopkeeper description to mention the POI, got %q", shopkeeper.Description)
	}
}
