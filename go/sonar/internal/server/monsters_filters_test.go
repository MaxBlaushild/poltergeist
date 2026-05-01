package server

import "testing"

func TestParseOptionalZoneKindFilter(t *testing.T) {
	t.Run("treats empty values as unset", func(t *testing.T) {
		if got := parseOptionalZoneKindFilter(""); got != "" {
			t.Fatalf("expected empty zone kind filter, got %q", got)
		}
		if got := parseOptionalZoneKindFilter(" all "); got != "" {
			t.Fatalf("expected all zone kind filter to be unset, got %q", got)
		}
	})

	t.Run("normalizes explicit zone kind filters", func(t *testing.T) {
		got := parseOptionalZoneKindFilter(" Ancient Forest ")
		if got != "ancient-forest" {
			t.Fatalf("expected normalized zone kind filter, got %q", got)
		}
	})
}
