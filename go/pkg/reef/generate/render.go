package generate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/reef/procexec"
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
//
// openscadVersion is a caller-resolved value (see Version), not re-resolved
// here: R-3.3's geometry_hash formula already requires every caller to know
// the version *before* it can even check the cache, so re-querying it again
// inside Render would just be a second, redundant subprocess call on every
// single generation.
func Render(ctx context.Context, cfg RenderConfig, scad string, openscadVersion string) (*RenderResult, error) {
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

	return &RenderResult{STLPath: outPath, WorkDir: workDir, OpenSCADVersion: openscadVersion}, nil
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
