package generate

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

func requireOpenSCAD(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("openscad")
	if err != nil {
		t.Skip("openscad not installed, skipping render test")
	}
	return bin
}

func renderCfg(t *testing.T) RenderConfig {
	return RenderConfig{
		OpenSCADBin: requireOpenSCAD(t),
		BaseTempDir: t.TempDir(),
		Timeout:     30 * time.Second,
		MemoryMB:    1024,
	}
}

func mustVersion(t *testing.T, cfg RenderConfig) string {
	t.Helper()
	v, err := Version(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	return v
}

func TestFragRack_RendersValidSTL(t *testing.T) {
	cfg := renderCfg(t)
	m, err := Get("frag_rack")
	if err != nil {
		t.Fatal(err)
	}
	params := map[string]interface{}{
		"tankProfileId":      nil,
		"glassThicknessMm":   10.0,
		"tierCount":          2.0,
		"widthMm":            150.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerTier":       6.0,
		"color":              "black",
	}
	scad, err := m.SCAD(params, Preview)
	if err != nil {
		t.Fatalf("SCAD: %v", err)
	}

	result, err := Render(context.Background(), cfg, scad, mustVersion(t, cfg))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	defer os.RemoveAll(result.WorkDir)

	info, err := os.Stat(result.STLPath)
	if err != nil {
		t.Fatalf("output STL missing: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output STL is empty")
	}
	if result.OpenSCADVersion == "" || result.OpenSCADVersion == "unknown" {
		t.Fatalf("expected a resolved OpenSCAD version, got %q", result.OpenSCADVersion)
	}
	assertBinarySTL(t, result.STLPath)
}

func TestFragRack_FullDetailAlsoRenders(t *testing.T) {
	cfg := renderCfg(t)
	cfg.Timeout = 90 * time.Second // full-res CSG with many holes is the slowest case
	m, _ := Get("frag_rack")
	params := map[string]interface{}{
		"glassThicknessMm":   8.0,
		"tierCount":          3.0,
		"widthMm":            180.0,
		"plugHoleDiameterMm": 15.0,
		"holesPerTier":       8.0,
		"color":              "white",
	}
	scad, err := m.SCAD(params, Full)
	if err != nil {
		t.Fatalf("SCAD: %v", err)
	}
	result, err := Render(context.Background(), cfg, scad, mustVersion(t, cfg))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	defer os.RemoveAll(result.WorkDir)
	assertBinarySTL(t, result.STLPath)
}

func TestLidClip_RendersValidSTL(t *testing.T) {
	cfg := renderCfg(t)
	m, err := Get("lid_clip")
	if err != nil {
		t.Fatal(err)
	}
	params := map[string]interface{}{
		"rimThicknessMm":  8.0,
		"rimWidthMm":      22.0,
		"euroBrace":       false,
		"meshThicknessMm": 1.2,
		"quantity":        4.0,
	}
	scad, err := m.SCAD(params, Preview)
	if err != nil {
		t.Fatalf("SCAD: %v", err)
	}
	result, err := Render(context.Background(), cfg, scad, mustVersion(t, cfg))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	defer os.RemoveAll(result.WorkDir)
	assertBinarySTL(t, result.STLPath)
}

func TestLidClip_EuroBraceVariantRenders(t *testing.T) {
	cfg := renderCfg(t)
	m, _ := Get("lid_clip")
	params := map[string]interface{}{
		"rimThicknessMm":  10.0,
		"rimWidthMm":      25.0,
		"euroBrace":       true,
		"meshThicknessMm": 1.5,
		"quantity":        8.0,
	}
	scad, err := m.SCAD(params, Preview)
	if err != nil {
		t.Fatalf("SCAD: %v", err)
	}
	result, err := Render(context.Background(), cfg, scad, mustVersion(t, cfg))
	if err != nil {
		t.Fatalf("Render (euro brace): %v", err)
	}
	defer os.RemoveAll(result.WorkDir)
	assertBinarySTL(t, result.STLPath)
}

func TestVersion_ResolvesRealBinary(t *testing.T) {
	cfg := renderCfg(t)
	v, err := Version(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if v == "" {
		t.Fatal("expected a non-empty version string")
	}
	t.Logf("resolved OpenSCAD version: %s", v)
}

// assertBinarySTL does a minimal structural sanity check: binary STL files
// start with an 80-byte header followed by a uint32 triangle count, and the
// file size must match 84 + count*50 exactly.
func assertBinarySTL(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading STL: %v", err)
	}
	if len(data) < 84 {
		t.Fatalf("STL too small to be binary (%d bytes)", len(data))
	}
	count := uint32(data[80]) | uint32(data[81])<<8 | uint32(data[82])<<16 | uint32(data[83])<<24
	wantSize := 84 + int(count)*50
	if len(data) != wantSize {
		t.Fatalf("STL size %d doesn't match binary STL layout for %d triangles (want %d) — is this ASCII STL instead of binary?", len(data), count, wantSize)
	}
	if count == 0 {
		t.Fatal("STL has zero triangles — generator produced empty geometry")
	}
}
