package processors

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func testQuestArchetypeSuggestionDraft(
	name string,
	steps []models.QuestArchetypeSuggestionStep,
	warnings []string,
) *models.QuestArchetypeSuggestionDraft {
	return &models.QuestArchetypeSuggestionDraft{
		Name:               name,
		Hook:               name + " hook",
		Description:        name + " description",
		WhyThisScales:      "Escalates with player pressure.",
		AcceptanceDialogue: models.StringArray{"Line one", "Line two"},
		Steps:              steps,
		Warnings:           models.StringArray(warnings),
	}
}

func suggestionDraftHasWarning(
	draft *models.QuestArchetypeSuggestionDraft,
	expected string,
) bool {
	if draft == nil {
		return false
	}
	for _, warning := range draft.Warnings {
		if warning == expected {
			return true
		}
	}
	return false
}

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

func TestSanitizeQuestArchetypeSuggestionStepBridgesFantasyLiteralChallengeToReality(t *testing.T) {
	step, warnings := sanitizeQuestArchetypeSuggestionStep(
		questArchetypeSuggestionStepPayload{
			Source:                  "location",
			Content:                 "challenge",
			LocationConcept:         "cozy book nook",
			LocationArchetypeName:   "Cozy Book Nook",
			LocationMetadataTags:    []string{"scholarly", "mystical", "secretive"},
			TemplateConcept:         "identify the hidden sigil",
			ChallengeQuestion:       "Photograph the hidden sigil within the tome.",
			ChallengeDescription:    "Look for arcane symbols or hidden runes within the tome's pages. Capture a photo of the sigil.",
			ChallengeSubmissionType: "photo",
		},
		map[string]locationArchetypeIndexEntry{
			"cozy book nook": {Name: "Cozy Book Nook"},
		},
		map[string]monsterTemplateIndexEntry{},
	)

	if step.Content != "challenge" {
		t.Fatalf("expected fantasy-literal task to remain a challenge, got %q", step.Content)
	}
	if step.ChallengeQuestion != "Photograph what looks like a hidden sigil within a book of your choice." {
		t.Fatalf("expected question to bridge back to a real-world proxy, got %q", step.ChallengeQuestion)
	}
	if !strings.Contains(step.ChallengeDescription, "Use something actually present on site") {
		t.Fatalf("expected description to clarify the real-world stand-in, got %q", step.ChallengeDescription)
	}
	if !strings.Contains(step.ChallengeDescription, "could pass for a hidden sigil within a book of your choice") {
		t.Fatalf("expected description to preserve flavorful target with a proxy, got %q", step.ChallengeDescription)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for a bridged concrete challenge, got %v", warnings)
	}
}

func TestSanitizeQuestArchetypeSuggestionStepExpandsSparseScenarioPrompt(t *testing.T) {
	step, _ := sanitizeQuestArchetypeSuggestionStep(
		questArchetypeSuggestionStepPayload{
			Source:                "location",
			Content:               "scenario",
			LocationConcept:       "moonlit observatory",
			LocationArchetypeName: "Moonlit Observatory",
			LocationMetadataTags:  []string{"ritual_site", "storm_struck"},
			TemplateConcept:       "Interrupt the runaway rite before the district tears open",
			ScenarioPrompt:        "The ritual is underway. How do you stop it?",
			ScenarioBeats: []string{
				"Blue fire is crawling up the cracked brass lenses",
				"Masked chanters are forcing terrified onlookers to keep the circle intact",
			},
		},
		map[string]locationArchetypeIndexEntry{
			"moonlit observatory": {Name: "Moonlit Observatory"},
		},
		map[string]monsterTemplateIndexEntry{},
	)

	if !strings.Contains(step.ScenarioPrompt, "At the moonlit observatory,") {
		t.Fatalf("expected expanded prompt to anchor the location, got %q", step.ScenarioPrompt)
	}
	if !strings.Contains(step.ScenarioPrompt, "Blue fire is crawling up the cracked brass lenses.") {
		t.Fatalf("expected expanded prompt to include a vivid beat, got %q", step.ScenarioPrompt)
	}
	if !strings.Contains(step.ScenarioPrompt, "How do you stop it?") {
		t.Fatalf("expected expanded prompt to preserve the player-facing problem, got %q", step.ScenarioPrompt)
	}
	if suggestionScenarioSentenceCount(step.ScenarioPrompt) < 3 {
		t.Fatalf("expected expanded prompt to have multiple scene-setting sentences, got %q", step.ScenarioPrompt)
	}
}

