package models

import "testing"

func TestNormalizeZoneKind(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "blank", input: "   ", want: ""},
		{name: "lowercases", input: "Forest", want: "forest"},
		{name: "spaces become dashes", input: "Ancient Forest", want: "ancient-forest"},
		{name: "underscores collapse", input: "coastal__city", want: "coastal-city"},
		{name: "mixed separators collapse", input: "  ruined - waterfront  ", want: "ruined-waterfront"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NormalizeZoneKind(test.input); got != test.want {
				t.Fatalf("NormalizeZoneKind(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestDistrictZoneSeedSettingsHasContentWithZoneKind(t *testing.T) {
	settings := DistrictZoneSeedSettings{ZoneKind: "forest"}
	if !settings.HasContent() {
		t.Fatal("expected zone kind alone to count as district seed content")
	}
}

func TestZoneKindApplyToCounts(t *testing.T) {
	zoneKind := ZoneKind{
		PlaceCountRatio:             1.5,
		MonsterCountRatio:           0.5,
		BossEncounterCountRatio:     2.0,
		RaidEncounterCountRatio:     0,
		InputEncounterCountRatio:    1.25,
		OptionEncounterCountRatio:   1.0,
		TreasureChestCountRatio:     0.5,
		HealingFountainCountRatio:   2.0,
		HerbalismResourceCountRatio: 2.0,
		MiningResourceCountRatio:    1.5,
	}

	got := zoneKind.ApplyToCounts(ZoneSeedResolvedCounts{
		PlaceCount:             4,
		MonsterCount:           6,
		BossEncounterCount:     2,
		RaidEncounterCount:     3,
		InputEncounterCount:    4,
		OptionEncounterCount:   5,
		TreasureChestCount:     2,
		HealingFountainCount:   1,
		HerbalismResourceCount: 2,
		MiningResourceCount:    2,
	})

	want := ZoneSeedResolvedCounts{
		PlaceCount:             6,
		MonsterCount:           3,
		BossEncounterCount:     4,
		RaidEncounterCount:     0,
		InputEncounterCount:    5,
		OptionEncounterCount:   5,
		TreasureChestCount:     1,
		HealingFountainCount:   2,
		HerbalismResourceCount: 4,
		MiningResourceCount:    3,
		ResourceCount:          7,
	}

	if got != want {
		t.Fatalf("ApplyToCounts() = %+v, want %+v", got, want)
	}
}

func TestNormalizeHexColor(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "blank", input: "   ", want: ""},
		{name: "six digit", input: "#A1B2C3", want: "#a1b2c3"},
		{name: "without hash", input: "445566", want: "#445566"},
		{name: "three digit", input: "#abc", want: "#aabbcc"},
		{name: "invalid length", input: "#abcd", want: ""},
		{name: "invalid character", input: "#zzzzzz", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NormalizeHexColor(test.input); got != test.want {
				t.Fatalf("NormalizeHexColor(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
