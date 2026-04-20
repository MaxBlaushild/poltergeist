package models

import "testing"

func TestNormalizeTutorialStageCollapsesRemovedBaseStages(t *testing.T) {
	if got := normalizeTutorialStage(TutorialStagePostScenarioDialogue); got != TutorialStagePostScenarioDialogue {
		t.Fatalf("expected %q to normalize to itself, got %q", TutorialStagePostScenarioDialogue, got)
	}
	if got := normalizeTutorialStage(TutorialStagePostBasePlacement); got != TutorialStagePostBaseDialogue {
		t.Fatalf("expected %q to normalize to %q, got %q", TutorialStagePostBasePlacement, TutorialStagePostBaseDialogue, got)
	}
	if got := normalizeTutorialStage(TutorialStageHearth); got != TutorialStagePostBaseDialogue {
		t.Fatalf("expected %q to normalize to %q, got %q", TutorialStageHearth, TutorialStagePostBaseDialogue, got)
	}
}
