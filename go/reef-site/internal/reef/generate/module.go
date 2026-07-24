// Package generate turns validated configurator parameters into OpenSCAD
// source (R-2.4) for each product's generator_module, and renders that
// source to an STL through the sandboxed subprocess runner (R-2.5). There is
// no client-side geometry code anywhere in this system — generation is
// server-authoritative and single-implementation, which is what structurally
// rules out preview/output drift (R-2.4).
package generate

import "fmt"

// Module is implemented once per product's generator_module (frag_rack,
// lid_clip, ...). R-1.1: a second configurator exists specifically to prove
// this abstraction generalizes — if adding lid_clip required touching
// anything outside its own file plus a registry entry, the abstraction is
// wrong.
type Module interface {
	// Slug matches reef_parameter_schemas.generator_module.
	Slug() string
	// Version matches reef_parameter_schemas.generator_version and feeds
	// geomhash.Hash — bump it whenever SCAD changes in a way that should
	// invalidate the cache for otherwise-identical params.
	Version() string
	// SCAD renders validated, defaulted params (already checked against the
	// JSON Schema by the validate package) into OpenSCAD source. detail
	// controls $fn: Preview uses a coarse value for fast, cheap meshes;
	// Full uses a print-quality value.
	SCAD(params map[string]interface{}, detail Detail) (string, error)
	// Analyze reports structural facts about the geometry SCAD would
	// produce for these params, computed analytically from the same
	// parametric math rather than by inspecting the mesh. This is the
	// generator's own answer for R-5.2's wall-thickness rule and R-5.3's
	// sealed-void rule — since generation here is single-implementation and
	// server-authoritative (R-2.4), the generator that built a cavity is
	// also the only thing that can say whether it vented it, which is exact
	// rather than a generic mesh-analysis estimate.
	Analyze(params map[string]interface{}) (Analysis, error)
}

// Analysis is a generator's self-report of structural facts about its own
// output for a given set of params.
type Analysis struct {
	MinWallMm   float64
	SealedVoid  bool
	DrainPathMm float64 // meaningful only when the generator has any internal cavities at all
}

type Detail int

const (
	Preview Detail = iota
	Full
)

func (d Detail) fn() int {
	if d == Preview {
		return 24
	}
	// 48-sided circles are already smoother than a 0.4mm nozzle can resolve
	// at these part sizes; $fn=96 render time scales roughly with the
	// square of $fn on a part with this many holes for negligible print
	// quality gain, so Full stays at 48 rather than reaching for the
	// OpenSCAD default "very smooth" ceiling.
	return 48
}

var registry = map[string]Module{}

func Register(m Module) {
	registry[m.Slug()] = m
}

func Get(slug string) (Module, error) {
	m, ok := registry[slug]
	if !ok {
		return nil, fmt.Errorf("generate: no module registered for generator_module %q", slug)
	}
	return m, nil
}

func init() {
	Register(&FragRack{})
	Register(&LidClip{})
}

// paramFloat/paramBool/paramString pull a typed value out of the decoded
// params map, erroring clearly rather than panicking on a type assertion —
// this is defense in depth on top of JSON-Schema validation, not a
// replacement for it (R-4.1/R-4.4 already keep bad values out upstream).
func paramFloat(params map[string]interface{}, key string) (float64, error) {
	v, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("generate: missing required parameter %q", key)
	}
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	default:
		return 0, fmt.Errorf("generate: parameter %q is not numeric (got %T)", key, v)
	}
}

func paramBool(params map[string]interface{}, key string, defaultValue bool) (bool, error) {
	v, ok := params[key]
	if !ok {
		return defaultValue, nil
	}
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("generate: parameter %q is not a boolean (got %T)", key, v)
	}
	return b, nil
}
