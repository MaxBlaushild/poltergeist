package server

import "testing"

func TestScaledScenarioDifficultyForUserLevel(t *testing.T) {
	testCases := []struct {
		level    int
		expected int
	}{
		{level: 1, expected: 10},
		{level: 5, expected: 18},
		{level: 10, expected: 28},
		{level: 15, expected: 38},
		{level: 20, expected: 48},
		{level: 25, expected: 58},
		{level: 30, expected: 68},
		{level: 40, expected: 88},
		{level: 50, expected: 108},
		{level: 70, expected: 148},
		{level: 100, expected: 208},
	}

	for _, tc := range testCases {
		if got := scaledScenarioDifficultyForUserLevel(tc.level); got != tc.expected {
			t.Fatalf("expected scaled scenario difficulty for level %d to be %d, got %d", tc.level, tc.expected, got)
		}
	}
}

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
