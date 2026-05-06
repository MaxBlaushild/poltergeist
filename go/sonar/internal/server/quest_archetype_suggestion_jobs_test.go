package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestBuildQuestArchetypeSuggestionExpositionTemplateUsesReusableDefaults(t *testing.T) {
		template := buildQuestArchetypeSuggestionExpositionTemplate(
			models.QuestArchetypeSuggestionStep{
				ExpositionTitle:       "Witness Echo",
				ExpositionDescription: "At the observatory, soot-dark notes still cling to the brass housing. The whole message reads like a warning left behind in a hurry.",
				ExpositionSpeakerName: "Archivist Echo",
				ExpositionPortraitURL: "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png",
				ExpositionDialogue: []string{
					"The lens saw the breach before we did.",
					"If you follow the sparks, do not trust the first safe-looking door.",
				},
		},
		&models.ZoneKind{Slug: "academy"},
	)

	if template == nil {
		t.Fatalf("expected template")
	}
	if template.ZoneKind != "academy" {
		t.Fatalf("expected academy zone kind, got %q", template.ZoneKind)
	}
	if template.Title != "Witness Echo" {
		t.Fatalf("expected title to carry through, got %q", template.Title)
	}
	if len(template.Dialogue) != 2 {
		t.Fatalf("expected dialogue lines to carry through, got %+v", template.Dialogue)
	}
		if template.Dialogue[0].SpeakerName != "Archivist Echo" {
			t.Fatalf("expected reusable speaker label to carry through, got %+v", template.Dialogue[0])
		}
		if template.Dialogue[0].PortraitURL != "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png" {
			t.Fatalf("expected reusable portrait url to carry through, got %+v", template.Dialogue[0])
		}
		if template.RewardMode != models.RewardModeRandom {
			t.Fatalf("expected random reward mode, got %q", template.RewardMode)
		}
	if template.RandomRewardSize != models.RandomRewardSizeSmall {
		t.Fatalf("expected small random reward size, got %q", template.RandomRewardSize)
	}
}
