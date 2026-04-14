package server

import "testing"

func TestMonsterBattleDefeatResourceFloor(t *testing.T) {
	tests := []struct {
		name        string
		maxResource int
		floorPct    int
		want        int
	}{
		{name: "zero max stays zero", maxResource: 0, floorPct: 30, want: 0},
		{name: "health floor rounds up", maxResource: 99, floorPct: 30, want: 30},
		{name: "mana floor rounds up", maxResource: 11, floorPct: 25, want: 3},
		{name: "minimum floor is one when resource exists", maxResource: 1, floorPct: 25, want: 1},
		{name: "floor cannot exceed max", maxResource: 2, floorPct: 90, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := monsterBattleDefeatResourceFloor(tt.maxResource, tt.floorPct); got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestMonsterBattleDefeatWoundedStatusTemplate(t *testing.T) {
	status := monsterBattleDefeatWoundedStatusTemplate()
	if status.Name != monsterBattleDefeatWoundedStatusName {
		t.Fatalf(
			"expected %s status name, got %q",
			monsterBattleDefeatWoundedStatusName,
			status.Name,
		)
	}
	if status.Positive {
		t.Fatal("expected Wounded to be detrimental")
	}
	if status.DurationSeconds != int(monsterBattleDefeatWoundedDuration.Seconds()) {
		t.Fatalf(
			"expected duration %d, got %d",
			int(monsterBattleDefeatWoundedDuration.Seconds()),
			status.DurationSeconds,
		)
	}
	if status.StrengthMod != -2 ||
		status.DexterityMod != -2 ||
		status.IntelligenceMod != -2 ||
		status.WisdomMod != -2 {
		t.Fatalf("unexpected wounded stat modifiers: %+v", status)
	}
}
