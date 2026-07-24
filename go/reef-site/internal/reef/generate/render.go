package generate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/reef-site/internal/reef/procexec"
)

type RenderConfig struct {
	OpenSCADBin string
	BaseTempDir string
	Timeout     time.Duration
	MemoryMB    int
}

type RenderResult struct {
	STLPath         string
	WorkDir         string
	OpenSCADVersion string
}

// Render writes scad to a temp .scad file and invokes OpenSCAD (R-2.4) to
// produce a binary STL (R-2.6: a compact binary mesh format — three.js's
// STLLoader consumes this directly for the preview viewer, so no separate
// GLB conversion step is needed). detail controls preview vs full-res via
// the $fn baked into scad by the caller (see Module.SCAD).
func Render(ctx context.Context, cfg RenderConfig, scad string) (*RenderResult, error) {
	workDir, err := procexec.NewWorkDir(cfg.BaseTempDir)
	if err != nil {
		return nil, err
	}

	inPath := filepath.Join(workDir, "in.scad")
	if err := os.WriteFile(inPath, []byte(scad), 0o600); err != nil {
		return nil, fmt.Errorf("generate: writing scad source: %w", err)
	}
	outPath := filepath.Join(workDir, "out.stl")

	_, err = procexec.Run(ctx, workDir, cfg.OpenSCADBin, []string{
		"--export-format=binstl",
		"-o", outPath,
		inPath,
	}, procexec.Limits{Timeout: cfg.Timeout, MemoryMB: cfg.MemoryMB})
	if err != nil {
		return nil, err
	}

	if _, statErr := os.Stat(outPath); statErr != nil {
		return nil, fmt.Errorf("generate: openscad reported success but %s was not produced: %w", outPath, statErr)
	}

	version, err := Version(ctx, cfg)
	if err != nil {
		// Version metadata failing to resolve shouldn't fail an otherwise
		// successful render — it's recorded for provenance (R-2.5), not
		// correctness of the geometry itself.
		version = "unknown"
	}

	return &RenderResult{STLPath: outPath, WorkDir: workDir, OpenSCADVersion: version}, nil
}

var versionPattern = regexp.MustCompile(`OpenSCAD version ([0-9][0-9A-Za-z.\-]*)`)

// Version returns the pinned OpenSCAD binary's version string, recorded on
// every artifact per R-2.5.
func Version(ctx context.Context, cfg RenderConfig) (string, error) {
	workDir, err := procexec.NewWorkDir(cfg.BaseTempDir)
	if err != nil {
		return "", err
	}
	defer procexec.Cleanup(workDir)

	result, err := procexec.Run(ctx, workDir, cfg.OpenSCADBin, []string{"--version"}, procexec.Limits{
		Timeout:  10 * time.Second,
		MemoryMB: 128,
	})
	if err != nil {
		return "", err
	}
	combined := string(result.Stdout) + string(result.Stderr)
	if m := versionPattern.FindStringSubmatch(combined); len(m) == 2 {
		return m[1], nil
	}
	return strings.TrimSpace(combined), nil
}
