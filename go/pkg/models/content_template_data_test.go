package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCharacterTemplateDataInstantiateClonesSourceCharacter(t *testing.T) {
	dialogueCharacterID := uuid.New()
	errMessage := "  generation failed  "
	source := &Character{
		Name:             "  Mire Seer  ",
		Description:      "  Keeps watch over the crossing.  ",
		InternalTags:     StringArray{"guide", "tutorial"},
		MapIconURL:       "  map.png  ",
		DialogueImageURL: "  dialogue.png  ",
		ThumbnailURL:     "  thumb.png  ",
		StoryVariants: CharacterStoryVariants{
			{
				Priority:           2,
				RequiredStoryFlags: StringArray{"met_mire_seer"},
				Description:        "  A sterner warning.  ",
				Dialogue: DialogueSequence{
					{
						Speaker:     "character",
						Text:        "  Beware the lantern light.  ",
						Order:       3,
						Effect:      DialogueEffectMysterious,
						CharacterID: &dialogueCharacterID,
					},
				},
			},
		},
		ImageGenerationStatus: "  queued  ",
		ImageGenerationError:  &errMessage,
	}

	template := CharacterTemplateDataFromCharacter(source)
	source.InternalTags[0] = "changed"
	source.StoryVariants[0].RequiredStoryFlags[0] = "changed"
	source.StoryVariants[0].Dialogue[0].Text = "changed"

	userID := uuid.New()
	pointOfInterestID := uuid.New()
	generated := template.Instantiate(CharacterTemplateInstanceOptions{
		ID:                uuid.New(),
		CreatedAt:         time.Unix(100, 0),
		UpdatedAt:         time.Unix(120, 0),
		OwnerUserID:       &userID,
		Ephemeral:         true,
		PointOfInterestID: &pointOfInterestID,
		InternalTags:      StringArray{"generated", "tutorial"},
	})

	if generated.Name != "Mire Seer" {
		t.Fatalf("expected trimmed name, got %q", generated.Name)
	}
	if generated.Description != "Keeps watch over the crossing." {
		t.Fatalf("expected trimmed description, got %q", generated.Description)
	}
	if len(generated.InternalTags) != 2 || generated.InternalTags[0] != "generated" || generated.InternalTags[1] != "tutorial" {
		t.Fatalf("expected override internal tags, got %+v", generated.InternalTags)
	}
	if generated.MapIconURL != "map.png" || generated.DialogueImageURL != "dialogue.png" || generated.ThumbnailURL != "thumb.png" {
		t.Fatalf("expected trimmed image urls, got map=%q dialogue=%q thumb=%q", generated.MapIconURL, generated.DialogueImageURL, generated.ThumbnailURL)
	}
	if generated.OwnerUserID == nil || *generated.OwnerUserID != userID {
		t.Fatal("expected owner user id to be copied")
	}
	if generated.PointOfInterestID == nil || *generated.PointOfInterestID != pointOfInterestID {
		t.Fatal("expected point of interest id to be copied")
	}
	if generated.ImageGenerationError == nil || *generated.ImageGenerationError != "generation failed" {
		t.Fatalf("expected trimmed image generation error, got %+v", generated.ImageGenerationError)
	}
	if len(generated.StoryVariants) != 1 {
		t.Fatalf("expected one story variant, got %d", len(generated.StoryVariants))
	}
	if generated.StoryVariants[0].RequiredStoryFlags[0] != "met_mire_seer" {
		t.Fatalf("expected deep-copied story flags, got %+v", generated.StoryVariants[0].RequiredStoryFlags)
	}
	if generated.StoryVariants[0].Dialogue[0].Text != "Beware the lantern light." {
		t.Fatalf("expected deep-copied normalized dialogue, got %+v", generated.StoryVariants[0].Dialogue)
	}
}

