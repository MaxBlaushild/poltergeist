package server

import (
	"reflect"
	"testing"
)

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

func TestNormalizeZoneKindPayloadNormalizesDefaultShopkeeperItemTags(t *testing.T) {
	zoneKind, err := normalizeZoneKindPayload(zoneKindPayload{
		Name:                      "Forest",
		DefaultShopkeeperItemTags: []string{" Potions ", "herbs", "POTIONS"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := []string(zoneKind.DefaultShopkeeperItemTags), []string{"potions", "herbs"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected normalized default shopkeeper tags %v, got %v", want, got)
	}
}
