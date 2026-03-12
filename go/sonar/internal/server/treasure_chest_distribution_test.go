package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestAllocateTreasureChestDistributionCounts(t *testing.T) {
	counts := allocateTreasureChestDistributionCounts(7, []int{10, 15, 20, 20, 20, 15})
	want := []int{1, 1, 2, 1, 1, 1}
	if len(counts) != len(want) {
		t.Fatalf("expected %d counts, got %d", len(want), len(counts))
	}
	total := 0
	for i := range want {
		if counts[i] != want[i] {
			t.Fatalf("expected count[%d]=%d, got %d", i, want[i], counts[i])
		}
		total += counts[i]
	}
	if total != 7 {
		t.Fatalf("expected total 7, got %d", total)
	}
}

func TestTreasureChestRewardSizeForUnlockTier(t *testing.T) {
	tier10 := 10
	tier25 := 25
	tier26 := 26
	tier50 := 50
	tier51 := 51
	tier100 := 100

	if got := treasureChestRewardSizeForUnlockTier(nil); got != models.RandomRewardSizeSmall {
		t.Fatalf("expected unlocked chest to be small, got %s", got)
	}
	if got := treasureChestRewardSizeForUnlockTier(&tier10); got != models.RandomRewardSizeSmall {
		t.Fatalf("expected 1-10 chest to be small, got %s", got)
	}
	if got := treasureChestRewardSizeForUnlockTier(&tier25); got != models.RandomRewardSizeSmall {
		t.Fatalf("expected 11-25 chest to be small, got %s", got)
	}
	if got := treasureChestRewardSizeForUnlockTier(&tier26); got != models.RandomRewardSizeMedium {
		t.Fatalf("expected 26-50 chest to be medium, got %s", got)
	}
	if got := treasureChestRewardSizeForUnlockTier(&tier50); got != models.RandomRewardSizeMedium {
		t.Fatalf("expected 26-50 chest to be medium, got %s", got)
	}
	if got := treasureChestRewardSizeForUnlockTier(&tier51); got != models.RandomRewardSizeLarge {
		t.Fatalf("expected 51-75 chest to be large, got %s", got)
	}
	if got := treasureChestRewardSizeForUnlockTier(&tier100); got != models.RandomRewardSizeLarge {
		t.Fatalf("expected 76-100 chest to be large, got %s", got)
	}
}