func TestSanitizeQuestArchetypeSuggestionStepKeepsDetailedScenarioPrompt(t *testing.T) {
	original := "At the rooftop shrine, copper bells are screaming in the wind while masked chanters pull sparks out of the wardstones. Every flare is waking more of the district's ghosts, and nearby worshippers are starting to panic. How do you stop it?"
	step, _ := sanitizeQuestArchetypeSuggestionStep(
		questArchetypeSuggestionStepPayload{
			Source:                "location",
			Content:               "scenario",
			LocationConcept:       "rooftop shrine",
			LocationArchetypeName: "Rooftop Shrine",
			LocationMetadataTags:  []string{"shrine", "storm"},
			TemplateConcept:       "Break the rooftop rite",
			ScenarioPrompt:        original,
			ScenarioBeats: []string{
				"Copper bells scream in the wind",
				"Nearby worshippers panic as ghosts stir awake",
			},
		},
		map[string]locationArchetypeIndexEntry{
			"rooftop shrine": {Name: "Rooftop Shrine"},
		},
		map[string]monsterTemplateIndexEntry{},
	)

	if step.ScenarioPrompt != original {
		t.Fatalf("expected detailed scenario prompt to remain unchanged, got %q", step.ScenarioPrompt)
	}
}

func TestSanitizeQuestArchetypeSuggestionStepBuildsExpositionFromSparseFields(t *testing.T) {
	locationID := uuid.New()
	step, warnings := sanitizeQuestArchetypeSuggestionStep(
		questArchetypeSuggestionStepPayload{
			Source:                "proximity",
			Content:               "exposition",
			LocationConcept:       "cozy book nook",
			LocationArchetypeName: "Cozy Book Nook",
			LocationMetadataTags:  []string{"scholarly", "mystical", "secretive"},
			TemplateConcept:       "read the margin warning",
			PotentialContent: []string{
				"charred notes in the margins",
				"the warning feels freshly relevant",
			},
		},
		map[string]locationArchetypeIndexEntry{
			"cozy book nook": {ID: locationID, Name: "Cozy Book Nook"},
		},
		map[string]monsterTemplateIndexEntry{},
	)

	if step.Content != "exposition" {
		t.Fatalf("expected exposition content, got %q", step.Content)
	}
	if step.Source != "location" {
		t.Fatalf("expected exposition step to be coerced to location, got %q", step.Source)
	}
	if step.LocationArchetypeID == nil || *step.LocationArchetypeID != locationID {
		t.Fatalf("expected resolved location archetype id, got %+v", step.LocationArchetypeID)
	}
	if step.ExpositionTitle != "Read The Margin Warning" {
		t.Fatalf("expected generated exposition title, got %q", step.ExpositionTitle)
	}
	if !strings.Contains(step.ExpositionDescription, "At the cozy book nook,") {
		t.Fatalf("expected generated exposition description to anchor the location, got %q", step.ExpositionDescription)
	}
	if len(step.ExpositionDialogue) < 2 {
		t.Fatalf("expected exposition dialogue fallback lines, got %+v", step.ExpositionDialogue)
	}
	if strings.TrimSpace(step.ExpositionSpeakerName) == "" {
		t.Fatalf("expected generated exposition speaker name, got %+v", step)
	}
	if step.ExpositionPortraitURL != "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png" {
		t.Fatalf("expected generated exposition portrait placeholder, got %q", step.ExpositionPortraitURL)
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warnings for sparse exposition fields")
	}
}

