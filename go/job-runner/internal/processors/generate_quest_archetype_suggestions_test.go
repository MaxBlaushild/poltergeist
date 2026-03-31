package processors

import "testing"

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
