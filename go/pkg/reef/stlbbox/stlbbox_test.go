package stlbbox

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/reef/generate"
)

func TestFromFile_SyntheticCube(t *testing.T) {
	// A hand-built single-triangle-degenerate-free binary STL isn't worth
	// constructing by hand here; instead render a known-dimension part with
	// the real generator (already validated end-to-end in the generate
	// package tests) and check the bbox against the parameters that produced
	// it.
	bin, err := exec.LookPath("openscad")
	if err != nil {
		t.Skip("openscad not installed, skipping")
	}

	m, err := generate.Get("frag_rack")
	if err != nil {
		t.Fatal(err)
	}
	params := map[string]interface{}{
		"glassThicknessMm":   10.0,
		"tierCount":          1.0,
		"widthMm":            100.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerTier":       4.0,
		"color":              "black",
	}
	scad, err := m.SCAD(params, generate.Preview)
	if err != nil {
		t.Fatal(err)
	}
	renderCfg := generate.RenderConfig{
		OpenSCADBin: bin,
		BaseTempDir: t.TempDir(),
		Timeout:     30 * time.Second,
		MemoryMB:    1024,
	}
	version, err := generate.Version(context.Background(), renderCfg)
	if err != nil {
		t.Fatal(err)
	}
	result, err := generate.Render(context.Background(), renderCfg, scad, version)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(result.WorkDir)

	box, err := FromFile(result.STLPath)
	if err != nil {
		t.Fatalf("FromFile: %v", err)
	}

	// The X extent (width) is the part's widest axis and should land close
	// to widthMm=100 (the inner rack spans exactly width_mm on X).
	if box.XMm() < 95 || box.XMm() > 105 {
		t.Fatalf("X extent = %.2f, want ~100 (widthMm)", box.XMm())
	}
	// Two parts (inner rack + outer plate) laid out on Y with a gap between
	// them, so Y extent should be noticeably larger than either part alone
	// but still comfortably inside the print envelope for these params.
	if box.YMm() <= 0 || box.YMm() > 210 {
		t.Fatalf("Y extent = %.2f, want > 0 and <= 210", box.YMm())
	}
	if box.MaxDimensionMm() > 210 {
		t.Fatalf("MaxDimensionMm = %.2f, want <= 210 for these modest params", box.MaxDimensionMm())
	}
}

func TestFromBytes_RejectsTooSmall(t *testing.T) {
	if _, err := FromBytes([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected an error for a too-small input")
	}
}

func TestFromBytes_RejectsSizeMismatch(t *testing.T) {
	data := make([]byte, 84) // claims a triangle count without any triangle data
	data[80] = 1             // count = 1, but no 50-byte record follows
	if _, err := FromBytes(data); err == nil {
		t.Fatal("expected an error for a size/triangle-count mismatch")
	}
}