func TestSanitizeQuestArchetypeSuggestionDraftExpandsNarrativeFields(t *testing.T) {
	draft := sanitizeQuestArchetypeSuggestionDraft(
		questArchetypeSuggestionDraftPayload{
			Name:               "Observatory Breach",
			Hook:               "Trouble is brewing.",
			Description:        "A tense situation unfolds.",
			WhyThisScales:      "It scales because it can happen anywhere.",
			AcceptanceDialogue: []string{},
			Nodes: []questArchetypeSuggestionNodePayload{
				{
					NodeKey: "entry",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:                "location",
						Content:               "scenario",
						LocationConcept:       "moonlit observatory",
						LocationArchetypeName: "Moonlit Observatory",
						LocationMetadataTags:  []string{"ritual_site", "storm_struck"},
						TemplateConcept:       "Interrupt the runaway rite before the district tears open",
						ScenarioPrompt:        "The ritual is underway. How do you stop it?",
						ScenarioBeats: []string{
							"Blue fire is crawling up the cracked brass lenses",
							"Masked chanters are forcing terrified onlookers to keep the circle intact",
						},
						PotentialContent: []string{
							"The ward-lines are already snapping and spilling sparks into the rain",
						},
					},
					Outcomes: []questArchetypeSuggestionOutcomePayload{
						{Outcome: "success", NextNodeKey: "aftermath"},
						{Outcome: "failure", NextNodeKey: "aftermath"},
					},
				},
				{
					NodeKey: "aftermath",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "monster",
						LocationConcept:      "observatory stairs",
						LocationMetadataTags: []string{"stairs", "lightning_scars"},
						TemplateConcept:      "Fight through the thing the rite wakes up",
					},
				},
			},
		},
		"city",
		map[string]locationArchetypeIndexEntry{
			"moonlit observatory": {Name: "Moonlit Observatory"},
		},
		map[string]monsterTemplateIndexEntry{},
		nil,
	)

	if !strings.Contains(draft.Description, "At the moonlit observatory,") {
		t.Fatalf("expected description to anchor the scene, got %q", draft.Description)
	}
	if !strings.Contains(draft.Description, "Blue fire is crawling up the cracked brass lenses.") {
		t.Fatalf("expected description to include vivid detail, got %q", draft.Description)
	}
	if draft.Hook != suggestionFirstSentence(draft.Description) {
		t.Fatalf("expected hook to align with the opening narrative beat, got %q / %q", draft.Hook, draft.Description)
	}
	if len(draft.AcceptanceDialogue) < 3 {
		t.Fatalf("expected acceptance dialogue to be synthesized, got %v", draft.AcceptanceDialogue)
	}
	if !strings.Contains(strings.ToLower(draft.AcceptanceDialogue[0]), "moonlit observatory") {
		t.Fatalf("expected acceptance dialogue to carry scene context, got %v", draft.AcceptanceDialogue)
	}
	if !strings.Contains(draft.WhyThisScales, "This premise scales cleanly because") {
		t.Fatalf("expected synthesized scaling explanation, got %q", draft.WhyThisScales)
	}
}

