package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestSanitizeInventoryItemSuggestionDraftClearsLinkedFieldsAndInfersCategory(t *testing.T) {
	job := &models.InventoryItemSuggestionJob{
		MinItemLevel: 10,
		MaxItemLevel: 20,
		InternalTags: models.StringArray{"alchemy", "starter"},
	}
	spec := inventoryItemSuggestionDraftPayload{
		Item: models.InventoryItem{
			Name:                  "  Nightglass Phial  ",
			RarityTier:            "epic",
			ItemLevel:             3,
			ConsumeManaDelta:      12,
			ConsumeSpellIDs:       models.StringArray{"abc"},
			ConsumeTeachRecipeIDs: models.StringArray{"recipe-1"},
			InternalTags:          models.StringArray{"mana", "starter"},
		},
	}

	draft := sanitizeInventoryItemSuggestionDraft(spec, job, map[string]struct{}{}, map[string]struct{}{})
	if draft.Category != "consumable" {
		t.Fatalf("expected consumable category, got %q", draft.Category)
	}
	if draft.ItemLevel != 10 {
		t.Fatalf("expected item level to clamp to 10, got %d", draft.ItemLevel)
	}
	if draft.Payload.Item.Name != "Nightglass Phial" {
		t.Fatalf("expected trimmed name, got %q", draft.Payload.Item.Name)
	}
	if len(draft.Payload.Item.ConsumeSpellIDs) != 0 {
		t.Fatalf("expected consume spell ids to be cleared")
	}
	if len(draft.Payload.Item.ConsumeTeachRecipeIDs) != 0 {
		t.Fatalf("expected teach recipe ids to be cleared")
	}
	if len(draft.InternalTags) != 3 {
		t.Fatalf("expected merged internal tags, got %v", draft.InternalTags)
	}
	if len(draft.Warnings) == 0 {
		t.Fatalf("expected warning about linked fields")
	}
}

func TestSanitizeInventoryItemSuggestionDraftClearsUnlockLocksStrength(t *testing.T) {
	job := &models.InventoryItemSuggestionJob{
		MinItemLevel: 1,
		MaxItemLevel: 10,
	}
	spec := inventoryItemSuggestionDraftPayload{
		Item: models.InventoryItem{
			Name:                "Latchspike",
			UnlockLocksStrength: intPtr(55),
		},
	}

	draft := sanitizeInventoryItemSuggestionDraft(spec, job, map[string]struct{}{}, map[string]struct{}{})
	if draft.Payload.Item.UnlockLocksStrength != nil {
		t.Fatalf("expected unlock locks strength to be cleared, got %v", *draft.Payload.Item.UnlockLocksStrength)
	}
	if draft.Category != "material" {
		t.Fatalf("expected cleared draft to fall back to material category, got %q", draft.Category)
	}
}
