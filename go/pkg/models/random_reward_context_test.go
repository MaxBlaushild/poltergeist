package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestFilterRewardItemsForContextPrefersKnowledgeAndZoneMatch(t *testing.T) {
	zoneKind := "haunted-streets"
	filtered := filterRewardItemsForContext(
		[]InventoryItem{
			{
				ID:           1,
				Name:         "Street Tonic",
				ItemLevel:    20,
				RarityTier:   "Common",
				InternalTags: StringArray{"martial"},
			},
			{
				ID:                    2,
				Name:                  "Archivist's Field Notes",
				ItemLevel:             21,
				RarityTier:            "Common",
				ZoneKind:              zoneKind,
				InternalTags:          StringArray{"scholar", "arcane"},
				ConsumeTeachRecipeIDs: StringArray{"recipe-1"},
			},
			{
				ID:         3,
				Name:       "Lantern Oil",
				ItemLevel:  19,
				RarityTier: "Common",
				ZoneKind:   zoneKind,
			},
		},
		20,
		false,
		&RandomRewardContext{
			ContentKind: RandomRewardContentExposition,
			ZoneKind:    zoneKind,
		},
	)

	if len(filtered) != 1 {
		t.Fatalf("expected only the strongest exposition match to remain, got %d items", len(filtered))
	}
	if filtered[0].ID != 2 {
		t.Fatalf("expected archivist notes to be the remaining reward candidate, got %d", filtered[0].ID)
	}
}

func TestScoreRewardItemForContextPrefersCombatGearForMonsterRewards(t *testing.T) {
	rewardContext := &RandomRewardContext{
		ContentKind: RandomRewardContentMonsterEncounter,
	}
	preferredTags := rewardContext.PreferredRewardTags()

	martialItem := InventoryItem{
		ID:         1,
		Name:       "Hunter Blade",
		ItemLevel:  25,
		RarityTier: "Common",
		EquipSlot:  stringPtrRandomRewardContext("dominant_hand"),
		InternalTags: StringArray{
			"martial",
			"hunter",
		},
	}
	loreItem := InventoryItem{
		ID:         2,
		Name:       "Museum Signet",
		ItemLevel:  25,
		RarityTier: "Common",
		EquipSlot:  stringPtrRandomRewardContext("dominant_hand"),
		InternalTags: StringArray{
			"scholar",
			"arcane",
		},
	}

	martialScore := scoreRewardItemForContext(martialItem, true, rewardContext, preferredTags)
	loreScore := scoreRewardItemForContext(loreItem, true, rewardContext, preferredTags)
	if martialScore <= loreScore {
		t.Fatalf("expected martial monster reward item score %d to exceed lore item score %d", martialScore, loreScore)
	}
}

func TestBuildRandomRewardPlanForContextUsesPointOfInterestProfile(t *testing.T) {
	plan := BuildRandomRewardPlanForContext(
		30,
		RandomRewardSizeMedium,
		"poi-seed",
		[]InventoryItem{
			{
				ID:           7,
				Name:         "Guide's Satchel",
				ItemLevel:    30,
				RarityTier:   "Common",
				ZoneKind:     "river-market",
				InternalTags: StringArray{"guide", "social", "street"},
			},
			{
				ID:           8,
				Name:         "Arena Pike",
				ItemLevel:    30,
				RarityTier:   "Common",
				ZoneKind:     "river-market",
				EquipSlot:    stringPtrRandomRewardContext("dominant_hand"),
				InternalTags: StringArray{"martial", "frontline"},
			},
			{
				ID:           9,
				Name:         "Forest Remedy",
				ItemLevel:    30,
				RarityTier:   "Common",
				ZoneKind:     "quiet-meadow",
				InternalTags: StringArray{"nature", "wild"},
			},
		},
		&RandomRewardContext{
			ContentKind:             RandomRewardContentPointOfInterest,
			ZoneKind:                "river-market",
			PointOfInterestCategory: PointOfInterestMarkerCategoryMarket,
		},
	)

	if len(plan.ItemGrants) == 0 {
		t.Fatalf("expected a guaranteed reward item for a medium reward plan")
	}
	if plan.ItemGrants[0].InventoryItemID != 7 {
		t.Fatalf("expected market POI rewards to prefer the guide's satchel, got item %d", plan.ItemGrants[0].InventoryItemID)
	}
}

func TestDefaultRewardProfileSlugsForContext(t *testing.T) {
	marketPOI := DefaultRewardProfileSlugsForContext(&RandomRewardContext{
		ContentKind:             RandomRewardContentPointOfInterest,
		PointOfInterestCategory: PointOfInterestMarkerCategoryMarket,
	})
	if len(marketPOI) != 1 || marketPOI[0] != "social" {
		t.Fatalf("expected market POI to map to social profile, got %#v", marketPOI)
	}

	mainStoryScenario := DefaultRewardProfileSlugsForContext(&RandomRewardContext{
		ContentKind:  RandomRewardContentScenario,
		InternalTags: []string{"main_story"},
	})
	if len(mainStoryScenario) != 1 || mainStoryScenario[0] != "story" {
		t.Fatalf("expected main story scenario to map to story profile, got %#v", mainStoryScenario)
	}
}

func TestApplyRewardProfilesMergesPreferences(t *testing.T) {
	resourceTypeID := uuid.New()
	context := &RandomRewardContext{
		PreferredItemTags:     []string{"guide"},
		PreferredMaterialKeys: []BaseResourceKey{BaseResourceHerbs},
		ResourceTypeIDs:       []uuid.UUID{resourceTypeID},
	}
	context.ApplyRewardProfiles([]RewardProfile{
		{
			Slug:                      "combat",
			PreferredItemTags:         StringArray{"martial", "hunter"},
			PreferredMaterialKeys:     StringArray{"monster_parts", "iron"},
			PreferredDamageAffinities: StringArray{"fire"},
			PreferredResourceTypeIDs:  StringArray{uuid.NewString(), resourceTypeID.String()},
			PreferEquipment:           true,
			PreferUtility:             true,
		},
	})

	if !context.PreferEquipment || !context.PreferUtility {
		t.Fatalf("expected profile preferences to enable equipment and utility flags")
	}
	if len(context.PreferredItemTags) != 3 {
		t.Fatalf("expected merged preferred item tags, got %#v", context.PreferredItemTags)
	}
	if len(context.PreferredMaterialKeys) != 3 {
		t.Fatalf("expected merged preferred material keys, got %#v", context.PreferredMaterialKeys)
	}
	if len(context.PreferredDamageTags) != 1 || context.PreferredDamageTags[0] != "fire" {
		t.Fatalf("expected fire damage preference, got %#v", context.PreferredDamageTags)
	}
	if len(context.ResourceTypeIDs) != 2 {
		t.Fatalf("expected merged resource type IDs, got %#v", context.ResourceTypeIDs)
	}
}

func stringPtrRandomRewardContext(value string) *string {
	return &value
}