func TestSanitizeQuestArchetypeSuggestionDraftKeepsDetailedNarrativeFields(t *testing.T) {
	description := "At the floodlit market gate, rain hisses off fresh ward-paint while rival couriers argue over a package neither side trusts. Onlookers are starting to choose sides, and every raised voice is drawing more attention to the contraband in plain sight. If nobody cuts through the panic, the whole handoff is going to collapse into a public scandal."
	hook := "At the floodlit market gate, rain hisses off fresh ward-paint while rival couriers argue over a package neither side trusts."
	whyThisScales := "This premise scales cleanly because the same kind of public handoff can erupt at markets, checkpoints, or transit hubs, while the social pressure and factional consequences can widen as the stakes rise."
	dialogue := []string{
		"At the floodlit market gate, rain is hissing off fresh ward-paint while two courier crews tear into each other over a package neither side trusts.",
		"People are already choosing sides, and if this turns public we lose the package and the route behind it.",
		"Get in there, calm it down, and make sure the handoff doesn't become a spectacle.",
	}
	draft := sanitizeQuestArchetypeSuggestionDraft(
		questArchetypeSuggestionDraftPayload{
			Name:               "Floodlit Handoff",
			Hook:               hook,
			Description:        description,
			WhyThisScales:      whyThisScales,
			AcceptanceDialogue: dialogue,
			Steps: []questArchetypeSuggestionStepPayload{
				{
					Source:                "location",
					Content:               "scenario",
					LocationConcept:       "floodlit market gate",
					LocationArchetypeName: "Floodlit Market Gate",
					LocationMetadataTags:  []string{"market", "checkpoint"},
					TemplateConcept:       "Keep the handoff from becoming a scandal",
					ScenarioPrompt:        description,
				},
			},
		},
		"city",
		map[string]locationArchetypeIndexEntry{
			"floodlit market gate": {Name: "Floodlit Market Gate"},
		},
		map[string]monsterTemplateIndexEntry{},
		nil,
	)

	if draft.Hook != hook {
		t.Fatalf("expected detailed hook to be preserved, got %q", draft.Hook)
	}
	if draft.Description != description {
		t.Fatalf("expected detailed description to be preserved, got %q", draft.Description)
	}
	if draft.WhyThisScales != whyThisScales {
		t.Fatalf("expected detailed whyThisScales to be preserved, got %q", draft.WhyThisScales)
	}
	if strings.Join(draft.AcceptanceDialogue, " | ") != strings.Join(dialogue, " | ") {
		t.Fatalf("expected detailed acceptance dialogue to be preserved, got %v", draft.AcceptanceDialogue)
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
		"",
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

func TestSanitizeQuestArchetypeSuggestionDraftNormalizesZoneKind(t *testing.T) {
	draft := sanitizeQuestArchetypeSuggestionDraft(
		questArchetypeSuggestionDraftPayload{
			Name:        "Glacier Relay",
			Description: "Carry messages through an icy ward.",
			Steps: []questArchetypeSuggestionStepPayload{
				{
					Source:               "location",
					Content:              "challenge",
					LocationConcept:      "watch post",
					LocationMetadataTags: []string{"frost", "outpost"},
					ChallengeQuestion:    "Photograph the signal marker you would use to confirm the route.",
					ChallengeDescription: "Capture the route marker that would guide the next courier.",
				},
			},
		},
		"Tundra",
		map[string]locationArchetypeIndexEntry{},
		map[string]monsterTemplateIndexEntry{},
		nil,
	)

	if draft.ZoneKind != "tundra" {
		t.Fatalf("expected normalized zone kind, got %q", draft.ZoneKind)
	}
}

func TestSanitizeQuestArchetypeSuggestionDraftDefaultsMissingSuccessBranch(t *testing.T) {
	draft := sanitizeQuestArchetypeSuggestionDraft(
		questArchetypeSuggestionDraftPayload{
			Name:        "Broken Relay",
			Description: "Carry intel through the district.",
			Nodes: []questArchetypeSuggestionNodePayload{
				{
					NodeKey: "entry",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "challenge",
						LocationConcept:      "signal tower",
						LocationMetadataTags: []string{"tower", "relay"},
						TemplateConcept:      "Mark the tower sigil",
						ChallengeQuestion:    "Photograph the signal marker at the tower base.",
						ChallengeDescription: "Capture the marker that confirms the route is still open.",
					},
				},
				{
					NodeKey: "handoff",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "scenario",
						LocationConcept:      "courier alcove",
						LocationMetadataTags: []string{"courier", "alcove"},
						TemplateConcept:      "Handle the tense handoff",
						ScenarioPrompt:       "A courier demands proof before accepting the package. How do you respond?",
					},
				},
			},
		},
		"",
		map[string]locationArchetypeIndexEntry{},
		map[string]monsterTemplateIndexEntry{},
		nil,
	)

	if len(draft.Nodes) != 2 {
		t.Fatalf("expected two nodes, got %d", len(draft.Nodes))
	}
	if len(draft.Nodes[0].Outcomes) != 1 {
		t.Fatalf("expected one defaulted outcome, got %d", len(draft.Nodes[0].Outcomes))
	}
	if draft.Nodes[0].Outcomes[0].Outcome != questArchetypeSuggestionOutcomeSuccess {
		t.Fatalf("expected success outcome, got %q", draft.Nodes[0].Outcomes[0].Outcome)
	}
	if draft.Nodes[0].Outcomes[0].NextNodeKey != "handoff" {
		t.Fatalf("expected success branch to handoff, got %q", draft.Nodes[0].Outcomes[0].NextNodeKey)
	}
	if !suggestionDraftHasWarning(
		draft,
		`node 1: success branch was missing and defaulted to node "handoff"`,
	) {
		t.Fatalf("expected missing-success warning, got %v", draft.Warnings)
	}
}

func TestSanitizeQuestArchetypeSuggestionDraftDropsBackwardEdge(t *testing.T) {
	draft := sanitizeQuestArchetypeSuggestionDraft(
		questArchetypeSuggestionDraftPayload{
			Name:        "Looping Route",
			Description: "Try to push a route backward.",
			Nodes: []questArchetypeSuggestionNodePayload{
				{
					NodeKey: "entry",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "challenge",
						LocationConcept:      "checkpoint",
						LocationMetadataTags: []string{"checkpoint"},
						TemplateConcept:      "Mark the route start",
						ChallengeQuestion:    "Photograph the route marker.",
						ChallengeDescription: "Capture the checkpoint marker that opens the route.",
					},
					Outcomes: []questArchetypeSuggestionOutcomePayload{
						{Outcome: "success", NextNodeKey: "pivot"},
					},
				},
				{
					NodeKey: "pivot",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "scenario",
						LocationConcept:      "waystation",
						LocationMetadataTags: []string{"waystation"},
						TemplateConcept:      "Negotiate passage",
						ScenarioPrompt:       "A guard blocks the way unless you can convince them.",
					},
					Outcomes: []questArchetypeSuggestionOutcomePayload{
						{Outcome: "success", NextNodeKey: "entry"},
					},
				},
			},
		},
		"",
		map[string]locationArchetypeIndexEntry{},
		map[string]monsterTemplateIndexEntry{},
		nil,
	)

	if len(draft.Nodes) != 2 {
		t.Fatalf("expected two nodes, got %d", len(draft.Nodes))
	}
	if len(draft.Nodes[1].Outcomes) != 0 {
		t.Fatalf("expected backward edge to be removed, got %v", draft.Nodes[1].Outcomes)
	}
	if !suggestionDraftHasWarning(
		draft,
		`node 2: success branch pointed backward to "entry" and was dropped`,
	) {
		t.Fatalf("expected backward-edge warning, got %v", draft.Warnings)
	}
}

