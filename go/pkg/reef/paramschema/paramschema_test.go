package paramschema

import "testing"

const fragRackSchemaJSON = `
{
  "type": "object",
  "required": ["glassThicknessMm", "tierCount", "widthMm", "plugHoleDiameterMm", "holesPerTier", "color"],
  "properties": {
    "tankProfileId": { "type": ["string", "null"] },
    "glassThicknessMm": { "type": "number", "minimum": 4, "maximum": 19 },
    "tierCount": { "type": "integer", "minimum": 1, "maximum": 4 },
    "widthMm": { "type": "number", "minimum": 60, "maximum": 250 },
    "plugHoleDiameterMm": { "type": "integer", "enum": [15, 20] },
    "holesPerTier": { "type": "integer", "minimum": 4, "maximum": 12 },
    "color": { "type": "string", "enum": ["black", "white"], "default": "black" }
  }
}
`

func mustParseFragRackSchema(t *testing.T) *Schema {
	t.Helper()
	s, err := Parse([]byte(fragRackSchemaJSON))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestValidate_AcceptsHealthyParams(t *testing.T) {
	s := mustParseFragRackSchema(t)
	params := map[string]interface{}{
		"tankProfileId":      nil,
		"glassThicknessMm":   10.0,
		"tierCount":          2.0,
		"widthMm":            150.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerTier":       6.0,
		"color":              "black",
	}
	if errs := Validate(s, params); len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidate_RejectsMissingRequired(t *testing.T) {
	s := mustParseFragRackSchema(t)
	params := map[string]interface{}{
		"glassThicknessMm": 10.0,
		// widthMm intentionally omitted
		"tierCount":          2.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerTier":       6.0,
		"color":              "black",
	}
	errs := Validate(s, params)
	if !containsParameter(errs, "widthMm") {
		t.Fatalf("expected an error naming widthMm, got %v", errs)
	}
}

func TestValidate_RejectsOutOfRangeNumber(t *testing.T) {
	s := mustParseFragRackSchema(t)
	params := healthyParams()
	params["glassThicknessMm"] = 25.0 // over the 19mm max
	errs := Validate(s, params)
	if !containsParameter(errs, "glassThicknessMm") {
		t.Fatalf("expected an error naming glassThicknessMm, got %v", errs)
	}
}

func TestValidate_RejectsValueNotInEnum(t *testing.T) {
	s := mustParseFragRackSchema(t)
	params := healthyParams()
	params["plugHoleDiameterMm"] = 17.0 // not 15 or 20
	errs := Validate(s, params)
	if !containsParameter(errs, "plugHoleDiameterMm") {
		t.Fatalf("expected an error naming plugHoleDiameterMm, got %v", errs)
	}
}

func TestValidate_RejectsWrongType(t *testing.T) {
	s := mustParseFragRackSchema(t)
	params := healthyParams()
	params["widthMm"] = "wide"
	errs := Validate(s, params)
	if !containsParameter(errs, "widthMm") {
		t.Fatalf("expected an error naming widthMm, got %v", errs)
	}
}

func TestValidate_AllowsNullForNullableTankProfile(t *testing.T) {
	s := mustParseFragRackSchema(t)
	params := healthyParams()
	params["tankProfileId"] = nil
	if errs := Validate(s, params); len(errs) != 0 {
		t.Fatalf("expected nil to be valid for a nullable field, got %v", errs)
	}
}

func TestValidate_ReportsMultipleErrorsAtOnce(t *testing.T) {
	s := mustParseFragRackSchema(t)
	params := healthyParams()
	params["glassThicknessMm"] = 100.0
	params["plugHoleDiameterMm"] = 99.0
	errs := Validate(s, params)
	if len(errs) < 2 {
		t.Fatalf("expected at least 2 errors reported at once, got %v", errs)
	}
}

func healthyParams() map[string]interface{} {
	return map[string]interface{}{
		"tankProfileId":      nil,
		"glassThicknessMm":   10.0,
		"tierCount":          2.0,
		"widthMm":            150.0,
		"plugHoleDiameterMm": 20.0,
		"holesPerTier":       6.0,
		"color":              "black",
	}
}

func containsParameter(errs []Error, param string) bool {
	for _, e := range errs {
		if e.Parameter == param {
			return true
		}
	}
	return false
}
