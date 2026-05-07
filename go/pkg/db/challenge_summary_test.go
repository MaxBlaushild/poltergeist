package db

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestSummarizeChallengeAdminRows(t *testing.T) {
	summary := summarizeChallengeAdminRows([]challengeAdminSummaryRow{
		{
			EffectiveZoneKind:  "academy",
			HasPointOfInterest: true,
			HasPolygon:         false,
			IsRecurring:        true,
			SubmissionType:     models.QuestNodeSubmissionTypePhoto,
			Difficulty:         2,
			StatTags:           models.StringArray{"intelligence", "wisdom", "wisdom"},
		},
		{
			EffectiveZoneKind:  "",
			HasPointOfInterest: false,
			HasPolygon:         true,
			IsRecurring:        false,
			SubmissionType:     models.QuestNodeSubmissionTypeText,
			Difficulty:         6,
			StatTags:           models.StringArray{"charisma", ""},
		},
		{
			EffectiveZoneKind:  "academy",
			HasPointOfInterest: false,
			HasPolygon:         false,
			IsRecurring:        false,
			SubmissionType:     "",
			Difficulty:         10,
			StatTags:           nil,
		},
	})

	if summary.TotalChallenges != 3 {
		t.Fatalf("expected 3 total challenges, got %d", summary.TotalChallenges)
	}
	if summary.PointOfInterestCount != 1 {
		t.Fatalf("expected 1 point-of-interest challenge, got %d", summary.PointOfInterestCount)
	}
	if summary.PolygonCount != 1 {
		t.Fatalf("expected 1 polygon challenge, got %d", summary.PolygonCount)
	}
	if summary.RecurringCount != 1 {
		t.Fatalf("expected 1 recurring challenge, got %d", summary.RecurringCount)
	}

	assertBucketCount(t, summary.ZoneKindCounts, "academy", 2)
	assertBucketCount(t, summary.ZoneKindCounts, "", 1)
	assertBucketCount(t, summary.SubmissionTypeCounts, "Photo", 1)
	assertBucketCount(t, summary.SubmissionTypeCounts, "Text", 1)
	assertBucketCount(t, summary.SubmissionTypeCounts, "Video", 1)
	assertBucketCount(t, summary.DifficultyBandCounts, "0-2", 1)
	assertBucketCount(t, summary.DifficultyBandCounts, "6-8", 1)
	assertBucketCount(t, summary.DifficultyBandCounts, "9+", 1)
	assertBucketCount(t, summary.PlacementCounts, "Point of interest", 1)
	assertBucketCount(t, summary.PlacementCounts, "Polygon area", 1)
	assertBucketCount(t, summary.PlacementCounts, "Coordinates", 1)
	assertBucketCount(t, summary.StatTagCounts, "wisdom", 1)
	assertBucketCount(t, summary.StatTagCounts, "intelligence", 1)
	assertBucketCount(t, summary.StatTagCounts, "charisma", 1)
}

func assertBucketCount(
	t *testing.T,
	buckets []ChallengeAdminDashboardBucket,
	key string,
	want int,
) {
	t.Helper()
	for _, bucket := range buckets {
		if bucket.Key == key {
			if bucket.Count != want {
				t.Fatalf("bucket %q = %d, want %d", key, bucket.Count, want)
			}
			return
		}
	}
	t.Fatalf("missing bucket %q", key)
}