func TestExpositionTemplateDataFromQuestArchetypeNodeInstantiatesContent(t *testing.T) {
	dialogueCharacterID := uuid.New()
	spellID := uuid.New()
	node := &QuestArchetypeNode{
		ExpositionTitle:       "  Doorstep Warning  ",
		ExpositionDescription: "  Listen closely.  ",
		ExpositionDialogue: DialogueSequence{
			{
				Speaker:     "character",
				Text:        "  Keep moving.  ",
				Order:       9,
				Effect:      DialogueEffectMysterious,
				CharacterID: &dialogueCharacterID,
			},
		},
		ExpositionRewardMode:       RewardModeExplicit,
		ExpositionRandomRewardSize: RandomRewardSizeMedium,
		ExpositionRewardExperience: 10,
		ExpositionRewardGold:       7,
		ExpositionMaterialRewards: BaseMaterialRewards{
			{ResourceKey: BaseResourceStone, Amount: 2},
		},
		ExpositionItemRewards: QuestArchetypeExpositionItemRewards{
			{InventoryItemID: 11, Quantity: 2},
		},
		ExpositionSpellRewards: QuestArchetypeExpositionSpellRewards{
			{SpellID: spellID},
		},
	}

	template := ExpositionTemplateDataFromQuestArchetypeNode(node)
	node.ExpositionDialogue[0].Text = "changed"
	node.ExpositionMaterialRewards[0].Amount = 99

	pointOfInterestID := uuid.New()
	instance := template.Instantiate(ExpositionTemplateInstanceOptions{
		ID:                uuid.New(),
		CreatedAt:         time.Unix(200, 0),
		UpdatedAt:         time.Unix(220, 0),
		ZoneID:            uuid.New(),
		PointOfInterestID: &pointOfInterestID,
		Latitude:          40.0,
		Longitude:         -73.0,
	})

	if instance.Title != "Doorstep Warning" {
		t.Fatalf("expected trimmed title, got %q", instance.Title)
	}
	if instance.Description != "Listen closely." {
		t.Fatalf("expected trimmed description, got %q", instance.Description)
	}
	if len(instance.Dialogue) != 1 || instance.Dialogue[0].Text != "Keep moving." {
		t.Fatalf("expected copied dialogue, got %+v", instance.Dialogue)
	}
	if len(instance.MaterialRewards) != 1 || instance.MaterialRewards[0].Amount != 2 {
		t.Fatalf("expected copied material rewards, got %+v", instance.MaterialRewards)
	}
	if instance.RewardMode != RewardModeExplicit || instance.RandomRewardSize != RandomRewardSizeMedium {
		t.Fatalf("expected reward settings to carry through, got mode=%q size=%q", instance.RewardMode, instance.RandomRewardSize)
	}

	itemRewards := template.ItemRewardsForExposition(instance.ID)
	if len(itemRewards) != 1 || itemRewards[0].ExpositionID != instance.ID || itemRewards[0].InventoryItemID != 11 || itemRewards[0].Quantity != 2 {
		t.Fatalf("expected instantiated item rewards, got %+v", itemRewards)
	}
	spellRewards := template.SpellRewardsForExposition(instance.ID)
	if len(spellRewards) != 1 || spellRewards[0].ExpositionID != instance.ID || spellRewards[0].SpellID != spellID {
		t.Fatalf("expected instantiated spell rewards, got %+v", spellRewards)
	}
}

func TestExpositionTemplateDataFromExpositionClonesRewardsAndFlags(t *testing.T) {
	spellID := uuid.New()
	source := &Exposition{
		Title:              "  Shrine Echo  ",
		Description:        "  The air remembers.  ",
		Dialogue:           DialogueSequence{{Speaker: "character", Text: "  Speak softly.  ", Order: 4}},
		RequiredStoryFlags: StringArray{"chapter_1_open"},
		ImageURL:           "  shrine.png  ",
		ThumbnailURL:       "  shrine-thumb.png  ",
		RewardMode:         RewardModeExplicit,
		RandomRewardSize:   RandomRewardSizeLarge,
		RewardExperience:   15,
		RewardGold:         3,
		MaterialRewards: BaseMaterialRewards{
			{ResourceKey: BaseResourceTimber, Amount: 1},
		},
		ItemRewards: []ExpositionItemReward{
			{InventoryItemID: 21, Quantity: 1},
		},
		SpellRewards: []ExpositionSpellReward{
			{SpellID: spellID},
		},
	}

	template := ExpositionTemplateDataFromExposition(source)
	source.RequiredStoryFlags[0] = "changed"
	source.ItemRewards[0].Quantity = 99
	source.SpellRewards[0].SpellID = uuid.New()

	if template.RequiredStoryFlags[0] != "chapter_1_open" {
		t.Fatalf("expected story flags to be cloned, got %+v", template.RequiredStoryFlags)
	}
	if len(template.ItemRewards) != 1 || template.ItemRewards[0].Quantity != 1 {
		t.Fatalf("expected item rewards to be cloned, got %+v", template.ItemRewards)
	}
	if len(template.SpellRewards) != 1 || template.SpellRewards[0].SpellID != spellID {
		t.Fatalf("expected spell rewards to be cloned, got %+v", template.SpellRewards)
	}
	if template.ImageURL != "shrine.png" || template.ThumbnailURL != "shrine-thumb.png" {
		t.Fatalf("expected image urls to be trimmed, got image=%q thumb=%q", template.ImageURL, template.ThumbnailURL)
	}
}
