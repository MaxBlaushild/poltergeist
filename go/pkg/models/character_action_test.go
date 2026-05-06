package models

import "testing"

func TestDialogueSequenceFromSpeakerNameLinesAssignsSpeakerName(t *testing.T) {
	dialogue := DialogueSequenceFromSpeakerIdentityLines(
		[]string{
			"Do not trust the first locked door.",
			"The ward broke from the inside.",
		},
		"Witness Echo",
		"https://example.com/witness-echo.png",
	)

	if len(dialogue) != 2 {
		t.Fatalf("expected 2 dialogue lines, got %d", len(dialogue))
	}
	if dialogue[0].Speaker != "character" {
		t.Fatalf("expected character speaker, got %q", dialogue[0].Speaker)
	}
	if dialogue[0].SpeakerName != "Witness Echo" {
		t.Fatalf("expected speaker name to carry through, got %+v", dialogue[0])
	}
	if dialogue[0].PortraitURL != "https://example.com/witness-echo.png" {
		t.Fatalf("expected portrait url to carry through, got %+v", dialogue[0])
	}
	if dialogue[0].CharacterID != nil {
		t.Fatalf("expected no character id for reusable speaker label, got %+v", dialogue[0].CharacterID)
	}
}