func TestSanitizeQuestArchetypeSuggestionDraftKeepsTerminalBranchNodeTerminal(t *testing.T) {
	draft := sanitizeQuestArchetypeSuggestionDraft(
		questArchetypeSuggestionDraftPayload{
			Name:        "Fail-Forward Courier Route",
			Description: "A success path and a fallback path should not merge by accident.",
			Nodes: []questArchetypeSuggestionNodePayload{
				{
					NodeKey: "entry",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "challenge",
						LocationConcept:      "signal bridge",
						LocationMetadataTags: []string{"bridge", "relay"},
						TemplateConcept:      "Confirm the signal mark",
						ChallengeQuestion:    "Photograph the courier mark on the bridge.",
						ChallengeDescription: "Capture proof that the route is still open.",
					},
					Outcomes: []questArchetypeSuggestionOutcomePayload{
						{Outcome: "success", NextNodeKey: "handoff"},
						{Outcome: "failure", NextNodeKey: "fallback"},
					},
				},
				{
					NodeKey: "handoff",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "scenario",
						LocationConcept:      "courier balcony",
						LocationMetadataTags: []string{"balcony"},
						TemplateConcept:      "Handle the rooftop exchange",
						ScenarioPrompt:       "The courier demands a faster deal than expected. What do you do?",
					},
				},
				{
					NodeKey: "fallback",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "monster",
						LocationConcept:      "underbridge stairs",
						LocationMetadataTags: []string{"stairs"},
						TemplateConcept:      "Fight through the backup ambush",
					},
				},
			},
		},
		"",
		map[string]locationArchetypeIndexEntry{},
		map[string]monsterTemplateIndexEntry{},
		nil,
	)

	if len(draft.Nodes) != 3 {
		t.Fatalf("expected three connected nodes, got %d", len(draft.Nodes))
	}
	if len(draft.Nodes[1].Outcomes) != 0 {
		t.Fatalf("expected success branch node to remain terminal, got %v", draft.Nodes[1].Outcomes)
	}
	if len(draft.Nodes[2].Outcomes) != 0 {
		t.Fatalf("expected fallback node to remain terminal, got %v", draft.Nodes[2].Outcomes)
	}
	if suggestionDraftHasWarning(
		draft,
		`node 2: success branch was missing and defaulted to node "fallback"`,
	) {
		t.Fatalf("did not expect synthetic branch merge warning, got %v", draft.Warnings)
	}
}

func TestSelectQuestArchetypeSuggestionLocationArchetypesForPromptKeepsRequiredUnderCap(t *testing.T) {
	requiredID := uuid.New()
	otherAID := uuid.New()
	otherBID := uuid.New()

	selected := selectQuestArchetypeSuggestionLocationArchetypesForPrompt(
		[]*models.LocationArchetype{
			{ID: otherBID, Name: "Zoo"},
			{ID: requiredID, Name: "Arcade"},
			{ID: otherAID, Name: "Bakery"},
		},
		[]string{requiredID.String()},
		2,
	)

	if len(selected) != 2 {
		t.Fatalf("expected 2 selected archetypes, got %d", len(selected))
	}
	if selected[0] == nil || selected[0].ID != requiredID {
		t.Fatalf("expected required archetype to be kept first, got %#v", selected[0])
	}
	if selected[1] == nil || selected[1].ID != otherAID {
		t.Fatalf("expected alphabetical supplemental archetype, got %#v", selected[1])
	}
}

