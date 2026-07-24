package slice

import "testing"

// Realistic PrusaSlicer G-code footer format (comment block PrusaSlicer
// appends after slicing). This is what ParseGCodeStats is written against.
const sampleGCodeNoSupport = `
;TYPE:Skirt/Brim
G1 X10 Y10 E1 F1500
;TYPE:Perimeter
G1 X20 Y20 E2 F1500
;TYPE:Internal infill
G1 X30 Y30 E3 F1500

; filament used [mm] = 3255.55
; filament used [cm3] = 7.83
; filament used [g] = 9.80
; filament cost = 0.20
; total filament used [g] = 9.80
; total filament cost = 0.20
; estimated printing time (normal mode) = 1h 12m 34s
; estimated first layer printing time (normal mode) = 1m 3s
`

const sampleGCodeWithSupport = `
;TYPE:Perimeter
G1 X20 Y20 E2 F1500
;TYPE:Support material
G1 X25 Y25 E2.5 F1500
;TYPE:Internal infill
G1 X30 Y30 E3 F1500

; total filament used [g] = 42.10
; estimated printing time (normal mode) = 3h 45m 10s
`

const sampleGCodeShortDuration = `
; total filament used [g] = 1.50
; estimated printing time (normal mode) = 45m 2s
`

func TestParseGCodeStats_NoSupport(t *testing.T) {
	result, err := ParseGCodeStats(sampleGCodeNoSupport)
	if err != nil {
		t.Fatalf("ParseGCodeStats: %v", err)
	}
	if result.WeightG != 9.80 {
		t.Errorf("WeightG = %v, want 9.80", result.WeightG)
	}
	wantSeconds := int64(1*60*60 + 12*60 + 34)
	if result.PrintTimeS != wantSeconds {
		t.Errorf("PrintTimeS = %d, want %d", result.PrintTimeS, wantSeconds)
	}
	if result.SupportRequired {
		t.Error("SupportRequired = true, want false (no ;TYPE:Support material in sample)")
	}
}

func TestParseGCodeStats_WithSupport(t *testing.T) {
	result, err := ParseGCodeStats(sampleGCodeWithSupport)
	if err != nil {
		t.Fatalf("ParseGCodeStats: %v", err)
	}
	if result.WeightG != 42.10 {
		t.Errorf("WeightG = %v, want 42.10", result.WeightG)
	}
	wantSeconds := int64(3*60*60 + 45*60 + 10)
	if result.PrintTimeS != wantSeconds {
		t.Errorf("PrintTimeS = %d, want %d", result.PrintTimeS, wantSeconds)
	}
	if !result.SupportRequired {
		t.Error("SupportRequired = false, want true (sample has ;TYPE:Support material extrusions)")
	}
}

func TestParseGCodeStats_MinutesSecondsOnlyDuration(t *testing.T) {
	result, err := ParseGCodeStats(sampleGCodeShortDuration)
	if err != nil {
		t.Fatalf("ParseGCodeStats: %v", err)
	}
	wantSeconds := int64(45*60 + 2)
	if result.PrintTimeS != wantSeconds {
		t.Errorf("PrintTimeS = %d, want %d", result.PrintTimeS, wantSeconds)
	}
}

func TestParseGCodeStats_MissingWeightErrors(t *testing.T) {
	if _, err := ParseGCodeStats("; estimated printing time (normal mode) = 1h 0m 0s"); err == nil {
		t.Fatal("expected an error when filament weight is missing from the gcode")
	}
}

func TestParseGCodeStats_MissingTimeErrors(t *testing.T) {
	if _, err := ParseGCodeStats("; total filament used [g] = 5.00"); err == nil {
		t.Fatal("expected an error when print time is missing from the gcode")
	}
}

func TestParseSlicerDuration(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"1h 12m 34s", 1*3600 + 12*60 + 34},
		{"45m 2s", 45*60 + 2},
		{"5s", 5},
		{"1d 2h 3m 4s", 24*3600 + 2*3600 + 3*60 + 4},
	}
	for _, c := range cases {
		got, err := parseSlicerDuration(c.in)
		if err != nil {
			t.Errorf("parseSlicerDuration(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseSlicerDuration(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
