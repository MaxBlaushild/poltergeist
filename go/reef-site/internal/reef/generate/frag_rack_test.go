package generate

import "testing"

// Regression test: an earlier version used a fixed 8mm edge margin for both
// plug holes and magnets. With a 20mm plug hole (R-4.2's larger standard
// size) that put the margin (8mm) *inside* the hole's own radius (10mm),
// meaning the hole would punch through the part's side edge. The margin now
// scales with the hole it protects.
func TestFragRackLayout_EdgeMarginScalesWithHoleDiameter(t *testing.T) {
	l, err := fragRackParamsToLayout(map[string]interface{}{
		"glassThicknessMm":   10.0,
		"tierCount":          2.0,
		"widthMm":            150.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerTier":       6.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if l.plugEdgeMarginMm <= l.plugHoleDiameterMm/2 {
		t.Fatalf("plugEdgeMarginMm = %.2f must exceed the hole radius %.2f, or the hole punches through the edge",
			l.plugEdgeMarginMm, l.plugHoleDiameterMm/2)
	}
}

func TestFragRack_Analyze_FlagsThinWallsFromPackedHoles(t *testing.T) {
	m := FragRack{}
	// Deliberately hostile: max holes_per_tier on the minimum width_mm with
	// the larger plug hole — should produce an unprintably thin (or
	// negative) wall between adjacent holes.
	a, err := m.Analyze(map[string]interface{}{
		"glassThicknessMm":   10.0,
		"tierCount":          1.0,
		"widthMm":            60.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerTier":       12.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.MinWallMm >= 2.0 {
		t.Fatalf("expected Analyze to catch the packed-hole thin wall, got MinWallMm = %.2f", a.MinWallMm)
	}
}

func TestFragRack_Analyze_HealthyParamsHaveSafeWalls(t *testing.T) {
	m := FragRack{}
	a, err := m.Analyze(map[string]interface{}{
		"glassThicknessMm":   10.0,
		"tierCount":          2.0,
		"widthMm":            150.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerTier":       5.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.MinWallMm < 2.0 {
		t.Fatalf("expected a comfortable wall for modest params, got MinWallMm = %.2f", a.MinWallMm)
	}
	if a.SealedVoid {
		t.Fatal("frag rack magnet pockets are always vented; SealedVoid must be false")
	}
	if a.DrainPathMm < 4.0 {
		t.Fatalf("DrainPathMm = %.2f, want >= 4 (R-5.3 minimum)", a.DrainPathMm)
	}
}

func TestFragRackMaxHolesPerTier_ScalesWithWidthAndHoleSize(t *testing.T) {
	small20mm := FragRackMaxHolesPerTier(60, 20)
	large20mm := FragRackMaxHolesPerTier(250, 20)
	if large20mm <= small20mm {
		t.Fatalf("expected more max holes on a wider rack: 60mm=%d, 250mm=%d", small20mm, large20mm)
	}
	if small20mm < 1 {
		t.Fatalf("MaxHolesPerTier must never be less than 1, got %d", small20mm)
	}

	// A smaller hole should allow at least as many holes as a larger one at
	// the same width.
	holes15 := FragRackMaxHolesPerTier(150, 15)
	holes20 := FragRackMaxHolesPerTier(150, 20)
	if holes15 < holes20 {
		t.Fatalf("expected 15mm holes to allow >= as many holes as 20mm at the same width: 15mm=%d, 20mm=%d", holes15, holes20)
	}
}
