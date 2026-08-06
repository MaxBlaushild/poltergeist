package set

import (
	"testing"

	"github.com/google/uuid"
)

func terraformingMarsManifest() []ComponentManifest {
	return []ComponentManifest{
		{ComponentType: "project_card", CardWidthMm: 44, CardHeightMm: 68, Count: 208},
		{ComponentType: "corporation_card", CardWidthMm: 44, CardHeightMm: 68, Count: 12},
	}
}

func projectCardTemplate() TrayTemplate {
	return TrayTemplate{ID: uuid.New(), ComponentType: "project_card", GeneratorModule: "bgi_card_tray"}
}

func TestAssemble_ExactFit_ProducesResolvedTraysWithinBoxDepth(t *testing.T) {
	box := BoxProfile{InteriorLengthMm: 286, InteriorWidthMm: 286, InteriorDepthMm: 200}
	sleeve := SleeveProfile{TotalCardThicknessMm: 0.47} // standard sleeves

	res, err := Assemble(
		[]ComponentManifest{{ComponentType: "project_card", CardWidthMm: 44, CardHeightMm: 68, Count: 208}},
		[]TrayTemplate{projectCardTemplate()},
		box, sleeve, "black",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !res.FitsBox {
		t.Fatalf("expected a generous 200mm box to fit, got FitsBox=false (assembled height %.2f)", res.AssembledHeightMm)
	}
	if len(res.ResolvedTrays) != 1 {
		t.Fatalf("expected exactly 1 resolved tray entry (one template, one component type), got %d", len(res.ResolvedTrays))
	}
	if res.AssembledHeightMm <= 0 {
		t.Fatal("expected a positive assembled height")
	}
	if res.AssembledHeightMm > box.InteriorDepthMm {
		t.Fatalf("AssembledHeightMm (%.2f) must not exceed the box depth (%.2f) when FitsBox is true", res.AssembledHeightMm, box.InteriorDepthMm)
	}
}

func TestAssemble_NeedsMoreTrays_WhenShallowerBoxForcesASplit(t *testing.T) {
	// A single 208-card tray at standard sleeving is ~114.76mm tall
	// (208*0.47 + 15mm scoop headroom + 2mm floor) — shallow (80mm) must
	// force a split into shorter, side-by-side trays; deep (400mm) fits at
	// a single tray.
	shallow := BoxProfile{InteriorLengthMm: 286, InteriorWidthMm: 286, InteriorDepthMm: 80}
	deep := BoxProfile{InteriorLengthMm: 286, InteriorWidthMm: 286, InteriorDepthMm: 400}
	sleeve := SleeveProfile{TotalCardThicknessMm: 0.47}
	manifest := []ComponentManifest{{ComponentType: "project_card", CardWidthMm: 44, CardHeightMm: 68, Count: 208}}
	templates := []TrayTemplate{projectCardTemplate()}

	shallowRes, err := Assemble(manifest, templates, shallow, sleeve, "black")
	if err != nil {
		t.Fatal(err)
	}
	deepRes, err := Assemble(manifest, templates, deep, sleeve, "black")
	if err != nil {
		t.Fatal(err)
	}

	shallowQty := shallowRes.ResolvedTrays[0].Quantity
	deepQty := deepRes.ResolvedTrays[0].Quantity
	if shallowQty <= deepQty {
		t.Fatalf("expected the shallower box to need at least as many trays as the deep one: shallow=%d, deep=%d", shallowQty, deepQty)
	}
}

func TestAssemble_RejectsWhenEvenTheCapDoesNotFit(t *testing.T) {
	// A single card at absurd thickness in a near-zero-depth box: no split
	// up to maxTraysPerComponent can make even 1 card's well fit.
	box := BoxProfile{InteriorLengthMm: 286, InteriorWidthMm: 286, InteriorDepthMm: 1}
	sleeve := SleeveProfile{TotalCardThicknessMm: 5.0}
	manifest := []ComponentManifest{{ComponentType: "project_card", CardWidthMm: 44, CardHeightMm: 68, Count: 208}}

	res, err := Assemble(manifest, []TrayTemplate{projectCardTemplate()}, box, sleeve, "black")
	if err != nil {
		t.Fatal(err)
	}
	if res.FitsBox {
		t.Fatal("expected FitsBox=false when even the tray-count cap can't make it fit")
	}
	// Still reports a best-effort resolved tray (at the cap) so the caller
	// can explain by how much it's over, rather than an empty result.
	if len(res.ResolvedTrays) != 1 {
		t.Fatalf("expected a best-effort resolved tray even on rejection, got %d entries", len(res.ResolvedTrays))
	}
	if res.ResolvedTrays[0].Quantity != maxTraysPerComponent {
		t.Fatalf("expected the cap attempt to use maxTraysPerComponent=%d trays, got %d", maxTraysPerComponent, res.ResolvedTrays[0].Quantity)
	}
}

func TestAssemble_SurfacesUnassembledComponentsWithoutDroppingThem(t *testing.T) {
	box := BoxProfile{InteriorLengthMm: 286, InteriorWidthMm: 286, InteriorDepthMm: 200}
	sleeve := SleeveProfile{TotalCardThicknessMm: 0.47}

	// corporation_card has no matching template (Terraforming Mars v1's real
	// gap — see the seed migration's own comment).
	res, err := Assemble(terraformingMarsManifest(), []TrayTemplate{projectCardTemplate()}, box, sleeve, "black")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.UnassembledComponents) != 1 || res.UnassembledComponents[0] != "corporation_card" {
		t.Fatalf("expected corporation_card surfaced as unassembled, got %v", res.UnassembledComponents)
	}
	if len(res.ResolvedTrays) != 1 {
		t.Fatalf("expected project_card to still resolve normally despite corporation_card having no template, got %d resolved trays", len(res.ResolvedTrays))
	}
}

