package slice

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

// Same shape as sampleGCodeWithSupport but with a much larger support
// section relative to the rest of the print, for the "extensive support"
// side of the threshold tests.
const sampleGCodeWithExtensiveSupport = `
;TYPE:Perimeter
G1 X20 Y20 E1 F1500
;TYPE:Support material
G1 X25 Y25 E10 F1500
G1 X26 Y26 E19 F1500
;TYPE:Internal infill
G1 X30 Y30 E20 F1500

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
	if result.SupportMaterialPercent != 0 {
		t.Errorf("SupportMaterialPercent = %v, want 0 (no ;TYPE:Support material in sample)", result.SupportMaterialPercent)
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
	// 0.5 support out of 3.0 total extrusion (E1->E2 perimeter, E2->E2.5
	// support, E2.5->E3 infill) = 16.67%.
	wantPercent := 0.5 / 3.0 * 100
	if diff := result.SupportMaterialPercent - wantPercent; diff > 0.01 || diff < -0.01 {
		t.Errorf("SupportMaterialPercent = %.4f, want %.4f", result.SupportMaterialPercent, wantPercent)
	}
}

func TestParseGCodeStats_ExtensiveSupportIsMostOfExtrusion(t *testing.T) {
	result, err := ParseGCodeStats(sampleGCodeWithExtensiveSupport)
	if err != nil {
		t.Fatalf("ParseGCodeStats: %v", err)
	}
	// 18 support out of 20 total extrusion (E0->E1 perimeter, E1->E10->E19
	// support, E19->E20 infill) = 90%.
	wantPercent := 18.0 / 20.0 * 100
	if diff := result.SupportMaterialPercent - wantPercent; diff > 0.01 || diff < -0.01 {
		t.Errorf("SupportMaterialPercent = %.4f, want %.4f", result.SupportMaterialPercent, wantPercent)
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

// TestSlice_PassesFilamentDensityFlag verifies the actual bug fix at the Go
// level: a stub "slicer" records its invoked args and writes a minimal
// valid gcode footer, so this doesn't need a real PrusaSlicer binary to
// confirm --filament-density is present exactly when configured. The claim
// that omitting it silently zeroes out weight was confirmed separately
// against a real PrusaSlicer 2.7.4 binary (see slice.go's doc comment).
func TestSlice_PassesFilamentDensityFlag(t *testing.T) {
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args.txt")
	stub := writeStubSlicer(t, dir, argsFile)

	cfg := Config{
		SlicerBin:           stub,
		BaseTempDir:         dir,
		Timeout:             5 * time.Second,
		MemoryMB:            256,
		FilamentDensityGCm3: 1.27,
	}
	if _, err := Slice(context.Background(), cfg, filepath.Join(dir, "in.stl")); err != nil {
		t.Fatalf("Slice: %v", err)
	}

	got, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("reading recorded args: %v", err)
	}
	if !strings.Contains(string(got), "--filament-density 1.27") {
		t.Fatalf("stub slicer args = %q, want it to contain --filament-density 1.27", got)
	}
}

func TestSlice_OmitsFilamentDensityFlagWhenUnset(t *testing.T) {
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args.txt")
	stub := writeStubSlicer(t, dir, argsFile)

	cfg := Config{
		SlicerBin:   stub,
		BaseTempDir: dir,
		Timeout:     5 * time.Second,
		MemoryMB:    256,
	}
	if _, err := Slice(context.Background(), cfg, filepath.Join(dir, "in.stl")); err != nil {
		t.Fatalf("Slice: %v", err)
	}

	got, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("reading recorded args: %v", err)
	}
	if strings.Contains(string(got), "--filament-density") {
		t.Fatalf("stub slicer args = %q, want no --filament-density flag when unset", got)
	}
}

// writeStubSlicer creates an executable shell script that records its
// argv to argsFile and writes a minimal valid gcode footer to whatever
// path follows -o, so Slice's own ParseGCodeStats call succeeds.
func writeStubSlicer(t *testing.T, dir, argsFile string) string {
	t.Helper()
	stub := filepath.Join(dir, "stub-slicer.sh")
	script := `#!/bin/sh
echo "$@" >> "` + argsFile + `"
out=""
prev=""
for arg in "$@"; do
  if [ "$prev" = "-o" ]; then
    out="$arg"
  fi
  prev="$arg"
done
cat > "$out" <<'EOF'
; total filament used [g] = 10.00
; estimated printing time (normal mode) = 10m 0s
EOF
`
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatalf("writing stub slicer: %v", err)
	}
	return stub
}
