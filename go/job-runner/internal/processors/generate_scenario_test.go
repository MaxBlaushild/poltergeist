package processors

import (
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestShouldUseExistingScenarioTemplateForZoneSeed(t *testing.T) {
	recurringID := uuid.New()
	frequency := "daily"
	nextAt := time.Now().UTC().Add(time.Hour)

	if !shouldUseExistingScenarioTemplateForZoneSeed(&models.ScenarioGenerationJob{
		RecurringScenarioID: &recurringID,
		RecurrenceFrequency: &frequency,
		NextRecurrenceAt:    &nextAt,
	}) {
		t.Fatalf("expected recurring zone-seed scenario jobs to use existing templates")
	}

	if shouldUseExistingScenarioTemplateForZoneSeed(&models.ScenarioGenerationJob{}) {
		t.Fatalf("expected non-recurring scenario jobs to keep the freeform generation path")
	}
}

func TestFilterZoneSeedScenarioTemplatesFiltersByZoneKindModeAndGenre(t *testing.T) {
	genreID := uuid.New()
	otherGenreID := uuid.New()

	templates := []models.ScenarioTemplate{
		{Prompt: "Forest rescue", GenreID: genreID, ZoneKind: "forest", OpenEnded: true},
		{Prompt: "Swamp rescue", GenreID: genreID, ZoneKind: "swamp", OpenEnded: true},
		{Prompt: "Forest choice", GenreID: genreID, ZoneKind: "forest", OpenEnded: false},
		{Prompt: "Forest sci-fi", GenreID: otherGenreID, ZoneKind: "forest", OpenEnded: true},
	}

	filtered := filterZoneSeedScenarioTemplates(templates, genreID, true, "forest")
	if len(filtered) != 1 || filtered[0].Prompt != "Forest rescue" {
		t.Fatalf("expected only matching forest open-ended templates, got %+v", filtered)
	}
}

func TestFilterZoneSeedScenarioTemplatesLeavesLegacyKindlessZonesBroad(t *testing.T) {
	genreID := uuid.New()
	templates := []models.ScenarioTemplate{
		{Prompt: "Forest rescue", GenreID: genreID, ZoneKind: "forest", OpenEnded: true},
		{Prompt: "Swamp rescue", GenreID: genreID, ZoneKind: "swamp", OpenEnded: true},
		{Prompt: "Choice encounter", GenreID: genreID, ZoneKind: "forest", OpenEnded: false},
	}

	filtered := filterZoneSeedScenarioTemplates(templates, genreID, true, "")
	if len(filtered) != 2 {
		t.Fatalf("expected blank-kind zones to keep all matching-mode templates, got %+v", filtered)
	}
}

func TestBuildScenarioFromTemplateCopiesTemplateContent(t *testing.T) {
	recurringID := uuid.New()
	frequency := "daily"
	nextAt := time.Now().UTC().Add(24 * time.Hour)
	genreID := uuid.New()
	spellID := uuid.New()

	job := &models.ScenarioGenerationJob{
		ScaleWithUserLevel:  true,
		RecurringScenarioID: &recurringID,
		RecurrenceFrequency: &frequency,
		NextRecurrenceAt:    &nextAt,
	}
	zone := &models.Zone{
		ID:   uuid.New(),
		Kind: "forest",
	}
	genre := &models.ZoneGenre{
		ID:   genreID,
		Name: "Fantasy",
	}
	template := models.ScenarioTemplate{
		GenreID:                   genreID,
		ZoneKind:                  "forest",
		Prompt:                    "An old shrine coughs green fire into the rain.",
		ImageURL:                  "",
		ThumbnailURL:              "",
		RewardMode:                models.RewardModeExplicit,
		RandomRewardSize:          models.RandomRewardSizeSmall,
		Difficulty:                18,
		RewardExperience:          44,
		RewardGold:                12,
		OpenEnded:                 false,
		SuccessHandoffText:        "The grove breathes easier.",
		FailureHandoffText:        "The corruption deepens.",
		FailurePenaltyMode:        models.ScenarioFailurePenaltyModeShared,
		FailureHealthDrainType:    models.ScenarioFailureDrainTypeFlat,
		FailureHealthDrainValue:   4,
		FailureManaDrainType:      models.ScenarioFailureDrainTypePercent,
		FailureManaDrainValue:     8,
		SuccessRewardMode:         models.ScenarioSuccessRewardModeShared,
		SuccessHealthRestoreType:  models.ScenarioFailureDrainTypeFlat,
		SuccessHealthRestoreValue: 6,
		SuccessManaRestoreType:    models.ScenarioFailureDrainTypeFlat,
		SuccessManaRestoreValue:   3,
		ItemRewards: models.ScenarioTemplateRewards{
			{InventoryItemID: 101, Quantity: 2},
		},
		ItemChoiceRewards: models.ScenarioTemplateRewards{
			{InventoryItemID: 202, Quantity: 1},
		},
		SpellRewards: models.ScenarioTemplateSpellRewards{
			{SpellID: spellID},
		},
		Options: models.ScenarioTemplateOptions{
			{
				OptionText:        "Cleanse the shrine with a careful rite.",
				SuccessText:       "The light steadies.",
				FailureText:       "The shrine lashes out.",
				StatTag:           "wisdom",
				Proficiencies:     models.StringArray{"rituals"},
				RewardExperience:  20,
				RewardGold:        7,
				ItemRewards:       models.ScenarioTemplateRewards{{InventoryItemID: 303, Quantity: 1}},
				ItemChoiceRewards: models.ScenarioTemplateRewards{{InventoryItemID: 404, Quantity: 1}},
				SpellRewards:      models.ScenarioTemplateSpellRewards{{SpellID: spellID}},
			},
		},
	}

	scenario, options, itemRewards, itemChoiceRewards, spellRewards := buildScenarioFromTemplate(
		job,
		zone,
		genre,
		template,
		40.1,
		-73.9,
	)

	if scenario.ZoneKind != "forest" {
		t.Fatalf("expected scenario to inherit zone kind, got %q", scenario.ZoneKind)
	}
	if scenario.ImageURL != scenarioPlaceholderImageURL || scenario.ThumbnailURL != scenarioPlaceholderImageURL {
		t.Fatalf("expected template without art to fall back to the placeholder, got image=%q thumb=%q", scenario.ImageURL, scenario.ThumbnailURL)
	}
	if !scenario.ScaleWithUserLevel || scenario.RecurringScenarioID == nil || scenario.RecurrenceFrequency == nil || scenario.NextRecurrenceAt == nil {
		t.Fatalf("expected scenario to inherit zone-seed recurrence settings, got %+v", scenario)
	}
	if len(options) != 1 || len(options[0].ItemRewards) != 1 || len(options[0].ItemChoiceRewards) != 1 || len(options[0].SpellRewards) != 1 {
		t.Fatalf("expected option rewards to be copied from the template, got %+v", options)
	}
	if len(itemRewards) != 1 || len(itemChoiceRewards) != 1 || len(spellRewards) != 1 {
		t.Fatalf("expected top-level rewards to be copied from the template, got item=%+v choice=%+v spell=%+v", itemRewards, itemChoiceRewards, spellRewards)
	}
}
