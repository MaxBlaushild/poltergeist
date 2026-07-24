package validate

import (
	"strings"
	"testing"
)

func healthyMetadata() Metadata {
	return Metadata{
		BboxMaxDimensionMm: 150,
		SupportRequired:    false,
		MinWallMm:          6,
		PrintTimeS:         3600,
		WeightG:            80,
		SealedVoid:         false,
		DrainPathMm:        6,
		HasInternalCavity:  true,
	}
}

func defaultThresholds() Thresholds {
	return Thresholds{
		MaxBboxMm:      210,
		MinWallMm:      2.0,
		MaxPrintTimeS:  4 * 60 * 60,
		MaxWeightG:     250,
		MinDrainPathMm: 4,
	}
}

func TestValidate_HealthyPartPasses(t *testing.T) {
	if rejection := Validate(healthyMetadata(), defaultThresholds()); rejection != nil {
		t.Fatalf("expected a healthy part to pass, got rejection: %+v", rejection)
	}
}

// R-10 acceptance criterion: "Each of the six rejection rules in R-5.2 can
// be triggered deliberately and produces a message naming the parameter to
// change." Each subtest below deliberately violates exactly one rule.

func TestValidate_BoundingBoxRule(t *testing.T) {
	meta := healthyMetadata()
	meta.BboxMaxDimensionMm = 225
	rejection := Validate(meta, defaultThresholds())
	requireRejection(t, rejection, RuleBoundingBox, []string{"width", "tier", "quantity"})
}

func TestValidate_SupportsRequiredRule(t *testing.T) {
	meta := healthyMetadata()
	meta.SupportRequired = true
	rejection := Validate(meta, defaultThresholds())
	requireRejection(t, rejection, RuleSupportsRequired, []string{"size", "tiers", "holes"})
}

func TestValidate_MinWallThicknessRule(t *testing.T) {
	meta := healthyMetadata()
	meta.MinWallMm = 1.2
	rejection := Validate(meta, defaultThresholds())
	requireRejection(t, rejection, RuleMinWallThickness, []string{"hole", "size"})
}

func TestValidate_PrintTimeRule(t *testing.T) {
	meta := healthyMetadata()
	meta.PrintTimeS = 5 * 60 * 60
	rejection := Validate(meta, defaultThresholds())
	requireRejection(t, rejection, RulePrintTime, []string{"size", "quantity"})
}

func TestValidate_WeightRule(t *testing.T) {
	meta := healthyMetadata()
	meta.WeightG = 300
	rejection := Validate(meta, defaultThresholds())
	requireRejection(t, rejection, RuleWeight, []string{"size", "quantity"})
}

func TestValidate_SealedVoidRule(t *testing.T) {
	meta := healthyMetadata()
	meta.SealedVoid = true
	meta.DrainPathMm = 0
	rejection := Validate(meta, defaultThresholds())
	requireRejection(t, rejection, RuleSealedVoid, []string{"tier", "magnet"})
}

func TestValidate_SealedVoidRule_IgnoredWhenNoCavityExists(t *testing.T) {
	meta := healthyMetadata()
	meta.HasInternalCavity = false
	meta.SealedVoid = true // shouldn't matter if there's no cavity at all
	meta.DrainPathMm = 0
	if rejection := Validate(meta, defaultThresholds()); rejection != nil {
		t.Fatalf("expected no rejection when HasInternalCavity is false, got %+v", rejection)
	}
}

// R-5.2: rules run "in this order" — a part failing multiple rules should
// report the first one, not the last.
func TestValidate_ReturnsFirstFailingRuleInOrder(t *testing.T) {
	meta := healthyMetadata()
	meta.BboxMaxDimensionMm = 225 // rule 1
	meta.WeightG = 300            // rule 5 — should not surface if rule 1 already failed
	rejection := Validate(meta, defaultThresholds())
	if rejection == nil {
		t.Fatal("expected a rejection")
	}
	if rejection.Rule != RuleBoundingBox {
		t.Fatalf("expected the first failing rule (%s) to win, got %s", RuleBoundingBox, rejection.Rule)
	}
}

func requireRejection(t *testing.T, rejection *Rejection, wantRule RuleName, anyOfWords []string) {
	t.Helper()
	if rejection == nil {
		t.Fatalf("expected a rejection for rule %s, got none", wantRule)
	}
	if rejection.Rule != wantRule {
		t.Fatalf("Rule = %s, want %s", rejection.Rule, wantRule)
	}
	if rejection.Reason == "" {
		t.Fatal("Reason must not be empty")
	}
	lower := strings.ToLower(rejection.Reason)
	found := false
	for _, w := range anyOfWords {
		if strings.Contains(lower, w) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Reason %q does not name any expected parameter from %v", rejection.Reason, anyOfWords)
	}
}
