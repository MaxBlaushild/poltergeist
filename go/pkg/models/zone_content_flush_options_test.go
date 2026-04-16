package models

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseZoneContentFlushTypes(t *testing.T) {
	t.Run("deduplicates while preserving order", func(t *testing.T) {
		contentTypes, err := ParseZoneContentFlushTypes([]string{
			"pointsOfInterest",
			"quests",
			"pointsOfInterest",
			"jobs",
		})
		if err != nil {
			t.Fatalf("ParseZoneContentFlushTypes returned error: %v", err)
		}

		expected := []ZoneContentFlushType{
			ZoneContentFlushTypePointsOfInterest,
			ZoneContentFlushTypeQuests,
			ZoneContentFlushTypeJobs,
		}
		if diff := cmp.Diff(expected, contentTypes); diff != "" {
			t.Fatalf("unexpected parsed content types (-want +got):\n%s", diff)
		}
	})

	t.Run("accepts case-insensitive values", func(t *testing.T) {
		contentTypes, err := ParseZoneContentFlushTypes([]string{
			"SCENARIOS",
			"healingFountains",
		})
		if err != nil {
			t.Fatalf("ParseZoneContentFlushTypes returned error: %v", err)
		}

		expected := []ZoneContentFlushType{
			ZoneContentFlushTypeScenarios,
			ZoneContentFlushTypeHealingFountains,
		}
		if diff := cmp.Diff(expected, contentTypes); diff != "" {
			t.Fatalf("unexpected parsed content types (-want +got):\n%s", diff)
		}
	})

	t.Run("rejects invalid values", func(t *testing.T) {
		if _, err := ParseZoneContentFlushTypes([]string{"unknown"}); err == nil {
			t.Fatal("expected invalid content type to return an error")
		}
	})

	t.Run("rejects empty values", func(t *testing.T) {
		if _, err := ParseZoneContentFlushTypes([]string{"  "}); err == nil {
			t.Fatal("expected empty content type to return an error")
		}
	})
}
