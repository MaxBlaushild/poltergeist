package processors

import "testing"

func TestZoneSeedEncounterMemberCountRange(t *testing.T) {
	for i := 0; i < 500; i++ {
		count := zoneSeedEncounterMemberCount()
		if count < 1 || count > 3 {
			t.Fatalf("expected encounter member count between 1 and 3, got %d", count)
		}
	}
}
