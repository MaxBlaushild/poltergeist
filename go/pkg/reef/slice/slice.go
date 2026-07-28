// Package slice wraps a slicer CLI (PrusaSlicer or OrcaSlicer — both share
// the same PrusaSlicer-derived CLI and G-code comment format) as the source
// of truth for weight, print time, and printability (R-2.7). Nothing in
// this package estimates those numbers; it slices real G-code and reads the
// statistics the slicer itself computed.
//
// Verified against a real PrusaSlicer 2.7.4 binary (the exec-invocation
// path was originally written blind — see git history — since package
// install was blocked in the authoring environment). No custom print
// profile is loaded (Config.ConfigIni is never set by any caller as of this
// writing), so support-material decisions come from PrusaSlicer's generic
// bundled defaults, not settings tuned for a specific printer/filament —
// worth knowing if support percentages look higher than expected for an
// otherwise self-supporting design. Config.FilamentDensityGCm3 is set
// independently of ConfigIni, though, since without it PrusaSlicer reports
// every part's weight as exactly 0.00g (confirmed against the real
// binary) — that's the one setting that isn't optional.
package slice

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/reef/procexec"
)

type Config struct {
	SlicerBin       string
	ConfigIni       string // path to a PrusaSlicer print/filament/printer profile
	BaseTempDir     string
	Timeout         time.Duration
	MemoryMB        int
	SupportsEnabled bool // slice once with supports auto-enabled to detect whether the model needs them

	// FilamentDensityGCm3 (g/cm³) is required for PrusaSlicer to convert
	// extruded volume into weight at all — without it (and with no
	// ConfigIni loaded to supply one), PrusaSlicer's generic bundled
	// defaults report every part's "total filament used [g]" as exactly
	// 0.00, silently zeroing out the material-cost component of every
	// price (see pricing.Price). Confirmed against a real PrusaSlicer 2.7.4
	// binary: identical slice, 0.00g with this unset vs. a real weight with
	// it set to PETG's ~1.24-1.27 g/cm³.
	FilamentDensityGCm3 float64
}

type Result struct {
	WeightG                float64
	PrintTimeS             int64
	SupportMaterialPercent float64 // of total extrusion; 0 if none was needed
	SlicerVersion          string
	GCodePath              string
}

// Slice slices stlPath to G-code and returns the statistics PrusaSlicer
// embeds as trailing comments. Slicing is run with support material
// auto-detection on: SupportMaterialPercent is the real fraction of total
// extrusion the slicer spent on support material for this specific model —
// not a heuristic in Go, and not just a boolean, since a design needing a
// trivial dusting of support is a different situation than one needing
// scaffolding through most of the print.
func Slice(ctx context.Context, cfg Config, stlPath string) (*Result, error) {
	workDir, err := procexec.NewWorkDir(cfg.BaseTempDir)
	if err != nil {
		return nil, err
	}
	gcodePath := workDir + "/out.gcode"

	args := []string{
		"--export-gcode",
		"--support-material",
		"--support-material-auto",
		"--gcode-comments",
	}
	if cfg.ConfigIni != "" {
		args = append(args, "--load", cfg.ConfigIni)
	}
	if cfg.FilamentDensityGCm3 > 0 {
		args = append(args, "--filament-density", strconv.FormatFloat(cfg.FilamentDensityGCm3, 'g', -1, 64))
	}
	args = append(args, "-o", gcodePath, stlPath)

	if _, err := procexec.Run(ctx, workDir, cfg.SlicerBin, args, procexec.Limits{
		Timeout:  cfg.Timeout,
		MemoryMB: cfg.MemoryMB,
	}); err != nil {
		return nil, err
	}

	gcode, err := os.ReadFile(gcodePath)
	if err != nil {
		return nil, fmt.Errorf("slice: slicer reported success but %s was not produced: %w", gcodePath, err)
	}

	result, err := ParseGCodeStats(string(gcode))
	if err != nil {
		return nil, fmt.Errorf("slice: parsing slicer output: %w", err)
	}
	result.GCodePath = gcodePath

	version, err := Version(ctx, cfg)
	if err != nil {
		version = "unknown"
	}
	result.SlicerVersion = version

	return result, nil
}

func Version(ctx context.Context, cfg Config) (string, error) {
	workDir, err := procexec.NewWorkDir(cfg.BaseTempDir)
	if err != nil {
		return "", err
	}
	defer procexec.Cleanup(workDir)

	result, err := procexec.Run(ctx, workDir, cfg.SlicerBin, []string{"--help"}, procexec.Limits{
		Timeout:  10 * time.Second,
		MemoryMB: 512,
	})
	// PrusaSlicer prints its version banner to stdout on --help even though
	// --help itself isn't a "successful" slice invocation; treat KindFailed
	// (nonzero exit on --help alone, seen on some builds) as non-fatal here
	// and still try to parse whatever it printed.
	var stdout, stderr string
	if result != nil {
		stdout, stderr = string(result.Stdout), string(result.Stderr)
	}
	if err != nil {
		if procErr, ok := asProcexecError(err); !ok || procErr.Kind != "failed" {
			return "", err
		}
	}
	if v := versionPattern.FindString(stdout + stderr); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("slice: could not find a version string in slicer output")
}