func TestSelectQuestArchetypeSuggestionMonsterTemplatesForPromptPrefersZoneMatches(t *testing.T) {
	selected := selectQuestArchetypeSuggestionMonsterTemplatesForPrompt(
		[]models.MonsterTemplate{
			{ID: uuid.New(), Name: "Bridge Wisp", ZoneKind: "", MonsterType: models.MonsterTemplateTypeMonster},
			{ID: uuid.New(), Name: "City Watcher", ZoneKind: "city", MonsterType: models.MonsterTemplateTypeMonster},
			{ID: uuid.New(), Name: "Harbor Brute", ZoneKind: "harbor", MonsterType: models.MonsterTemplateTypeMonster},
			{ID: uuid.New(), Name: "City Tyrant", ZoneKind: "city", MonsterType: models.MonsterTemplateTypeBoss},
		},
		"city",
		3,
	)

	if len(selected) != 3 {
		t.Fatalf("expected 3 selected monster templates, got %d", len(selected))
	}
	if selected[0].Name != "City Watcher" {
		t.Fatalf("expected city monster match first, got %q", selected[0].Name)
	}
	if selected[1].Name != "City Tyrant" {
		t.Fatalf("expected city boss match second, got %q", selected[1].Name)
	}
	if selected[2].Name != "Bridge Wisp" {
		t.Fatalf("expected generic monster before unrelated zone, got %q", selected[2].Name)
	}
}

func TestSanitizeQuestArchetypeSuggestionDraftDropsUnreachableNode(t *testing.T) {
	draft := sanitizeQuestArchetypeSuggestionDraft(
		questArchetypeSuggestionDraftPayload{
			Name:        "Forked Courier Route",
			Description: "Only connected nodes should survive.",
			Nodes: []questArchetypeSuggestionNodePayload{
				{
					NodeKey: "entry",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "challenge",
						LocationConcept:      "canal bridge",
						LocationMetadataTags: []string{"bridge"},
						TemplateConcept:      "Confirm the courier mark",
						ChallengeQuestion:    "Photograph the courier chalk mark on the bridge.",
						ChallengeDescription: "Capture the proof that the route still crosses the canal here.",
					},
					Outcomes: []questArchetypeSuggestionOutcomePayload{
						{Outcome: "success", NextNodeKey: "handoff"},
					},
				},
				{
					NodeKey: "handoff",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "monster",
						LocationConcept:      "canal bridge",
						LocationMetadataTags: []string{"bridge"},
						TemplateConcept:      "Beat back the bridge stalker",
					},
				},
				{
					NodeKey: "orphan",
					questArchetypeSuggestionStepPayload: questArchetypeSuggestionStepPayload{
						Source:               "location",
						Content:              "scenario",
						LocationConcept:      "hidden cellar",
						LocationMetadataTags: []string{"cellar"},
						TemplateConcept:      "Handle the secret meeting",
						ScenarioPrompt:       "Two smugglers offer conflicting deals. What do you do?",
					},
				},
			},
		},
		"",
		map[string]locationArchetypeIndexEntry{},
		map[string]monsterTemplateIndexEntry{},
		nil,
	)

	if len(draft.Nodes) != 2 {
		t.Fatalf("expected unreachable node to be removed, got %d nodes", len(draft.Nodes))
	}
	if draft.Nodes[0].NodeKey != "entry" || draft.Nodes[1].NodeKey != "handoff" {
		t.Fatalf("unexpected surviving node order: %+v", draft.Nodes)
	}
	if len(draft.Steps) != 2 {
		t.Fatalf("expected legacy steps to mirror reachable nodes, got %d", len(draft.Steps))
	}
	if !suggestionDraftHasWarning(
		draft,
		`node 3: node "orphan" was unreachable from the root and was dropped`,
	) {
		t.Fatalf("expected unreachable-node warning, got %v", draft.Warnings)
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

func TestSelectQuestArchetypeSuggestionDraftsPrefersDraftsWithoutMissingRequiredArchetypes(t *testing.T) {
	missingRequired := testQuestArchetypeSuggestionDraft(
		"Missing Required",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "market lane",
				LocationArchetypeName: "Lantern Market",
				TemplateConcept:       "Document the market sigil",
			},
			{
				Source:          "location",
				Content:         "monster",
				LocationConcept: "market lane",
				TemplateConcept: "Scare off the stalker",
			},
		},
		[]string{`required location archetype "Roof Garden" was not used in this draft`},
	)
	valid := testQuestArchetypeSuggestionDraft(
		"Valid Coverage",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "roof garden",
				LocationArchetypeName: "Roof Garden",
				TemplateConcept:       "Read the wind charms",
			},
			{
				Source:          "location",
				Content:         "scenario",
				LocationConcept: "market lane",
				TemplateConcept: "Negotiate the handoff",
			},
		},
		nil,
	)

	selected := selectQuestArchetypeSuggestionDrafts(
		[]*models.QuestArchetypeSuggestionDraft{missingRequired, valid},
		nil,
		1,
		nil,
	)
	if len(selected) != 1 {
		t.Fatalf("expected one selected draft, got %d", len(selected))
	}
	if selected[0].Name != "Valid Coverage" {
		t.Fatalf("expected valid coverage draft, got %q", selected[0].Name)
	}
}

