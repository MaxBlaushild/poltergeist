package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestClassifyScenarioTemplateZoneKindFallsBackHeuristicallyWithoutPriest(t *testing.T) {
	zoneKinds := []models.ZoneKind{
		{Slug: "forest", Name: "Forest", Description: "Ancient woods and tangled groves."},
		{Slug: "swamp", Name: "Swamp", Description: "Boggy wetlands full of mire and reeds."},
	}

	template := &models.ScenarioTemplate{
		Prompt:     "A ferryman begs for help guiding villagers across a foggy bog while something moves beneath the reeds.",
		OpenEnded:  true,
		Difficulty: 18,
		RewardMode: models.RewardModeExplicit,
	}

	got := classifyScenarioTemplateZoneKind(nil, template, zoneKinds, nil)
	if got != "swamp" {
		t.Fatalf("expected swamp, got %q", got)
	}
}
