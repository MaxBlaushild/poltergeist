package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
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

func TestRepairQuestArchetypeSuggestionDraftLocationArchetypesUsesConceptAndRequiredHints(t *testing.T) {
	bookNookID := uuid.New()
	observatoryID := uuid.New()
	draft := &models.QuestArchetypeSuggestionDraft{
		Nodes: models.QuestArchetypeSuggestionNodes{
			{
				NodeKey:                 "cozy_book_nook_entry",
				Source:                  "location",
				Content:                 "challenge",
				LocationConcept:         "Cozy Book Nook Entry",
				LocationMetadataTags:    []string{"scholarly", "mystical", "secretive"},
				TemplateConcept:         "identify the hidden sigil",
				ChallengeQuestion:       "Photograph what looks like a hidden sigil within a book of your choice.",
				ChallengeDescription:    "Look for arcane-looking shapes, symbols, or decorative marks that could pass for a hidden sigil.",
				ChallengeSubmissionType: models.QuestNodeSubmissionTypePhoto,
			},
		},
	}

	changed := repairQuestArchetypeSuggestionDraftLocationArchetypes(
		draft,
		[]*models.LocationArchetype{
			{ID: observatoryID, Name: "Celestial Observatory"},
			{ID: bookNookID, Name: "Cozy Book Nook"},
		},
		[]string{observatoryID.String(), bookNookID.String()},
	)
	if !changed {
		t.Fatalf("expected location archetype repair to apply")
	}
	if len(draft.Nodes) != 1 || draft.Nodes[0].LocationArchetypeID == nil {
		t.Fatalf("expected repaired node location archetype id, got %+v", draft.Nodes)
	}
	if *draft.Nodes[0].LocationArchetypeID != bookNookID {
		t.Fatalf("expected cozy book nook archetype to be chosen, got %+v", draft.Nodes[0].LocationArchetypeID)
	}
	if draft.Nodes[0].LocationArchetypeName != "Cozy Book Nook" {
		t.Fatalf("expected repaired node location archetype name, got %q", draft.Nodes[0].LocationArchetypeName)
	}
	if len(draft.Steps) != 1 || draft.Steps[0].LocationArchetypeID == nil {
		t.Fatalf("expected repaired steps to stay in sync, got %+v", draft.Steps)
	}
	if *draft.Steps[0].LocationArchetypeID != bookNookID {
		t.Fatalf("expected repaired step location archetype id, got %+v", draft.Steps[0].LocationArchetypeID)
	}
}

func TestRepairQuestArchetypeSuggestionDraftLocationArchetypesFallsBackToSingleRequiredArchetype(t *testing.T) {
	archiveID := uuid.New()
	draft := &models.QuestArchetypeSuggestionDraft{
		Steps: models.QuestArchetypeSuggestionSteps{
			{
				Source:               "location",
				Content:              "scenario",
				LocationConcept:      "Quiet Entry Hall",
				LocationMetadataTags: []string{"dusty", "scholarly"},
				TemplateConcept:      "read the warning note",
				ScenarioPrompt:       "A warning waits just inside the threshold. What do you do?",
			},
		},
	}

	changed := repairQuestArchetypeSuggestionDraftLocationArchetypes(
		draft,
		[]*models.LocationArchetype{
			{ID: archiveID, Name: "Forgotten Archive"},
		},
		[]string{archiveID.String()},
	)
	if !changed {
		t.Fatalf("expected single required archetype fallback to apply")
	}
	if len(draft.Steps) != 1 || draft.Steps[0].LocationArchetypeID == nil {
		t.Fatalf("expected repaired step location archetype id, got %+v", draft.Steps)
	}
	if *draft.Steps[0].LocationArchetypeID != archiveID {
		t.Fatalf("expected single required archetype to be used, got %+v", draft.Steps[0].LocationArchetypeID)
	}
	if draft.Steps[0].LocationArchetypeName != "Forgotten Archive" {
		t.Fatalf("expected single required archetype name to be filled, got %q", draft.Steps[0].LocationArchetypeName)
	}
}
