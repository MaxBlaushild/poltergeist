package server

import "testing"

func TestNormalizeZoneSeedDraftRequestNormalizesZoneKind(t *testing.T) {
	settings, err := normalizeZoneSeedDraftRequest(zoneSeedDraftRequest{
		ZoneKind: " Ancient Forest ",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if settings == nil {
		t.Fatal("expected normalized settings")
	}
	if settings.ZoneKind != "ancient-forest" {
		t.Fatalf("expected normalized zone kind, got %q", settings.ZoneKind)
	}
}
