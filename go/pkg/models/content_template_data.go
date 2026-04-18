package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type CharacterTemplateData struct {
	Name                  string
	Description           string
	InternalTags          StringArray
	MapIconURL            string
	DialogueImageURL      string
	ThumbnailURL          string
	GenreID               uuid.UUID
	Genre                 *ZoneGenre
	StoryVariants         CharacterStoryVariants
	ImageGenerationStatus string
	ImageGenerationError  *string
}

type CharacterTemplateInstanceOptions struct {
	ID                uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	OwnerUserID       *uuid.UUID
	Ephemeral         bool
	PointOfInterestID *uuid.UUID
	InternalTags      StringArray
}

func CharacterTemplateDataFromCharacter(source *Character) CharacterTemplateData {
	if source == nil {
		return CharacterTemplateData{}
	}
	return CharacterTemplateData{
		Name:                  strings.TrimSpace(source.Name),
		Description:           strings.TrimSpace(source.Description),
		InternalTags:          append(StringArray{}, source.InternalTags...),
		MapIconURL:            strings.TrimSpace(source.MapIconURL),
		DialogueImageURL:      strings.TrimSpace(source.DialogueImageURL),
		ThumbnailURL:          strings.TrimSpace(source.ThumbnailURL),
		GenreID:               source.GenreID,
		Genre:                 source.Genre,
		StoryVariants:         cloneCharacterStoryVariants(source.StoryVariants),
		ImageGenerationStatus: strings.TrimSpace(source.ImageGenerationStatus),
		ImageGenerationError:  cloneOptionalTrimmedString(source.ImageGenerationError),
	}
}

func CharacterTemplateDataFromCharacterTemplate(source *CharacterTemplate) CharacterTemplateData {
	if source == nil {
		return CharacterTemplateData{}
	}
	return CharacterTemplateData{
		Name:                  strings.TrimSpace(source.Name),
		Description:           strings.TrimSpace(source.Description),
		InternalTags:          append(StringArray{}, source.InternalTags...),
		MapIconURL:            strings.TrimSpace(source.MapIconURL),
		DialogueImageURL:      strings.TrimSpace(source.DialogueImageURL),
		ThumbnailURL:          strings.TrimSpace(source.ThumbnailURL),
		GenreID:               source.GenreID,
		Genre:                 source.Genre,
		StoryVariants:         cloneCharacterStoryVariants(source.StoryVariants),
		ImageGenerationStatus: strings.TrimSpace(source.ImageGenerationStatus),
		ImageGenerationError:  cloneOptionalTrimmedString(source.ImageGenerationError),
	}
}

func (t CharacterTemplateData) Instantiate(options CharacterTemplateInstanceOptions) *Character {
	id := options.ID
	if id == uuid.Nil {
		id = uuid.New()
	}
	createdAt := options.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	updatedAt := options.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	internalTags := append(StringArray{}, t.InternalTags...)
	if options.InternalTags != nil {
		internalTags = append(StringArray{}, options.InternalTags...)
	}
	return &Character{
		ID:                    id,
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
		Name:                  strings.TrimSpace(t.Name),
		Description:           strings.TrimSpace(t.Description),
		InternalTags:          internalTags,
		MapIconURL:            strings.TrimSpace(t.MapIconURL),
		DialogueImageURL:      strings.TrimSpace(t.DialogueImageURL),
		ThumbnailURL:          strings.TrimSpace(t.ThumbnailURL),
		GenreID:               t.GenreID,
		Genre:                 t.Genre,
		OwnerUserID:           options.OwnerUserID,
		Ephemeral:             options.Ephemeral,
		StoryVariants:         cloneCharacterStoryVariants(t.StoryVariants),
		PointOfInterestID:     cloneOptionalUUID(options.PointOfInterestID),
		ImageGenerationStatus: strings.TrimSpace(t.ImageGenerationStatus),
		ImageGenerationError:  cloneOptionalTrimmedString(t.ImageGenerationError),
	}
}

