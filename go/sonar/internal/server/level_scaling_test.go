package server

import "testing"

func TestScaledEncounterMonsterLevelForUserLevel(t *testing.T) {
	if got := scaledEncounterMonsterLevelForUserLevel(25, 1); got != 23 {
		t.Fatalf("expected 1-monster scaled level to be 23, got %d", got)
	}
	if got := scaledEncounterMonsterLevelForUserLevel(25, 2); got != 13 {
		t.Fatalf("expected 2-monster scaled level to be 13, got %d", got)
	}
	if got := scaledEncounterMonsterLevelForUserLevel(25, 3); got != 9 {
		t.Fatalf("expected 3-monster scaled level to be 9, got %d", got)
	}
	if got := scaledEncounterMonsterLevelForUserLevel(1, 3); got != 1 {
		t.Fatalf("expected minimum scaled level to be 1, got %d", got)
	}
}
