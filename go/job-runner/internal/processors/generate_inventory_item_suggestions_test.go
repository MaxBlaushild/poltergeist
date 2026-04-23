package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
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

	draft := sanitizeInventoryItemSuggestionDraft(spec, job, nil, map[string]struct{}{}, map[string]struct{}{})
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

	draft := sanitizeInventoryItemSuggestionDraft(spec, job, nil, map[string]struct{}{}, map[string]struct{}{})
	if draft.Payload.Item.UnlockLocksStrength != nil {
		t.Fatalf("expected unlock locks strength to be cleared, got %v", *draft.Payload.Item.UnlockLocksStrength)
	}
	if draft.Category != "material" {
		t.Fatalf("expected cleared draft to fall back to material category, got %q", draft.Category)
	}
}

func TestSanitizeInventoryItemSuggestionDraftNormalizesZoneKind(t *testing.T) {
	job := &models.InventoryItemSuggestionJob{
		MinItemLevel: 1,
		MaxItemLevel: 10,
	}
	spec := inventoryItemSuggestionDraftPayload{
		Item: models.InventoryItem{
			Name:     "Bogglass Tonic",
			ZoneKind: "Swamp Bog",
		},
	}
	zoneKinds := []models.ZoneKind{
		{Slug: "swamp_bog", Name: "Swamp Bog"},
		{Slug: "city_streets", Name: "City Streets"},
	}

	draft := sanitizeInventoryItemSuggestionDraft(spec, job, zoneKinds, map[string]struct{}{}, map[string]struct{}{})
	if draft.Payload.Item.ZoneKind != "swamp-bog" {
		t.Fatalf("expected normalized zone kind, got %q", draft.Payload.Item.ZoneKind)
	}
}

func TestEnforceInventoryResourceProgressionDraftNormalizesMaterialFields(t *testing.T) {
	resourceTypeID := uuid.New()
	job := &models.InventoryItemSuggestionJob{
		JobKind:        models.InventoryItemSuggestionJobKindResourceProgression,
		ZoneKind:       "volcanic-caldera",
		ResourceTypeID: &resourceTypeID,
	}
	resourceType := &models.ResourceType{
		ID:   resourceTypeID,
		Name: "Mining",
		Slug: "mining",
	}
	draft := &models.InventoryItemSuggestionDraft{
		Category:   "equippable",
		RarityTier: "Common",
		ItemLevel:  12,
		EquipSlot:  stringPtr("dominant_hand"),
		Payload: models.InventoryItemSuggestionPayloadValue{
			Category: "equippable",
			Item: models.InventoryItem{
				Name:                "Caldera Pick",
				EquipSlot:           stringPtr("dominant_hand"),
				StrengthMod:         4,
				ConsumeHealthDelta:  10,
				UnlockLocksStrength: intPtr(20),
			},
		},
	}

	enforceInventoryResourceProgressionDraft(
		draft,
		job,
		resourceType,
		inventoryResourceProgressionTarget{Label: "Apex", Level: 100, RarityTier: "Mythic"},
	)

	if draft.Category != "material" {
		t.Fatalf("expected resource progression category to be material, got %q", draft.Category)
	}
	if draft.Payload.Item.EquipSlot != nil {
		t.Fatalf("expected equip slot to be cleared")
	}
	if draft.Payload.Item.StrengthMod != 0 {
		t.Fatalf("expected combat stats to be cleared, got %d", draft.Payload.Item.StrengthMod)
	}
	if draft.Payload.Item.ConsumeHealthDelta != 0 {
		t.Fatalf("expected consume effects to be cleared, got %d", draft.Payload.Item.ConsumeHealthDelta)
	}
	if draft.Payload.Item.ResourceTypeID == nil || *draft.Payload.Item.ResourceTypeID != resourceTypeID {
		t.Fatalf("expected resource type id to be applied, got %+v", draft.Payload.Item.ResourceTypeID)
	}
	if draft.Payload.Item.ZoneKind != "volcanic-caldera" {
		t.Fatalf("expected zone kind to be enforced, got %q", draft.Payload.Item.ZoneKind)
	}
	if draft.Payload.Item.ItemLevel != 100 {
		t.Fatalf("expected target level to be enforced, got %d", draft.Payload.Item.ItemLevel)
	}
	if draft.Payload.Item.RarityTier != "Mythic" {
		t.Fatalf("expected target rarity to be enforced, got %q", draft.Payload.Item.RarityTier)
	}
	if len(draft.Warnings) == 0 {
		t.Fatalf("expected normalization warning for non-material inputs")
	}
}
