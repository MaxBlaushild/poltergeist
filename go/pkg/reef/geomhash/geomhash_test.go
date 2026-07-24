package geomhash

import (
	"encoding/json"
	"testing"
)

// R-3.3: "a hash that varies by float formatting silently destroys the
// cache" — this is the test that guards against that regressing.
func TestHash_StableAcrossFloatFormatting(t *testing.T) {
	variants := []string{
		`{"widthMm": 90, "tierCount": 2}`,
		`{"widthMm": 90.0, "tierCount": 2}`,
		`{"widthMm": 90.00000000003, "tierCount": 2}`,
		`{"widthMm": 89.99999999997, "tierCount": 2}`,
		`{"tierCount": 2, "widthMm": 90}`, // key order must not matter either
	}

	var hashes []string
	for _, v := range variants {
		hash, err := Hash("magnetic-frag-rack", "v1", "openscad-2021.01", json.RawMessage(v))
		if err != nil {
			t.Fatalf("Hash(%s) error: %v", v, err)
		}
		hashes = append(hashes, hash)
	}

	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Fatalf("variant %d (%s) hashed to %s, want %s (variant 0: %s)",
				i, variants[i], hashes[i], hashes[0], variants[0])
		}
	}
}

func TestHash_DifferentValuesHashDifferently(t *testing.T) {
	a, err := Hash("magnetic-frag-rack", "v1", "openscad-2021.01", json.RawMessage(`{"widthMm": 90}`))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Hash("magnetic-frag-rack", "v1", "openscad-2021.01", json.RawMessage(`{"widthMm": 91}`))
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatalf("expected different hashes for different widthMm, both got %s", a)
	}
}

func TestHash_DifferentProductOrVersionHashesDifferently(t *testing.T) {
	base, err := Hash("magnetic-frag-rack", "v1", "openscad-2021.01", json.RawMessage(`{"widthMm": 90}`))
	if err != nil {
		t.Fatal(err)
	}
	otherProduct, err := Hash("lid-mesh-clips", "v1", "openscad-2021.01", json.RawMessage(`{"widthMm": 90}`))
	if err != nil {
		t.Fatal(err)
	}
	otherGeneratorVersion, err := Hash("magnetic-frag-rack", "v2", "openscad-2021.01", json.RawMessage(`{"widthMm": 90}`))
	if err != nil {
		t.Fatal(err)
	}
	otherOpenSCADVersion, err := Hash("magnetic-frag-rack", "v1", "openscad-2023.06", json.RawMessage(`{"widthMm": 90}`))
	if err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{base: true}
	for _, h := range []string{otherProduct, otherGeneratorVersion, otherOpenSCADVersion} {
		if seen[h] {
			t.Fatalf("expected a distinct hash, got a collision: %s", h)
		}
		seen[h] = true
	}
}

func TestHash_NestedStructuresAndArrays(t *testing.T) {
	a, err := Hash("p", "v1", "os1", json.RawMessage(`{"a":{"z":1,"y":2},"b":[1,2,3]}`))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Hash("p", "v1", "os1", json.RawMessage(`{"b":[1,2,3],"a":{"y":2,"z":1}}`))
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("expected reordering nested object keys and top-level keys to be a no-op, got %s vs %s", a, b)
	}
}

func TestHash_RejectsNonObjectParams(t *testing.T) {
	if _, err := Hash("p", "v1", "os1", json.RawMessage(`[1,2,3]`)); err == nil {
		t.Fatal("expected an error for non-object params, got nil")
	}
	if _, err := Hash("p", "v1", "os1", json.RawMessage(`"just a string"`)); err == nil {
		t.Fatal("expected an error for non-object params, got nil")
	}
}

func TestCanonicalJSON_IsDeterministicBytes(t *testing.T) {
	out, err := CanonicalJSON(json.RawMessage(`{"b":2,"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"a":1,"b":2}` {
		t.Fatalf("got %s, want {\"a\":1,\"b\":2}", out)
	}
}
