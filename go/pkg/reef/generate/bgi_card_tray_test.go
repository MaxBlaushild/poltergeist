package generate

import (
	"strings"
	"testing"
)

func TestBgiCardTrayLayout_WellDepthScalesWithCardCountAndThickness(t *testing.T) {
	l, err := bgiCardTrayParamsToLayout(map[string]interface{}{
		"cardWidthMm":          44.0,
		"cardHeightMm":         68.0,
		"cardCount":            52.0,
		"totalCardThicknessMm": 0.47,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantWellDepth := 52*0.47 + bgiCardTrayScoopHeadroomMm
	if l.wellDepthMm != wantWellDepth {
		t.Fatalf("wellDepthMm = %.2f, want %.2f", l.wellDepthMm, wantWellDepth)
	}
	if l.outerDepthMm != wantWellDepth+bgiCardTrayFloorThicknessMm {
		t.Fatalf("outerDepthMm = %.2f, want %.2f", l.outerDepthMm, wantWellDepth+bgiCardTrayFloorThicknessMm)
	}
}

func TestBgiCardTray_Analyze_ReportsHeightMmForStacking(t *testing.T) {
	m := BgiCardTray{}
	a, err := m.Analyze(map[string]interface{}{
		"cardWidthMm":          44.0,
		"cardHeightMm":         68.0,
		"cardCount":            52.0,
		"totalCardThicknessMm": 0.47,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantHeight := 52*0.47 + bgiCardTrayScoopHeadroomMm + bgiCardTrayFloorThicknessMm
	if a.HeightMm != wantHeight {
		t.Fatalf("HeightMm = %.2f, want %.2f", a.HeightMm, wantHeight)
	}
	if a.HasInternalCavity != true {
		t.Fatal("expected HasInternalCavity=true — the card well is a cavity")
	}
	if a.SealedVoid {
		t.Fatal("the well is open-top; SealedVoid must be false")
	}
}

func TestBgiCardTray_Analyze_MinWallIsFixedRegardlessOfCardCount(t *testing.T) {
	m := BgiCardTray{}
	small, err := m.Analyze(map[string]interface{}{
		"cardWidthMm": 44.0, "cardHeightMm": 68.0, "cardCount": 4.0, "totalCardThicknessMm": 0.3,
	})
	if err != nil {
		t.Fatal(err)
	}
	large, err := m.Analyze(map[string]interface{}{
		"cardWidthMm": 44.0, "cardHeightMm": 68.0, "cardCount": 400.0, "totalCardThicknessMm": 0.3,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Unlike FragRack/ShelfRack, wall thickness here is a fixed constant
	// (floor/wall thickness), not a function of how many cards are packed —
	// cards stack along depth, not across a shared width axis.
	if small.MinWallMm != large.MinWallMm {
		t.Fatalf("expected MinWallMm to stay constant regardless of cardCount: small=%.2f, large=%.2f", small.MinWallMm, large.MinWallMm)
	}
}

func TestBgiCardTray_ValidateParams_RejectsOverEnvelopeHeight(t *testing.T) {
	m := BgiCardTray{}
	// 500 cards at 0.5mm each is 250mm of stack alone — comfortably over
	// the 210mm envelope even before headroom/floor are added.
	err := m.ValidateParams(map[string]interface{}{
		"cardWidthMm": 44.0, "cardHeightMm": 68.0, "cardCount": 500.0, "totalCardThicknessMm": 0.5,
	})
	if err == nil {
		t.Fatal("expected ValidateParams to reject an over-envelope tray height, got nil")
	}
	if !strings.Contains(err.Error(), "cardCount") {
		t.Fatalf("error %q should name cardCount as the thing to change", err.Error())
	}
}

func TestBgiCardTray_ValidateParams_AcceptsHealthyCardCount(t *testing.T) {
	m := BgiCardTray{}
	err := m.ValidateParams(map[string]interface{}{
		"cardWidthMm": 44.0, "cardHeightMm": 68.0, "cardCount": 52.0, "totalCardThicknessMm": 0.47,
	})
	if err != nil {
		t.Fatalf("expected a healthy card count to pass, got: %v", err)
	}
}

func TestBgiCardTray_SCAD_ProducesNonEmptyOutput(t *testing.T) {
	m := BgiCardTray{}
	scad, err := m.SCAD(map[string]interface{}{
		"cardWidthMm": 44.0, "cardHeightMm": 68.0, "cardCount": 52.0, "totalCardThicknessMm": 0.47,
	}, Preview)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(scad, "card_tray();") {
		t.Fatalf("expected SCAD output to instantiate card_tray(), got:\n%s", scad)
	}
	if !strings.Contains(scad, "finger_scoop") {
		t.Fatalf("expected SCAD output to include the finger scoop cutout, got:\n%s", scad)
	}
}
