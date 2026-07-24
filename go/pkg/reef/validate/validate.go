// Package validate implements R-5.2's six rejection rules as discrete,
// individually testable functions evaluated in the order the requirements
// list them. Thresholds are data (R-5.2: "thresholds in config, not
// literals"), passed in by the caller from the repo's config mechanism —
// nothing in this package hardcodes a limit.
package validate

import "fmt"

// Thresholds are the config-driven limits every rule checks against.
type Thresholds struct {
	MaxBboxMm      float64
	MinWallMm      float64
	MaxPrintTimeS  int64
	MaxWeightG     float64
	MinDrainPathMm float64
}

// Metadata is everything a rule needs to know about one generated,
// sliced part. BboxMaxDimensionMm and the Analyze-derived fields come from
// pure geometry (stlbbox, generate.Analysis); WeightG/PrintTimeS/
// SupportRequired come from the slicer (R-2.7).
type Metadata struct {
	BboxMaxDimensionMm float64
	SupportRequired    bool
	MinWallMm          float64
	PrintTimeS         int64
	WeightG            float64
	SealedVoid         bool
	DrainPathMm        float64
	HasInternalCavity  bool // only meaningful together with SealedVoid/DrainPathMm
}

// RuleName identifies which of the six rules fired, for rejection
// telemetry (R-5.4) and for the UI to key its "what to change" copy off of.
type RuleName string

const (
	RuleBoundingBox      RuleName = "bounding_box"
	RuleSupportsRequired RuleName = "supports_required"
	RuleMinWallThickness RuleName = "min_wall_thickness"
	RulePrintTime        RuleName = "print_time"
	RuleWeight           RuleName = "weight"
	RuleSealedVoid       RuleName = "sealed_void"
)

type Rejection struct {
	Rule   RuleName
	Reason string
}

type ruleFunc func(Metadata, Thresholds) *Rejection

// order matches R-5.2's numbered list exactly — Validate returns the first
// rule that fails, not every rule that would fail, since a UI can only walk
// a visitor through fixing one thing at a time.
var order = []ruleFunc{
	checkBoundingBox,
	checkSupportsRequired,
	checkMinWallThickness,
	checkPrintTime,
	checkWeight,
	checkSealedVoid,
}

// Validate runs all six rules in order and returns the first rejection, or
// nil if the part passes every rule.
func Validate(meta Metadata, thresholds Thresholds) *Rejection {
	for _, rule := range order {
		if rejection := rule(meta, thresholds); rejection != nil {
			return rejection
		}
	}
	return nil
}

// checkBoundingBox — R-5.2 rule 1.
func checkBoundingBox(meta Metadata, t Thresholds) *Rejection {
	if meta.BboxMaxDimensionMm <= t.MaxBboxMm {
		return nil
	}
	return &Rejection{
		Rule: RuleBoundingBox,
		Reason: fmt.Sprintf(
			"This part is %.1fmm on its longest axis, which is over the %.0fmm print envelope. Reduce a size parameter (width, tier count, or quantity) and try again.",
			meta.BboxMaxDimensionMm, t.MaxBboxMm,
		),
	}
}

// checkSupportsRequired — R-5.2 rule 2.
func checkSupportsRequired(meta Metadata, _ Thresholds) *Rejection {
	if !meta.SupportRequired {
		return nil
	}
	return &Rejection{
		Rule:   RuleSupportsRequired,
		Reason: "This configuration needs printed supports, which this catalog doesn't ship with. Try a smaller size or fewer tiers/holes so the geometry stays self-supporting.",
	}
}

// checkMinWallThickness — R-5.2 rule 3.
func checkMinWallThickness(meta Metadata, t Thresholds) *Rejection {
	if meta.MinWallMm >= t.MinWallMm {
		return nil
	}
	return &Rejection{
		Rule: RuleMinWallThickness,
		Reason: fmt.Sprintf(
			"A wall in this design would print at %.2fmm, below the %.1fmm minimum for a reliable part. Reduce the hole count or increase the overall size so there's more material between features.",
			meta.MinWallMm, t.MinWallMm,
		),
	}
}

// checkPrintTime — R-5.2 rule 4.
func checkPrintTime(meta Metadata, t Thresholds) *Rejection {
	if meta.PrintTimeS <= t.MaxPrintTimeS {
		return nil
	}
	return &Rejection{
		Rule: RulePrintTime,
		Reason: fmt.Sprintf(
			"Estimated print time is %.1f hours, over the %.1f hour limit. Reduce the size or quantity to bring it under the limit.",
			float64(meta.PrintTimeS)/3600, float64(t.MaxPrintTimeS)/3600,
		),
	}
}

// checkWeight — R-5.2 rule 5.
func checkWeight(meta Metadata, t Thresholds) *Rejection {
	if meta.WeightG <= t.MaxWeightG {
		return nil
	}
	return &Rejection{
		Rule: RuleWeight,
		Reason: fmt.Sprintf(
			"Estimated weight is %.0fg, over the %.0fg limit. Reduce the size or quantity to bring it under the limit.",
			meta.WeightG, t.MaxWeightG,
		),
	}
}

// checkSealedVoid — R-5.2 rule 6 / R-5.3. A part with no internal cavity at
// all trivially passes (there's nothing to trap air); a part with a cavity
// must have a drain path of at least MinDrainPathMm, or the generator must
// report it isn't sealed at all.
func checkSealedVoid(meta Metadata, t Thresholds) *Rejection {
	if !meta.HasInternalCavity {
		return nil
	}
	if !meta.SealedVoid && meta.DrainPathMm >= t.MinDrainPathMm {
		return nil
	}
	return &Rejection{
		Rule: RuleSealedVoid,
		Reason: fmt.Sprintf(
			"This design would trap air in an enclosed cavity (drain path %.1fmm, needs %.1fmm), making it buoyant and unstable underwater. Try a different tier count or magnet layout; if every option triggers this, it's a generator defect — contact support with the configuration link.",
			meta.DrainPathMm, t.MinDrainPathMm,
		),
	}
}
