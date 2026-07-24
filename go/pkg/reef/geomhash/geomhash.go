// Package geomhash implements R-3.3's geometry_hash: a cache key that must be
// stable for identical logical inputs, so identical configurations never
// regenerate or re-slice. The whole point of this package is that a hash
// which varies by float formatting silently destroys that cache — so
// canonicalization is written once, here, with tests, rather than left to
// each caller to get right.
package geomhash

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
)

// numericPrecision is the grid floats are rounded to before formatting.
// Configurator parameters are physical millimeter/gram measurements; nothing
// in this system needs more than 4 decimal places of precision, and rounding
// to a fixed grid is what makes 90, 90.0, and 90.00000000003 (the kind of
// noise that shows up after a derived-parameter computation) hash identically.
const numericPrecision = 4

// Hash computes geometry_hash = sha256(productSlug + generatorVersion +
// openscadVersion + canonicalJSON(params)) per R-3.3. params must be a JSON
// object (map at the top level); anything else is rejected since a
// parameter payload that isn't an object has no business being hashed as one.
func Hash(productSlug, generatorVersion, openscadVersion string, params json.RawMessage) (string, error) {
	canon, err := CanonicalJSON(params)
	if err != nil {
		return "", fmt.Errorf("geomhash: %w", err)
	}

	h := sha256.New()
	for _, part := range []string{productSlug, generatorVersion, openscadVersion} {
		h.Write([]byte(part))
		h.Write([]byte{0})
	}
	h.Write(canon)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CanonicalJSON re-serializes arbitrary JSON with sorted object keys and
// normalized numeric formatting, so byte-identical output implies
// logically-identical input regardless of key order or float noise.
func CanonicalJSON(raw json.RawMessage) ([]byte, error) {
	var v interface{}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("decoding params: %w", err)
	}
	if _, ok := v.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("params must be a JSON object, got %T", v)
	}

	var buf bytes.Buffer
	if err := writeCanonical(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeCanonical(buf *bytes.Buffer, v interface{}) error {
	switch val := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case json.Number:
		writeCanonicalNumber(buf, val)
	case string:
		encoded, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(encoded)
	case []interface{}:
		buf.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeCanonical(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return err
			}
			buf.Write(keyBytes)
			buf.WriteByte(':')
			if err := writeCanonical(buf, val[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	default:
		return fmt.Errorf("unsupported JSON value type %T", v)
	}
	return nil
}

func writeCanonicalNumber(buf *bytes.Buffer, n json.Number) {
	f, err := n.Float64()
	if err != nil {
		// Not representable as a float64 (shouldn't happen from
		// encoding/json's own decoder) — fall back to the raw literal
		// rather than losing information.
		buf.WriteString(string(n))
		return
	}
	rounded := roundTo(f, numericPrecision)
	buf.WriteString(strconv.FormatFloat(rounded, 'f', -1, 64))
}

func roundTo(f float64, decimals int) float64 {
	scale := math.Pow(10, float64(decimals))
	return math.Round(f*scale) / scale
}
