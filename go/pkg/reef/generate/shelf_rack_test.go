package generate

import (
	"strings"
	"testing"
)

func TestShelfRackLayout_EdgeMarginScalesWithHoleDiameter(t *testing.T) {
	l, err := shelfRackParamsToLayout(map[string]interface{}{
		"widthMm":            150.0,
		"depthMm":            80.0,
		"legHeightMm":        30.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerRow":        6.0,
		"rowCount":           2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if l.colEdgeMarginMm <= l.plugHoleDiameterMm/2 {
		t.Fatalf("colEdgeMarginMm = %.2f must exceed the hole radius %.2f, or the hole punches through the edge",
			l.colEdgeMarginMm, l.plugHoleDiameterMm/2)
	}
	if l.rowEdgeMarginMm <= l.plugHoleDiameterMm/2 {
		t.Fatalf("rowEdgeMarginMm = %.2f must exceed the hole radius %.2f, or the hole punches through the edge",
			l.rowEdgeMarginMm, l.plugHoleDiameterMm/2)
	}
}

func TestShelfRack_Analyze_FlagsThinWallsFromPackedHoles(t *testing.T) {
	m := ShelfRack{}
	// Deliberately hostile: max holesPerRow on the minimum widthMm with the
	// larger plug hole — should produce an unprintably thin (or negative)
	// wall between adjacent holes.
	a, err := m.Analyze(map[string]interface{}{
		"widthMm":            60.0,
		"depthMm":            40.0,
		"legHeightMm":        30.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerRow":        12.0,
		"rowCount":           1.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.MinWallMm >= 2.0 {
		t.Fatalf("expected Analyze to catch the packed-hole thin wall, got MinWallMm = %.2f", a.MinWallMm)
	}
}

func TestShelfRack_Analyze_HealthyParamsHaveSafeWalls(t *testing.T) {
	m := ShelfRack{}
	a, err := m.Analyze(map[string]interface{}{
		"widthMm":            150.0,
		"depthMm":            80.0,
		"legHeightMm":        30.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerRow":        5.0,
		"rowCount":           2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.MinWallMm < 2.0 {
		t.Fatalf("expected a comfortable wall for modest params, got MinWallMm = %.2f", a.MinWallMm)
	}
	if a.HasInternalCavity {
		t.Fatal("shelf rack holes go straight through and legs are solid; HasInternalCavity must be false")
	}
	if a.SealedVoid {
		t.Fatal("shelf rack has no cavity at all; SealedVoid must be false")
	}
}

func TestShelfRackMaxHolesPerRow_ScalesWithWidthAndHoleSize(t *testing.T) {
	small20mm := ShelfRackMaxHolesPerRow(60, 20)
	large20mm := ShelfRackMaxHolesPerRow(250, 20)
	if large20mm <= small20mm {
		t.Fatalf("expected more max holes on a wider deck: 60mm=%d, 250mm=%d", small20mm, large20mm)
	}
	if small20mm < 1 {
		t.Fatalf("MaxHolesPerRow must never be less than 1, got %d", small20mm)
	}
}

func TestShelfRackMaxRows_ScalesWithDepthAndHoleSize(t *testing.T) {
	small20mm := ShelfRackMaxRows(40, 20)
	large20mm := ShelfRackMaxRows(150, 20)
	if large20mm <= small20mm {
		t.Fatalf("expected more max rows on a deeper deck: 40mm=%d, 150mm=%d", small20mm, large20mm)
	}
	if small20mm < 1 {
		t.Fatalf("MaxRows must never be less than 1, got %d", small20mm)
	}
}

func TestShelfRack_ValidateParams_RejectsHolesPerRowExceedingDerivedMax(t *testing.T) {
	m := ShelfRack{}
	max := ShelfRackMaxHolesPerRow(152, 20)
	err := m.ValidateParams(map[string]interface{}{
		"widthMm":            152.0,
		"depthMm":            80.0,
		"legHeightMm":        30.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerRow":        float64(max + 1),
		"rowCount":           1.0,
	})
	if err == nil {
		t.Fatal("expected ValidateParams to reject holesPerRow one past the derived max, got nil")
	}
	if !strings.Contains(err.Error(), "holesPerRow") {
		t.Fatalf("error %q should name holesPerRow as the thing to change", err.Error())
	}
}

func TestShelfRack_ValidateParams_RejectsRowCountExceedingDerivedMax(t *testing.T) {
	m := ShelfRack{}
	max := ShelfRackMaxRows(80, 20)
	err := m.ValidateParams(map[string]interface{}{
		"widthMm":            150.0,
		"depthMm":            80.0,
		"legHeightMm":        30.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerRow":        4.0,
		"rowCount":           float64(max + 1),
	})
	if err == nil {
		t.Fatal("expected ValidateParams to reject rowCount one past the derived max, got nil")
	}
	if !strings.Contains(err.Error(), "rowCount") {
		t.Fatalf("error %q should name rowCount as the thing to change", err.Error())
	}
}

func TestShelfRack_ValidateParams_AcceptsHolesAtDerivedMax(t *testing.T) {
	m := ShelfRack{}
	maxHoles := ShelfRackMaxHolesPerRow(152, 20)
	maxRows := ShelfRackMaxRows(80, 20)
	err := m.ValidateParams(map[string]interface{}{
		"widthMm":            152.0,
		"depthMm":            80.0,
		"legHeightMm":        30.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerRow":        float64(maxHoles),
		"rowCount":           float64(maxRows),
	})
	if err != nil {
		t.Fatalf("expected holesPerRow/rowCount at exactly the derived max to pass, got: %v", err)
	}
}
