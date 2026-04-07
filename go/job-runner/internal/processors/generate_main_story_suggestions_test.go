package processors

import "testing"

func TestMainStoryBeatNeedsExpansionForSingleScenarioStep(t *testing.T) {
	beat := mainStorySuggestionBeatPayload{
		ChapterTitle: "The False Saint's Sermon",
		Steps: []questArchetypeSuggestionStepPayload{
			{
				Source:            "location",
				Content:           "scenario",
				ScenarioPrompt:    "The false saint gives a sermon. What do you do?",
				ScenarioOpenEnded: true,
			},
		},
	}

	if !mainStoryBeatNeedsExpansion(beat) {
		t.Fatalf("expected single open-ended scenario beat to need expansion")
	}
}

func TestApplyMainStoryBeatFallbackArcBuildsMultiStepQuest(t *testing.T) {
	beat := mainStorySuggestionBeatPayload{
		OrderIndex:                     2,
		Act:                            1,
		StoryRole:                      "complication",
		ChapterTitle:                   "The False Saint's Sermon",
		ChapterSummary:                 "A weary dockside organizer slips into a packed sermon to spot who is being recruited.",
		Purpose:                        "Learn how the cult is approaching workers in public.",
		WhatChanges:                    "The player realizes the sermon is a recruitment funnel and the crowd includes planted enforcers.",
		Hook:                           "Blend into the crowd before the sermon turns dangerous.",
		Description:                    "",
		PreferredContentMix:            []string{"scenario", "challenge", "monster"},
		RequiredZoneTags:               []string{"nightlife", "market"},
		RequiredLocationArchetypeNames: []string{"Bustling City Square", "Back Alley"},
		InternalTags:                   []string{"occult", "false_saint"},
		Steps: []questArchetypeSuggestionStepPayload{
			{
				Source:            "location",
				Content:           "scenario",
				ScenarioPrompt:    "A false saint gives a sermon. What do you do?",
				ScenarioOpenEnded: true,
			},
		},
	}

	enriched := applyMainStoryBeatFallbackArc(beat)

	if len(enriched.Steps) < 2 {
		t.Fatalf("expected fallback arc to create at least 2 steps, got %d", len(enriched.Steps))
	}
	if enriched.Steps[0].Source == "proximity" {
		t.Fatalf("expected first fallback step to avoid proximity source")
	}
	if enriched.Steps[0].Content == "" || enriched.Steps[1].Content == "" {
		t.Fatalf("expected fallback steps to have content types")
	}
	if enriched.Description == "" {
		t.Fatalf("expected fallback arc to populate description")
	}
}