func TestSelectQuestArchetypeSuggestionDraftsKeepsBatchMixWhenAvailable(t *testing.T) {
	nonCombat := testQuestArchetypeSuggestionDraft(
		"Lantern Mediation",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "lantern market",
				LocationArchetypeName: "Lantern Market",
				TemplateConcept:       "Photograph the trade sigil",
			},
			{
				Source:          "location",
				Content:         "scenario",
				LocationConcept: "teahouse",
				TemplateConcept: "Mediate a tense exchange",
			},
			{
				Source:          "proximity",
				Content:         "scenario",
				LocationConcept: "alley crossing",
				TemplateConcept: "Choose how to move the goods",
			},
		},
		nil,
	)
	combatA := testQuestArchetypeSuggestionDraft(
		"Canal Ambush",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "canal lock",
				LocationArchetypeName: "Canal Lock",
				TemplateConcept:       "Mark the warning sign",
			},
			{
				Source:          "location",
				Content:         "monster",
				LocationConcept: "canal lock",
				TemplateConcept: "Beat back the canal stalker",
			},
		},
		nil,
	)
	combatB := testQuestArchetypeSuggestionDraft(
		"Bridge Ambush",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "old bridge",
				LocationArchetypeName: "Old Bridge",
				TemplateConcept:       "Mark the warning sign",
			},
			{
				Source:          "location",
				Content:         "monster",
				LocationConcept: "old bridge",
				TemplateConcept: "Beat back the bridge stalker",
			},
		},
		nil,
	)
	combatC := testQuestArchetypeSuggestionDraft(
		"Harbor Clash",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "scenario",
				LocationConcept:       "harbor office",
				LocationArchetypeName: "Harbor Office",
				TemplateConcept:       "Decide which captain to trust",
			},
			{
				Source:          "location",
				Content:         "monster",
				LocationConcept: "harbor office",
				TemplateConcept: "Repel the harbor enforcer",
			},
		},
		nil,
	)

	selected := selectQuestArchetypeSuggestionDrafts(
		[]*models.QuestArchetypeSuggestionDraft{combatA, combatB, combatC, nonCombat},
		nil,
		3,
		nil,
	)
	if len(selected) != 3 {
		t.Fatalf("expected three selected drafts, got %d", len(selected))
	}
	foundNonCombat := false
	for _, draft := range selected {
		if draft.Name == "Lantern Mediation" {
			foundNonCombat = true
			break
		}
	}
	if !foundNonCombat {
		t.Fatalf("expected non-combat draft to be preserved in the batch mix")
	}
}

func TestSelectQuestArchetypeSuggestionDraftsSkipsNearDuplicatesWhenAlternativesExist(t *testing.T) {
	duplicateA := testQuestArchetypeSuggestionDraft(
		"Signal Run",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "signal tower",
				LocationArchetypeName: "Signal Tower",
				TemplateConcept:       "Photograph the beacon marker",
			},
			{
				Source:          "location",
				Content:         "monster",
				LocationConcept: "signal tower",
				TemplateConcept: "Drive off the rooftop stalker",
			},
		},
		nil,
	)
	duplicateB := testQuestArchetypeSuggestionDraft(
		"Signal Run Variant",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "signal tower",
				LocationArchetypeName: "Signal Tower",
				TemplateConcept:       "Photograph the beacon marker",
			},
			{
				Source:          "location",
				Content:         "monster",
				LocationConcept: "signal tower",
				TemplateConcept: "Drive off the rooftop stalker",
			},
		},
		nil,
	)
	distinct := testQuestArchetypeSuggestionDraft(
		"Harbor Bargain",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "scenario",
				LocationConcept:       "harbor gate",
				LocationArchetypeName: "Harbor Gate",
				TemplateConcept:       "Broker a tense exchange",
			},
			{
				Source:          "proximity",
				Content:         "scenario",
				LocationConcept: "warehouse row",
				TemplateConcept: "Decide where to stash the cargo",
			},
		},
		nil,
	)

	selected := selectQuestArchetypeSuggestionDrafts(
		[]*models.QuestArchetypeSuggestionDraft{duplicateA, duplicateB, distinct},
		nil,
		2,
		nil,
	)
	if len(selected) != 2 {
		t.Fatalf("expected two selected drafts, got %d", len(selected))
	}
	duplicateCount := 0
	for _, draft := range selected {
		if draft.Name == "Signal Run" || draft.Name == "Signal Run Variant" {
			duplicateCount++
		}
	}
	if duplicateCount != 1 {
		t.Fatalf("expected exactly one signal-run duplicate, got %d", duplicateCount)
	}
}

