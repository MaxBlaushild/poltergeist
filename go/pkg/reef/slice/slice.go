// Package slice wraps a slicer CLI (PrusaSlicer or OrcaSlicer — both share
// the same PrusaSlicer-derived CLI and G-code comment format) as the source
// of truth for weight, print time, and printability (R-2.7). Nothing in
// this package estimates those numbers; it slices real G-code and reads the
// statistics the slicer itself computed.
//
// Caveat, stated plainly: the exec-invocation path here (Slice/Version) was
// written against PrusaSlicer's documented CLI and its well-known G-code
// comment footer format, but could not be exercised against a real
// PrusaSlicer/OrcaSlicer binary in the environment that authored it (package
// install was blocked by an unrelated broken apt mirror snapshot — see
// go/reef-site/INVENTORY.md). The G-code comment parser (ParseGCodeStats)
// *is* fully unit-tested against realistic sample output. Run the slicer
// integration tests once a real binary is available before trusting this in
// production.
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
}

type Result struct {
	WeightG         float64
	PrintTimeS      int64
	SupportRequired bool
	SlicerVersion   string
	GCodePath       string
}

// Slice slices stlPath to G-code and returns the statistics PrusaSlicer
// embeds as trailing comments. Slicing is run with support material
// auto-detection on: if the slicer decides the model needs support and
// generates any, SupportRequired is true — this is a real per-model
// determination by the slicer, not a heuristic in Go.
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
	supportTypePattern   = regexp.MustCompile(`(?m)^;TYPE:Support material`)
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
		WeightG:         weightG,
		PrintTimeS:      printTimeS,
		SupportRequired: supportTypePattern.MatchString(gcode),
	}, nil
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
