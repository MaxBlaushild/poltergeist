package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestInventoryProgressionBandNameStripsExistingBandPrefix(t *testing.T) {
	got := inventoryProgressionBandName("Greater Nightglass Phial", "Major")
	if got != "Major Nightglass Phial" {
		t.Fatalf("expected renamed progression item, got %q", got)
	}
}

func TestCloneAndScaleInventoryRecipesRemapsIDsAndQuantities(t *testing.T) {
	recipes, idMap := cloneAndScaleInventoryRecipes(
		models.InventoryRecipes{
			{
				ID:       "old_recipe",
				Tier:     1,
				IsPublic: true,
				Ingredients: []models.InventoryRecipeIngredient{
					{ItemID: 101, Quantity: 2},
				},
			},
		},
		4.0,
		3,
	)

	if len(recipes) != 1 {
		t.Fatalf("expected 1 recipe, got %d", len(recipes))
	}
	if recipes[0].ID == "old_recipe" || recipes[0].ID == "" {
		t.Fatalf("expected cloned recipe ID, got %q", recipes[0].ID)
	}
	if recipes[0].Tier != 3 {
		t.Fatalf("expected recipe tier 3, got %d", recipes[0].Tier)
	}
	if recipes[0].Ingredients[0].Quantity <= 2 {
		t.Fatalf("expected scaled ingredient quantity, got %d", recipes[0].Ingredients[0].Quantity)
	}
	if idMap["old_recipe"] != recipes[0].ID {
		t.Fatalf("expected id map to point to cloned recipe id")
	}
}

func TestRemapTeachRecipeIDsUsesClonedRecipeIDs(t *testing.T) {
	remapped := remapTeachRecipeIDs(
		models.StringArray{"old_recipe", "external_recipe", "old_recipe"},
		map[string]string{"old_recipe": "new_recipe"},
	)
	if len(remapped) != 2 {
		t.Fatalf("expected 2 unique recipe ids, got %d", len(remapped))
	}
	if remapped[0] != "new_recipe" {
		t.Fatalf("expected first recipe id to remap, got %q", remapped[0])
	}
	if remapped[1] != "external_recipe" {
		t.Fatalf("expected external recipe id to remain unchanged, got %q", remapped[1])
	}
}

func TestInventoryItemUpsertRequestFromDraftPayloadClearsUnlockLocksStrength(t *testing.T) {
	request := inventoryItemUpsertRequestFromDraftPayload(models.InventoryItem{
		Name:                "Latchspike",
		UnlockLocksStrength: intPtr(42),
	})

	if request.UnlockLocksStrength != nil {
		t.Fatalf("expected generated draft conversion to clear unlock locks strength")
	}
}

func TestClearGeneratedInventoryUnlockLocksStrength(t *testing.T) {
	item := &models.InventoryItem{
		Name:                "Latchspike",
		UnlockLocksStrength: intPtr(40),
	}

	clearGeneratedInventoryUnlockLocksStrength(item)
	if item.UnlockLocksStrength != nil {
		t.Fatalf("expected helper to clear unlock locks strength")
	}
}