type ExpositionTemplateData struct {
	Title              string
	Description        string
	Dialogue           DialogueSequence
	RequiredStoryFlags StringArray
	ImageURL           string
	ThumbnailURL       string
	RewardMode         RewardMode
	RandomRewardSize   RandomRewardSize
	RewardExperience   int
	RewardGold         int
	MaterialRewards    BaseMaterialRewards
	ItemRewards        []ExpositionItemReward
	SpellRewards       []ExpositionSpellReward
}

type ExpositionTemplateInstanceOptions struct {
	ID                uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ZoneID            uuid.UUID
	PointOfInterestID *uuid.UUID
	Latitude          float64
	Longitude         float64
}

func ExpositionTemplateDataFromExposition(source *Exposition) ExpositionTemplateData {
	if source == nil {
		return ExpositionTemplateData{}
	}
	return ExpositionTemplateData{
		Title:              strings.TrimSpace(source.Title),
		Description:        strings.TrimSpace(source.Description),
		Dialogue:           cloneDialogueSequence(source.Dialogue),
		RequiredStoryFlags: append(StringArray{}, source.RequiredStoryFlags...),
		ImageURL:           strings.TrimSpace(source.ImageURL),
		ThumbnailURL:       strings.TrimSpace(source.ThumbnailURL),
		RewardMode:         source.RewardMode,
		RandomRewardSize:   source.RandomRewardSize,
		RewardExperience:   source.RewardExperience,
		RewardGold:         source.RewardGold,
		MaterialRewards:    append(BaseMaterialRewards{}, source.MaterialRewards...),
		ItemRewards:        cloneExpositionItemRewards(source.ItemRewards),
		SpellRewards:       cloneExpositionSpellRewards(source.SpellRewards),
	}
}

