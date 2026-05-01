package processors

import (
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
