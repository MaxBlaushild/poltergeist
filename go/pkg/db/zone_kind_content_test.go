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
