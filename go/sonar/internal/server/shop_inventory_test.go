package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestNormalizeShopMetadataDefaults(t *testing.T) {
	metadata := normalizeShopMetadata(nil)

	if metadata["shopMode"] != shopModeExplicit {
		t.Fatalf("expected default shopMode to be %q, got %v", shopModeExplicit, metadata["shopMode"])
	}
	if len(parseShopInventoryItems(metadata["inventory"])) != 0 {
		t.Fatalf("expected default inventory to be empty")
	}
	if len(parseShopItemTags(metadata["shopItemTags"])) != 0 {
		t.Fatalf("expected default shopItemTags to be empty")
	}
}

func TestNormalizeShopMetadataTagsMode(t *testing.T) {
	metadata := normalizeShopMetadata(map[string]interface{}{
		"shopMode": " tags ",
		"shopItemTags": []interface{}{
			"Potion",
			" potion ",
			"alchemy",
		},
		"inventory": []interface{}{
			map[string]interface{}{"itemId": 10.0, "price": 12.0},
		},
	})

	if metadata["shopMode"] != shopModeTags {
		t.Fatalf("expected shopMode to be %q, got %v", shopModeTags, metadata["shopMode"])
	}
	tags := parseShopItemTags(metadata["shopItemTags"])
	if len(tags) != 2 {
		t.Fatalf("expected 2 normalized tags, got %d (%v)", len(tags), tags)
	}
	if tags[0] != "alchemy" || tags[1] != "potion" {
		t.Fatalf("unexpected normalized tags: %v", tags)
	}
}

func TestResolveTaggedShopInventoryFiltersByTagAndLevelBand(t *testing.T) {
	buyPrice := 50
	items := []models.InventoryItem{
		{ID: 1, ItemLevel: 40, InternalTags: models.StringArray{"potion"}, BuyPrice: &buyPrice},
		{ID: 2, ItemLevel: 60, InternalTags: models.StringArray{"potion"}},
		{ID: 3, ItemLevel: 62, InternalTags: models.StringArray{"elixir"}},
		{ID: 4, ItemLevel: 20, InternalTags: models.StringArray{"potion"}},
	}

	resolved := resolveTaggedShopInventory(items, 50, []string{"potion"})
	if len(resolved) != 2 {
		t.Fatalf("expected 2 matching items in level range, got %d (%v)", len(resolved), resolved)
	}
	if resolved[0].ItemID != 1 || resolved[1].ItemID != 2 {
		t.Fatalf("unexpected item IDs in resolved inventory: %+v", resolved)
	}
	if resolved[0].Price != 50 {
		t.Fatalf("expected buy-price-based price for item 1 to be 50, got %d", resolved[0].Price)
	}
	if resolved[1].Price <= 0 {
		t.Fatalf("expected generated price for item 2 to be positive, got %d", resolved[1].Price)
	}
}

func TestAdjustedShopPriceForCharisma(t *testing.T) {
	if got := adjustedShopPurchasePrice(100, 0); got != 100 {
		t.Fatalf("expected charisma 0 purchase price to stay at 100, got %d", got)
	}
	if got := adjustedShopSellPrice(100, 0); got != 50 {
		t.Fatalf("expected charisma 0 sell price to be 50, got %d", got)
	}
	if got := adjustedShopPurchasePrice(100, 100); got != 75 {
		t.Fatalf("expected charisma 100 purchase price to be 75, got %d", got)
	}
	if got := adjustedShopSellPrice(100, 100); got != 75 {
		t.Fatalf("expected charisma 100 sell price to be 75, got %d", got)
	}
}

func TestPriceShopInventoryForUserSkipsMissingItems(t *testing.T) {
	buyPrice := 25
	itemByID := map[int]models.InventoryItem{
		1: {ID: 1, BuyPrice: &buyPrice},
	}

	priced := priceShopInventoryForUser(
		[]shopInventoryItem{
			{ItemID: 1, Price: 25},
			{ItemID: 999, Price: 50},
		},
		itemByID,
		0,
	)

	if len(priced) != 1 {
		t.Fatalf("expected only active/mapped items to remain, got %+v", priced)
	}
	if priced[0].ItemID != 1 {
		t.Fatalf("expected item 1 to remain, got %+v", priced)
	}
}