func asProcexecError(err error) (*procexec.Error, bool) {
	pe, ok := err.(*procexec.Error)
	return pe, ok
}

var versionPattern = regexp.MustCompile(`(?:PrusaSlicer|OrcaSlicer|Slic3r)[^\n]*?\d+\.\d+[\.\d]*[^\s]*`)

var (
	filamentGramsPattern = regexp.MustCompile(`(?m)^;\s*total filament used \[g\]\s*=\s*([0-9.]+)`)
	printTimePattern     = regexp.MustCompile(`(?m)^;\s*estimated printing time \(normal mode\)\s*=\s*(.+)$`)
	typeMarkerPattern    = regexp.MustCompile(`^;TYPE:(.+)$`)
	extrusionMovePattern = regexp.MustCompile(`^G[01]\b.*\bE(-?[0-9.]+)`)
	resetExtruderPattern = regexp.MustCompile(`^G92\b.*\bE(-?[0-9.]+)`)
)

// ParseGCodeStats extracts weight/time/support-usage from a PrusaSlicer (or
// OrcaSlicer, which uses the same comment conventions) G-code footer. This
// is the piece of the slicer integration that's fully testable without the
// binary — see slice_test.go for the canned sample this is verified against.
func ParseGCodeStats(gcode string) (*Result, error) {
	weightMatch := filamentGramsPattern.FindStringSubmatch(gcode)
	if weightMatch == nil {
		return nil, fmt.Errorf("slice: could not find 'total filament used [g]' in gcode output")
	}
	weightG, err := strconv.ParseFloat(weightMatch[1], 64)
	if err != nil {
		return nil, fmt.Errorf("slice: parsing filament weight %q: %w", weightMatch[1], err)
	}

	timeMatch := printTimePattern.FindStringSubmatch(gcode)
	if timeMatch == nil {
		return nil, fmt.Errorf("slice: could not find 'estimated printing time' in gcode output")
	}
	printTimeS, err := parseSlicerDuration(timeMatch[1])
	if err != nil {
		return nil, fmt.Errorf("slice: parsing print time %q: %w", timeMatch[1], err)
	}

	return &Result{
		WeightG:                weightG,
		PrintTimeS:             printTimeS,
		SupportMaterialPercent: computeSupportMaterialPercent(gcode),
	}, nil
}

// computeSupportMaterialPercent walks every extrusion move, tracking which
// ;TYPE: section it falls in, and returns what fraction of total extrusion
// happened inside a "Support material"/"Support material interface"
// section. This is a real measurement from the slicer's own toolpath, not
// an estimate — R-5.2's "needs supports" rule used to be a bare boolean
// (any support at all rejects the part), which meant a design needing a
// trivial dusting of support material was rejected exactly the same as one
// needing scaffolding through its entire height. A percentage lets the
// threshold distinguish "structurally sound design, one small overhang"
// from "this really doesn't work without a printer sitting there filling it
// with support."
func computeSupportMaterialPercent(gcode string) float64 {
	var currentType string
	var lastE, totalExtrusion, supportExtrusion float64

	for _, line := range strings.Split(gcode, "\n") {
		line = strings.TrimSpace(line)

		if m := typeMarkerPattern.FindStringSubmatch(line); m != nil {
			currentType = m[1]
			continue
		}
		if m := resetExtruderPattern.FindStringSubmatch(line); m != nil {
			if v, err := strconv.ParseFloat(m[1], 64); err == nil {
				lastE = v
			}
			continue
		}
		m := extrusionMovePattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		e, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			continue
		}
		if delta := e - lastE; delta > 0 {
			totalExtrusion += delta
			if strings.HasPrefix(currentType, "Support material") {
				supportExtrusion += delta
			}
		}
		lastE = e
	}

	if totalExtrusion <= 0 {
		return 0
	}
	return supportExtrusion / totalExtrusion * 100
}

// parseSlicerDuration parses PrusaSlicer's "1d 2h 3m 4s" style duration
// (any subset of the four units, always in that order).
func parseSlicerDuration(s string) (int64, error) {
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(`(\d+)([dhms])`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("no recognizable duration components in %q", s)
	}
	var total int64
	for _, m := range matches {
		n, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return 0, err
		}
		switch m[2] {
		case "d":
			total += n * 24 * 60 * 60
		case "h":
			total += n * 60 * 60
		case "m":
			total += n * 60
		case "s":
			total += n
		}
	}
	return total, nil
}
