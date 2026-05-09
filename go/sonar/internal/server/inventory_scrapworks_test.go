package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestBestScrapworksRecipeForTierReturnsHighestAvailableTier(t *testing.T) {
	recipes := models.InventorySalvageRecipes{
		{
			Tier: 1,
			Outputs: []models.InventorySalvageOutput{
				{ItemID: 10, Quantity: 1},
			},
		},
		{
			Tier: 3,
			Outputs: []models.InventorySalvageOutput{
				{ItemID: 11, Quantity: 2},
			},
		},
		{
			Tier: 2,
			Outputs: []models.InventorySalvageOutput{
				{ItemID: 12, Quantity: 3},
			},
		},
	}

	selected := bestScrapworksRecipeForTier(recipes, 2)
	if selected == nil {
		t.Fatal("expected a recipe for room tier 2")
	}
	if selected.Tier != 2 {
		t.Fatalf("expected tier 2 recipe, got %d", selected.Tier)
	}
	if len(selected.Outputs) != 1 || selected.Outputs[0].ItemID != 12 {
		t.Fatalf("expected tier 2 outputs to be selected, got %+v", selected.Outputs)
	}
}

func TestBestScrapworksRecipeForTierReturnsNilWhenNoRecipeMatches(t *testing.T) {
	recipes := models.InventorySalvageRecipes{
		{
			Tier: 2,
			Outputs: []models.InventorySalvageOutput{
				{ItemID: 10, Quantity: 1},
			},
		},
	}

	selected := bestScrapworksRecipeForTier(recipes, 1)
	if selected != nil {
		t.Fatalf("expected no recipe for room tier 1, got tier %d", selected.Tier)
	}
}
