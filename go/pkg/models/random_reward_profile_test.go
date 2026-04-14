package models

import "testing"

func TestRandomRewardProfileForSizeEquippableRates(t *testing.T) {
	small := randomRewardProfileForSize(RandomRewardSizeSmall)
	if small.equippableChance != 0.1 {
		t.Fatalf("expected small equippable chance 0.1, got %v", small.equippableChance)
	}
	if small.equippableGuaranteed {
		t.Fatalf("expected small equippable rewards to remain chance-based")
	}

	medium := randomRewardProfileForSize(RandomRewardSizeMedium)
	if medium.equippableChance != 0.35 {
		t.Fatalf("expected medium equippable chance 0.35, got %v", medium.equippableChance)
	}
	if medium.equippableGuaranteed {
		t.Fatalf("expected medium equippable rewards to remain chance-based")
	}

	large := randomRewardProfileForSize(RandomRewardSizeLarge)
	if !large.equippableGuaranteed {
		t.Fatalf("expected large equippable rewards to remain guaranteed")
	}
}
