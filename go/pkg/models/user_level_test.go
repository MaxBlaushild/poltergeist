package models

import "testing"

func TestUserLevelXPToNextLevelScalesUpWithoutPlateau(t *testing.T) {
	level := &UserLevel{}

	got := map[int]int{}
	for _, currentLevel := range []int{1, 2, 5, 10, 15, 20} {
		level.Level = currentLevel
		got[currentLevel] = level.XPToNextLevel()
	}

	want := map[int]int{
		1:  100,
		2:  180,
		5:  660,
		10: 2260,
		15: 4860,
		20: 8460,
	}

	for currentLevel, expected := range want {
		if got[currentLevel] != expected {
			t.Fatalf("level %d expected %d XP to next level, got %d", currentLevel, expected, got[currentLevel])
		}
	}

	if got[15] <= 3000 {
		t.Fatalf("expected high-level XP requirement to exceed 3000, got %d", got[15])
	}
}

func TestUserLevelAddExperiencePointsSupportsMultipleLevelUps(t *testing.T) {
	userLevel := &UserLevel{
		Level:                   1,
		ExperiencePointsOnLevel: 0,
		TotalExperiencePoints:   0,
	}

	userLevel.AddExperiencePoints(500)

	if userLevel.Level != 3 {
		t.Fatalf("expected level 3 after a 500 XP gain, got %d", userLevel.Level)
	}
	if userLevel.LevelsGained != 2 {
		t.Fatalf("expected 2 levels gained, got %d", userLevel.LevelsGained)
	}
	if userLevel.ExperiencePointsOnLevel != 220 {
		t.Fatalf("expected 220 XP carried into the new level, got %d", userLevel.ExperiencePointsOnLevel)
	}
	if userLevel.ExperienceToNextLevel != 300 {
		t.Fatalf("expected next level threshold to be recomputed to 300, got %d", userLevel.ExperienceToNextLevel)
	}
	if userLevel.TotalExperiencePoints != 500 {
		t.Fatalf("expected total experience to track the full award, got %d", userLevel.TotalExperiencePoints)
	}
}
