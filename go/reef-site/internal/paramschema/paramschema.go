// Package paramschema is R-4.4's "single source of parameter truth" made
// real on the Go side: it validates a configuration's params against the
// same JSON Schema document (reef_parameter_schemas.schema) the TS client
// fetches at runtime to render its form. It intentionally implements only
// the subset of JSON Schema this repo's configurators actually use (type,
// required, minimum/maximum, enum) rather than pulling in a full external
// JSON Schema library for a handful of primitive-typed properties.
package paramschema

import (
	"encoding/json"
	"fmt"
	"sort"
)

type Schema struct {
	Type       interface{}         `json:"type"`
	Required   []string            `json:"required"`
	Properties map[string]Property `json:"properties"`
}

type Property struct {
	Type    interface{}   `json:"type"`
	Minimum *float64      `json:"minimum"`
	Maximum *float64      `json:"maximum"`
	Enum    []interface{} `json:"enum"`
	Default interface{}   `json:"default"`
}

// Error names the offending parameter (R-4.5: "the UI must state which
// parameter to change and why").
type Error struct {
	Parameter string
	Message   string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Parameter, e.Message)
}

// Parse decodes a stored schema document.
func Parse(schemaJSON []byte) (*Schema, error) {
	var s Schema
	if err := json.Unmarshal(schemaJSON, &s); err != nil {
		return nil, fmt.Errorf("paramschema: invalid schema document: %w", err)
	}
	return &s, nil
}

// Validate checks params against the schema, returning every violation
// found (not just the first) so a form can highlight every offending field
// at once. An empty, non-nil slice means valid.
func Validate(schema *Schema, params map[string]interface{}) []Error {
	var errs []Error

	for _, name := range schema.Required {
		if _, ok := params[name]; !ok {
			errs = append(errs, Error{Parameter: name, Message: "is required"})
		}
	}

	// Deterministic order for stable API responses / tests.
	names := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		prop := schema.Properties[name]
		value, present := params[name]
		if !present {
			continue // already reported above if required; optional+absent is fine
		}
		if err := validateProperty(name, prop, value); err != nil {
			errs = append(errs, *err)
		}
	}

	return errs
}

func validateProperty(name string, prop Property, value interface{}) *Error {
	if value == nil {
		if allowsNull(prop.Type) {
			return nil
		}
		return &Error{Parameter: name, Message: "must not be null"}
	}

	if len(prop.Enum) > 0 {
		if !enumContains(prop.Enum, value) {
			return &Error{Parameter: name, Message: fmt.Sprintf("must be one of %v", prop.Enum)}
		}
		// Enum already fully constrains valid values; numeric range checks
		// below would be redundant (and enum values aren't necessarily
		// numeric, e.g. color).
		return nil
	}

	switch typeName(prop.Type) {
	case "number", "integer":
		num, ok := asFloat(value)
		if !ok {
			return &Error{Parameter: name, Message: "must be a number"}
		}
		if prop.Minimum != nil && num < *prop.Minimum {
			return &Error{Parameter: name, Message: fmt.Sprintf("must be at least %g", *prop.Minimum)}
		}
		if prop.Maximum != nil && num > *prop.Maximum {
			return &Error{Parameter: name, Message: fmt.Sprintf("must be at most %g", *prop.Maximum)}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return &Error{Parameter: name, Message: "must be true or false"}
		}
	case "string":
		if _, ok := value.(string); !ok {
			return &Error{Parameter: name, Message: "must be a string"}
		}
	}

	return nil
}

func typeName(t interface{}) string {
	switch v := t.(type) {
	case string:
		return v
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "null" {
				return s
			}
		}
	}
	return ""
}

func allowsNull(t interface{}) bool {
	if list, ok := t.([]interface{}); ok {
		for _, item := range list {
			if s, ok := item.(string); ok && s == "null" {
				return true
			}
		}
	}
	return false
}

func enumContains(enum []interface{}, value interface{}) bool {
	valueFloat, valueIsNumber := asFloat(value)
	for _, candidate := range enum {
		if candidateFloat, ok := asFloat(candidate); ok && valueIsNumber {
			if candidateFloat == valueFloat {
				return true
			}
			continue
		}
		if candidate == value {
			return true
		}
	}
	return false
}

func asFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
