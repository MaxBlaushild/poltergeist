package processors

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestCollectExpositionTemplateSpeakerPortraitCandidatesSkipsAbstractOrSatisfiedSpeakers(t *testing.T) {
	template := &models.ExpositionTemplate{
		Title:       "Overheard Couriers",
		Description: "A courier dispute at the floodlit market gate is still hanging in the rain.",
		Dialogue: models.DialogueSequence{
			{SpeakerName: "Witness Echo", Text: "Abstract source label.", PortraitURL: "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png"},
			{SpeakerName: "Night Porter", Text: "Keep your voice down.", PortraitURL: "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png"},
			{SpeakerName: "Harried Courier", Text: "Too late. They already know.", PortraitURL: "https://example.com/courier.png"},
		},
	}

	candidates := collectExpositionTemplateSpeakerPortraitCandidates(template)
	if len(candidates) != 1 {
		t.Fatalf("expected only one portrait candidate, got %+v", candidates)
	}
	if candidates[0].SpeakerName != "Night Porter" {
		t.Fatalf("expected Night Porter to remain as the candidate, got %+v", candidates[0])
	}
	if !strings.Contains(candidates[0].Description, "Night Porter") {
		t.Fatalf("expected candidate description to include speaker name, got %q", candidates[0].Description)
	}
}

func TestApplyExpositionSpeakerPortraitURLsOnlyReplacesMissingPortraits(t *testing.T) {
	dialogue, changed := applyExpositionSpeakerPortraitURLs(
		models.DialogueSequence{
			{SpeakerName: "Night Porter", Text: "Keep your voice down.", PortraitURL: "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png"},
			{SpeakerName: "Harried Courier", Text: "Move.", PortraitURL: "https://example.com/existing.png"},
		},
		map[string]string{
			normalizeExpositionSpeakerPortraitKey("Night Porter"):    "https://example.com/night-porter.png",
			normalizeExpositionSpeakerPortraitKey("Harried Courier"): "https://example.com/new-courier.png",
		},
	)

	if !changed {
		t.Fatal("expected dialogue to be updated")
	}
	if dialogue[0].PortraitURL != "https://example.com/night-porter.png" {
		t.Fatalf("expected missing portrait to be filled, got %+v", dialogue[0])
	}
	if dialogue[1].PortraitURL != "https://example.com/existing.png" {
		t.Fatalf("expected existing portrait to be preserved, got %+v", dialogue[1])
	}
}

func TestShouldGeneratePortraitForExpositionSpeaker(t *testing.T) {
	if shouldGeneratePortraitForExpositionSpeaker("Witness Echo") {
		t.Fatal("expected abstract witness echo label to skip portrait generation")
	}
	if shouldGeneratePortraitForExpositionSpeaker("Overheard Couriers") {
		t.Fatal("expected collective overheard source label to skip portrait generation")
	}
	if !shouldGeneratePortraitForExpositionSpeaker("Street Witness") {
		t.Fatal("expected grounded role label to allow portrait generation")
	}
}