func TestSelectQuestArchetypeSuggestionDraftsHonorsFamilyMixTargets(t *testing.T) {
	investigation := testQuestArchetypeSuggestionDraft(
		"Signal Inquiry",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "signal tower",
				LocationArchetypeName: "Signal Tower",
				TemplateConcept:       "Document the suspicious signal pattern",
			},
			{
				Source:          "location",
				Content:         "scenario",
				LocationConcept: "tower office",
				TemplateConcept: "Investigate who tampered with the beacon log",
			},
		},
		nil,
	)
	investigation.InternalTags = models.StringArray{"investigation", "signal"}

	negotiation := testQuestArchetypeSuggestionDraft(
		"Dockside Truce",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "harbor stairs",
				LocationArchetypeName: "Harbor Stairs",
				TemplateConcept:       "Record the rival faction markers",
			},
			{
				Source:          "location",
				Content:         "scenario",
				LocationConcept: "harbor office",
				TemplateConcept: "Negotiate a fragile truce between couriers",
			},
		},
		nil,
	)
	negotiation.InternalTags = models.StringArray{"negotiation", "harbor"}

	delivery := testQuestArchetypeSuggestionDraft(
		"Parcel Relay",
		[]models.QuestArchetypeSuggestionStep{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "market lane",
				LocationArchetypeName: "Market Lane",
				TemplateConcept:       "Confirm the drop marker",
			},
			{
				Source:          "location",
				Content:         "scenario",
				LocationConcept: "courier vault",
				TemplateConcept: "Deliver the package before the route closes",
			},
		},
		nil,
	)
	delivery.InternalTags = models.StringArray{"delivery", "market"}

	selected := selectQuestArchetypeSuggestionDrafts(
		[]*models.QuestArchetypeSuggestionDraft{delivery, investigation, negotiation},
		nil,
		2,
		models.QuestArchetypeSuggestionFamilyMixTargets{
			"investigation": 1,
			"negotiation":   1,
		},
	)
	if len(selected) != 2 {
		t.Fatalf("expected two selected drafts, got %d", len(selected))
	}
	selectedNames := map[string]struct{}{}
	for _, draft := range selected {
		selectedNames[draft.Name] = struct{}{}
	}
	if _, ok := selectedNames["Signal Inquiry"]; !ok {
		t.Fatalf("expected investigation draft to be selected, got %v", selectedNames)
	}
	if _, ok := selectedNames["Dockside Truce"]; !ok {
		t.Fatalf("expected negotiation draft to be selected, got %v", selectedNames)
	}
}

func TestQuestArchetypeSuggestionDraftRouteKeyIncludesBranchShape(t *testing.T) {
	linear := &models.QuestArchetypeSuggestionDraft{
		Name: "Courier Route",
		Steps: models.QuestArchetypeSuggestionSteps{
			{
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "signal bridge",
				LocationArchetypeName: "Signal Bridge",
				TemplateConcept:       "Mark the bridge route",
			},
			{
				Source:          "location",
				Content:         "monster",
				LocationConcept: "underbridge stairs",
				TemplateConcept: "Fight the ambush",
			},
		},
	}
	branched := &models.QuestArchetypeSuggestionDraft{
		Name: "Courier Route Branched",
		Nodes: models.QuestArchetypeSuggestionNodes{
			{
				NodeKey:               "entry",
				Source:                "location",
				Content:               "challenge",
				LocationConcept:       "signal bridge",
				LocationArchetypeName: "Signal Bridge",
				TemplateConcept:       "Mark the bridge route",
				Outcomes: models.QuestArchetypeSuggestionNodeOutcomes{
					{Outcome: "success", NextNodeKey: "handoff"},
					{Outcome: "failure", NextNodeKey: "ambush"},
				},
			},
			{
				NodeKey:         "handoff",
				Source:          "location",
				Content:         "scenario",
				LocationConcept: "courier balcony",
				TemplateConcept: "Handle the rooftop exchange",
			},
			{
				NodeKey:         "ambush",
				Source:          "location",
				Content:         "monster",
				LocationConcept: "underbridge stairs",
				TemplateConcept: "Fight the ambush",
			},
		},
	}

	linearKey := questArchetypeSuggestionDraftRouteKey(linear)
	branchedKey := questArchetypeSuggestionDraftRouteKey(branched)
	if linearKey == "" || branchedKey == "" {
		t.Fatalf("expected non-empty route keys, got linear=%q branched=%q", linearKey, branchedKey)
	}
	if linearKey == branchedKey {
		t.Fatalf("expected branch shape to affect route key, got %q", linearKey)
	}
}
