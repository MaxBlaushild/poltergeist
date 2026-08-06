// Package set is bgi-site's contribution back to the shared platform
// (go/pkg/reef): composing several independently generated parts into one
// validated, priced product. Reef's single-part model never needed this —
// see go/bgi-site/PLATFORM_FINDINGS.md for the full reuse audit this
// package is the answer to.
package set

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/reef/generate"
	"github.com/google/uuid"
)

// ComponentManifest is R-2.4's "facts, not designs" — Assemble takes plain
// structs, not DB models, so it stays a pure function with no DB/subprocess
// dependency, independently unit-testable (mirrors generate.Module's own
// "pure function of params" discipline).
type ComponentManifest struct {
	ComponentType string
	CardWidthMm   float64
	CardHeightMm  float64
	Count         int
}

// TrayTemplate is the caller-supplied catalog of hand-designed templates
// (R-1.1) available to assemble from.
type TrayTemplate struct {
	ID              uuid.UUID
	ComponentType   string
	GeneratorModule string
}

type BoxProfile struct {
	InteriorLengthMm float64
	InteriorWidthMm  float64
	InteriorDepthMm  float64
}

type SleeveProfile struct {
	TotalCardThicknessMm float64
}

// ResolvedTray is one entry in an assembled recipe: Quantity identical
// copies of the tray described by Params, generated/cached/priced once
// regardless of how many the set needs (R-4.3's cross-config cache
// sharing — see go/bgi-site's configure.go for how geometry_hash caching
// composes with this).
type ResolvedTray struct {
	TrayTemplateID  uuid.UUID
	GeneratorModule string
	Params          map[string]interface{}
	Quantity        int
	HeightMm        float64 // one tray's own height, for display/debugging — Resolution.AssembledHeightMm is the stacked total
}

type Resolution struct {
	ResolvedTrays []ResolvedTray
	// UnassembledComponents lists component types present in the manifest
	// with no matching tray template — surfaced explicitly rather than
	// silently dropped (R-6.2's coverage rule).
	UnassembledComponents []string
	AssembledHeightMm     float64
	FitsBox               bool
}

// maxTraysPerComponent is the search cap: past this many trays for a single
// component type, Assemble gives up and reports a non-fit rather than
// searching forever.
const maxTraysPerComponent = 8

// Assemble composes a tray set (R-3.3, the module's one genuinely new hard
// problem): for each component type in the manifest, finds its matching
// template and searches upward for the fewest trays whose *individual*
// height fits the target box. Same-type trays sit side by side within the
// box's footprint at one shared height (splitting a deck across more,
// shorter trays is exactly what buys headroom against a shallow box) —
// only *different* component types stack on top of one another, so
// AssembledHeightMm sums one layer height per component type, not per
// tray. (Footprint/bin-packing across the box's length and width is not
// modeled in this v1 slice — see R-1.1's generative-layout deferral to v2.)
// Pure — no DB/subprocess calls, everything it needs is passed in — using
// each module's own Analyze() (analytical, no render) to learn a candidate
// split's height, the same way FragRack.ValidateParams catches an
// impossible combination before paying for a render+slice cycle.
func Assemble(manifests []ComponentManifest, templates []TrayTemplate, box BoxProfile, sleeve SleeveProfile, color string) (Resolution, error) {
	templateByType := make(map[string]TrayTemplate, len(templates))
	for _, t := range templates {
		templateByType[t.ComponentType] = t
	}

	// Deterministic order for stable test expectations and API responses.
	sorted := make([]ComponentManifest, len(manifests))
	copy(sorted, manifests)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ComponentType < sorted[j].ComponentType })

	var resolution Resolution
	allComponentsFit := true

	for _, m := range sorted {
		tmpl, ok := templateByType[m.ComponentType]
		if !ok {
			resolution.UnassembledComponents = append(resolution.UnassembledComponents, m.ComponentType)
			continue
		}

		module, err := generate.Get(tmpl.GeneratorModule)
		if err != nil {
			return Resolution{}, fmt.Errorf("set: resolve generator module %q for component %q: %w", tmpl.GeneratorModule, m.ComponentType, err)
		}

		componentFits := false
		for traysNeeded := 1; traysNeeded <= maxTraysPerComponent; traysNeeded++ {
			cardsPerTray := ceilDiv(m.Count, traysNeeded)
			params := trayParams(m, cardsPerTray, sleeve, color)

			analysis, err := module.Analyze(params)
			if err != nil {
				return Resolution{}, fmt.Errorf("set: analyze %s at %d tray(s): %w", m.ComponentType, traysNeeded, err)
			}

			fits := analysis.HeightMm <= box.InteriorDepthMm
			if fits || traysNeeded == maxTraysPerComponent {
				resolution.ResolvedTrays = append(resolution.ResolvedTrays, ResolvedTray{
					TrayTemplateID:  tmpl.ID,
					GeneratorModule: tmpl.GeneratorModule,
					Params:          params,
					Quantity:        traysNeeded,
					HeightMm:        analysis.HeightMm,
				})
				resolution.AssembledHeightMm += analysis.HeightMm
				componentFits = fits
				break
			}
		}
		if !componentFits {
			allComponentsFit = false
		}
	}

	// A component's own split can fit in isolation while the sum across
	// every component-type layer still exceeds the box — re-check the
	// total, not just each component's individual search result.
	resolution.FitsBox = allComponentsFit && resolution.AssembledHeightMm <= box.InteriorDepthMm
	return resolution, nil
}

func trayParams(m ComponentManifest, cardCount int, sleeve SleeveProfile, color string) map[string]interface{} {
	return map[string]interface{}{
		"cardWidthMm":          m.CardWidthMm,
		"cardHeightMm":         m.CardHeightMm,
		"cardCount":            float64(cardCount),
		"totalCardThicknessMm": sleeve.TotalCardThicknessMm,
		"color":                color,
	}
}

func ceilDiv(a, b int) int {
	if b <= 0 {
		return a
	}
	return (a + b - 1) / b
}

// ConfigHash is R-4.3's cache key, computed over the pure selection inputs
// — before any assembly/geometry work happens, so a cache hit skips
// Assemble entirely. Deliberately does not fold in resolved tray geometry
// hashes (a literal reading of the requirements doc's own formula would):
// those only exist after resolution runs, so hashing them would make the
// cache lookup depend on already having done the work it exists to skip.
func ConfigHash(gameSlug string, expansionSlugs []string, sleeveClassKey, boxSlug, color string) string {
	sorted := append([]string(nil), expansionSlugs...)
	sort.Strings(sorted)
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%s|%s", gameSlug, strings.Join(sorted, ","), sleeveClassKey, boxSlug, color)
	return hex.EncodeToString(h.Sum(nil))
}
