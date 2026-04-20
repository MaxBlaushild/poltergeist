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