func TestAssemble_ErrorsOnUnregisteredGeneratorModule(t *testing.T) {
	box := BoxProfile{InteriorLengthMm: 286, InteriorWidthMm: 286, InteriorDepthMm: 200}
	sleeve := SleeveProfile{TotalCardThicknessMm: 0.47}
	manifest := []ComponentManifest{{ComponentType: "project_card", CardWidthMm: 44, CardHeightMm: 68, Count: 10}}
	badTemplate := []TrayTemplate{{ID: uuid.New(), ComponentType: "project_card", GeneratorModule: "does_not_exist"}}

	if _, err := Assemble(manifest, badTemplate, box, sleeve, "black"); err == nil {
		t.Fatal("expected an error for an unregistered generator module")
	}
}

// Acceptance criterion 4: two customers with overlapping expansions and the
// same sleeve/color should be able to share cached trays. Assemble itself
// doesn't cache (that's the config_hash/geometry_hash layer in
// go/bgi-site's configure.go), but it must produce byte-identical Params
// for byte-identical (manifest, sleeve, color) inputs regardless of which
// box profile is passed in — that determinism is exactly what makes
// cross-config geometry_hash sharing possible upstream.
func TestAssemble_ProducesIdenticalTrayParams_ForIdenticalManifestAndSleeveAcrossDifferentBoxes(t *testing.T) {
	manifest := []ComponentManifest{{ComponentType: "project_card", CardWidthMm: 44, CardHeightMm: 68, Count: 208}}
	templates := []TrayTemplate{projectCardTemplate()}
	sleeve := SleeveProfile{TotalCardThicknessMm: 0.47}

	deepA := BoxProfile{InteriorLengthMm: 286, InteriorWidthMm: 286, InteriorDepthMm: 400}
	deepB := BoxProfile{InteriorLengthMm: 300, InteriorWidthMm: 300, InteriorDepthMm: 400} // different footprint, same depth

	resA, err := Assemble(manifest, templates, deepA, sleeve, "black")
	if err != nil {
		t.Fatal(err)
	}
	resB, err := Assemble(manifest, templates, deepB, sleeve, "black")
	if err != nil {
		t.Fatal(err)
	}

	if resA.ResolvedTrays[0].Quantity != resB.ResolvedTrays[0].Quantity {
		t.Fatalf("expected the same tray split for two boxes with the same usable depth: A=%d, B=%d",
			resA.ResolvedTrays[0].Quantity, resB.ResolvedTrays[0].Quantity)
	}
	paramsA := resA.ResolvedTrays[0].Params
	paramsB := resB.ResolvedTrays[0].Params
	for _, key := range []string{"cardWidthMm", "cardHeightMm", "cardCount", "totalCardThicknessMm", "color"} {
		if paramsA[key] != paramsB[key] {
			t.Fatalf("param %q differs across configs that should share a cached tray: %v vs %v", key, paramsA[key], paramsB[key])
		}
	}
}

func TestConfigHash_SortsExpansionsForOrderIndependence(t *testing.T) {
	a := ConfigHash("terraforming-mars", []string{"prelude", "colonies"}, "standard", "original", "black")
	b := ConfigHash("terraforming-mars", []string{"colonies", "prelude"}, "standard", "original", "black")
	if a != b {
		t.Fatalf("expected ConfigHash to be order-independent over expansions, got %q vs %q", a, b)
	}
}

func TestConfigHash_DiffersWhenAnyInputDiffers(t *testing.T) {
	base := ConfigHash("terraforming-mars", nil, "standard", "original", "black")
	game := ConfigHash("wingspan", nil, "standard", "original", "black")
	sleeve := ConfigHash("terraforming-mars", nil, "thin", "original", "black")
	box := ConfigHash("terraforming-mars", nil, "standard", "aftermarket", "black")
	color := ConfigHash("terraforming-mars", nil, "standard", "original", "white")

	all := map[string]string{"game": game, "sleeve": sleeve, "box": box, "color": color}
	for label, hash := range all {
		if hash == base {
			t.Fatalf("expected changing %s to change the hash, both were %q", label, hash)
		}
	}
}