func ExpositionTemplateDataFromExpositionTemplate(source *ExpositionTemplate) ExpositionTemplateData {
	if source == nil {
		return ExpositionTemplateData{}
	}
	itemRewards := make([]ExpositionItemReward, 0, len(source.ItemRewards))
	for _, reward := range source.ItemRewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		itemRewards = append(itemRewards, ExpositionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	spellRewards := make([]ExpositionSpellReward, 0, len(source.SpellRewards))
	for _, reward := range source.SpellRewards {
		if reward.SpellID == uuid.Nil {
			continue
		}
		spellRewards = append(spellRewards, ExpositionSpellReward{
			SpellID: reward.SpellID,
		})
	}
	return ExpositionTemplateData{
		Title:              strings.TrimSpace(source.Title),
		Description:        strings.TrimSpace(source.Description),
		Dialogue:           cloneDialogueSequence(source.Dialogue),
		RequiredStoryFlags: append(StringArray{}, source.RequiredStoryFlags...),
		ImageURL:           strings.TrimSpace(source.ImageURL),
		ThumbnailURL:       strings.TrimSpace(source.ThumbnailURL),
		RewardMode:         source.RewardMode,
		RandomRewardSize:   source.RandomRewardSize,
		RewardExperience:   source.RewardExperience,
		RewardGold:         source.RewardGold,
		MaterialRewards:    append(BaseMaterialRewards{}, source.MaterialRewards...),
		ItemRewards:        itemRewards,
		SpellRewards:       spellRewards,
	}
}

func ExpositionTemplateDataFromQuestArchetypeNode(node *QuestArchetypeNode) ExpositionTemplateData {
	if node == nil {
		return ExpositionTemplateData{}
	}
	itemRewards := make([]ExpositionItemReward, 0, len(node.ExpositionItemRewards))
	for _, reward := range node.ExpositionItemRewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		itemRewards = append(itemRewards, ExpositionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	spellRewards := make([]ExpositionSpellReward, 0, len(node.ExpositionSpellRewards))
	for _, reward := range node.ExpositionSpellRewards {
		if reward.SpellID == uuid.Nil {
			continue
		}
		spellRewards = append(spellRewards, ExpositionSpellReward{
			SpellID: reward.SpellID,
		})
	}
	return ExpositionTemplateData{
		Title:            strings.TrimSpace(node.ExpositionTitle),
		Description:      strings.TrimSpace(node.ExpositionDescription),
		Dialogue:         cloneDialogueSequence(node.ExpositionDialogue),
		ImageURL:         "",
		ThumbnailURL:     "",
		RewardMode:       node.ExpositionRewardMode,
		RandomRewardSize: node.ExpositionRandomRewardSize,
		RewardExperience: node.ExpositionRewardExperience,
		RewardGold:       node.ExpositionRewardGold,
		MaterialRewards:  append(BaseMaterialRewards{}, node.ExpositionMaterialRewards...),
		ItemRewards:      itemRewards,
		SpellRewards:     spellRewards,
	}
}

func (t ExpositionTemplateData) Instantiate(options ExpositionTemplateInstanceOptions) *Exposition {
	id := options.ID
	if id == uuid.Nil {
		id = uuid.New()
	}
	createdAt := options.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	updatedAt := options.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	return &Exposition{
		ID:                 id,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		ZoneID:             options.ZoneID,
		PointOfInterestID:  cloneOptionalUUID(options.PointOfInterestID),
		Latitude:           options.Latitude,
		Longitude:          options.Longitude,
		Title:              strings.TrimSpace(t.Title),
		Description:        strings.TrimSpace(t.Description),
		Dialogue:           cloneDialogueSequence(t.Dialogue),
		RequiredStoryFlags: append(StringArray{}, t.RequiredStoryFlags...),
		ImageURL:           strings.TrimSpace(t.ImageURL),
		ThumbnailURL:       strings.TrimSpace(t.ThumbnailURL),
		RewardMode:         t.RewardMode,
		RandomRewardSize:   t.RandomRewardSize,
		RewardExperience:   t.RewardExperience,
		RewardGold:         t.RewardGold,
		MaterialRewards:    append(BaseMaterialRewards{}, t.MaterialRewards...),
	}
}

func (t ExpositionTemplateData) ItemRewardsForExposition(expositionID uuid.UUID) []ExpositionItemReward {
	rewards := make([]ExpositionItemReward, 0, len(t.ItemRewards))
	for _, reward := range t.ItemRewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		rewards = append(rewards, ExpositionItemReward{
			ExpositionID:    expositionID,
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func (t ExpositionTemplateData) SpellRewardsForExposition(expositionID uuid.UUID) []ExpositionSpellReward {
	rewards := make([]ExpositionSpellReward, 0, len(t.SpellRewards))
	for _, reward := range t.SpellRewards {
		if reward.SpellID == uuid.Nil {
			continue
		}
		rewards = append(rewards, ExpositionSpellReward{
			ExpositionID: expositionID,
			SpellID:      reward.SpellID,
		})
	}
	return rewards
}

func cloneOptionalTrimmedString(input *string) *string {
	if input == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*input)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func cloneOptionalUUID(input *uuid.UUID) *uuid.UUID {
	if input == nil {
		return nil
	}
	value := *input
	return &value
}

func cloneExpositionItemRewards(input []ExpositionItemReward) []ExpositionItemReward {
	rewards := make([]ExpositionItemReward, 0, len(input))
	for _, reward := range input {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		rewards = append(rewards, ExpositionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func cloneExpositionSpellRewards(input []ExpositionSpellReward) []ExpositionSpellReward {
	rewards := make([]ExpositionSpellReward, 0, len(input))
	for _, reward := range input {
		if reward.SpellID == uuid.Nil {
			continue
		}
		rewards = append(rewards, ExpositionSpellReward{
			SpellID: reward.SpellID,
		})
	}
	return rewards
}

func cloneCharacterStoryVariants(input CharacterStoryVariants) CharacterStoryVariants {
	variants := make(CharacterStoryVariants, 0, len(input))
	for _, variant := range input {
		variants = append(variants, CharacterStoryVariant{
			ID:                 variant.ID,
			CreatedAt:          variant.CreatedAt,
			UpdatedAt:          variant.UpdatedAt,
			Priority:           variant.Priority,
			RequiredStoryFlags: append(StringArray{}, variant.RequiredStoryFlags...),
			Description:        strings.TrimSpace(variant.Description),
			Dialogue:           cloneDialogueSequence(variant.Dialogue),
		})
	}
	return variants
}

func cloneDialogueSequence(input DialogueSequence) DialogueSequence {
	return normalizeDialogueSequence(append(DialogueSequence{}, input...))
}
