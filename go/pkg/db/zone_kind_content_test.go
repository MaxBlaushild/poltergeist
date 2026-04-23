package db

import (
	"strings"
	"testing"
)

func TestInventoryItemZoneKindEvidenceCTEQualifiesGroupedZoneKind(t *testing.T) {
	if !strings.Contains(inventoryItemZoneKindEvidenceCTE, "GROUP BY e.inventory_item_id, e.zone_kind") {
		t.Fatalf("inventory item zone kind evidence query must qualify grouped zone_kind to avoid ambiguous column errors")
	}
}

func TestZoneKindReferenceTablesIncludesInventoryItemSuggestionJobs(t *testing.T) {
	for _, tableName := range zoneKindReferenceTables {
		if tableName == "inventory_item_suggestion_jobs" {
			return
		}
	}
	t.Fatal("expected inventory_item_suggestion_jobs to participate in zone kind reference replacement")
}
