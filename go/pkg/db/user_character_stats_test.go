package db

import "testing"

func TestApplyResourceDeficitDelta(t *testing.T) {
	t.Run("clamps at zero when healing past full", func(t *testing.T) {
		got := applyResourceDeficitDelta(5, -10)
		if got != 0 {
			t.Fatalf("applyResourceDeficitDelta(5, -10) = %d, want 0", got)
		}
	})

	t.Run("allows deficits above raw derived max", func(t *testing.T) {
		got := applyResourceDeficitDelta(100, 19)
		if got != 119 {
			t.Fatalf("applyResourceDeficitDelta(100, 19) = %d, want 119", got)
		}
	})
}
