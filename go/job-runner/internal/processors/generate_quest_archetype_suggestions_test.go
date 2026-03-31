package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestSanitizeQuestArchetypeSuggestionStepConvertsScenarioLikeChallenge(t *testing.T) {
	step, warnings := sanitizeQuestArchetypeSuggestionStep(
		questArchetypeSuggestionStepPayload{
			Source:                  "location",
			Content:                 "challenge",
			LocationConcept:         "bustling city square",
			LocationArchetypeName:   "Bustling City Square",
			LocationMetadataTags:    []string{"plaza", "crowded"},
			ChallengeQuestion:       "Where did the artist last see their sketchbook?",
			ChallengeDescription:    "Interview locals and search the area to find clues about the sketchbook's location.",
			ChallengeSubmissionType: "text",
		},
		map[string]locationArchetypeIndexEntry{
			"bustling city square": {Name: "Bustling City Square"},
		},
		map[string]monsterTemplateIndexEntry{},
	)

	if step.Content != "scenario" {
		t.Fatalf("expected scenario conversion, got %q", step.Content)
	}
	if !step.ScenarioOpenEnded {
		t.Fatalf("expected converted scenario to be open ended")
	}
	if step.ScenarioPrompt == "" {
		t.Fatalf("expected converted scenario prompt to be populated")
	}
	if step.ChallengeQuestion != "" || step.ChallengeDescription != "" {
		t.Fatalf("expected challenge fields to be cleared after conversion")
	}
	if len(warnings) == 0 {
		t.Fatalf("expected conversion warning")
	}
}

func TestSanitizeQuestArchetypeSuggestionStepKeepsConcreteChallenge(t *testing.T) {
	step, warnings := sanitizeQuestArchetypeSuggestionStep(
		questArchetypeSuggestionStepPayload{
			Source:                  "location",
			Content:                 "challenge",
			LocationConcept:         "lantern market",
			LocationArchetypeName:   "Lantern Market",
			LocationMetadataTags:    []string{"market", "signage"},
			ChallengeQuestion:       "Photograph the most elaborate hand-painted sign you can find.",
			ChallengeDescription:    "Capture a sign whose colors and lettering feel unmistakably local.",
			ChallengeSubmissionType: "photo",
		},
		map[string]locationArchetypeIndexEntry{
			"lantern market": {Name: "Lantern Market"},
		},
		map[string]monsterTemplateIndexEntry{},
	)

	if step.Content != "challenge" {
		t.Fatalf("expected concrete challenge to stay a challenge, got %q", step.Content)
	}
	if step.ScenarioPrompt != "" {
		t.Fatalf("expected no scenario prompt for concrete challenge")
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for concrete challenge, got %v", warnings)
	}
}

func TestSanitizeQuestArchetypeSuggestionDraftWarnsWhenRequiredLocationArchetypeMissing(t *testing.T) {
	usedID := uuid.New()
	missingID := uuid.New()
	draft := sanitizeQuestArchetypeSuggestionDraft(
		questArchetypeSuggestionDraftPayload{
			Name:        "Night Courier Run",
			Description: "Move a parcel through the district.",
			Steps: []questArchetypeSuggestionStepPayload{
				{
					Source:                "location",
					Content:               "challenge",
					LocationConcept:       "market lane",
					LocationArchetypeName: "Lantern Market",
					LocationMetadataTags:  []string{"market", "signage"},
					TemplateConcept:       "Document the delivery marker",
					ChallengeQuestion:     "Photograph the lantern-marked drop sign.",
					ChallengeDescription:  "Capture the sign that would guide a courier to the handoff point.",
				},
			},
		},
		map[string]locationArchetypeIndexEntry{
			"lantern market": {ID: usedID, Name: "Lantern Market"},
			"roof garden":    {ID: missingID, Name: "Roof Garden"},
		},
		map[string]monsterTemplateIndexEntry{},
		[]locationArchetypeIndexEntry{
			{ID: usedID, Name: "Lantern Market"},
			{ID: missingID, Name: "Roof Garden"},
		},
	)

	found := false
	for _, warning := range draft.Warnings {
		if warning == `required location archetype "Roof Garden" was not used in this draft` {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected missing required location warning, got %v", draft.Warnings)
	}
}

func TestMissingRequiredLocationArchetypes(t *testing.T) {
	usedID := uuid.New()
	missingID := uuid.New()
	missing := missingRequiredLocationArchetypes(
		models.QuestArchetypeSuggestionSteps{
			{
				LocationArchetypeID: &usedID,
			},
		},
		[]locationArchetypeIndexEntry{
			{ID: usedID, Name: "Lantern Market"},
			{ID: missingID, Name: "Riverside Shrine"},
		},
	)

	if len(missing) != 1 || missing[0] != "Riverside Shrine" {
		t.Fatalf("unexpected missing required archetypes: %v", missing)
	}
}
