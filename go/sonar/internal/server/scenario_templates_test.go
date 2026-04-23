package server

import "testing"

func TestParseScenarioTemplateUpsertRequestNormalizesZoneKind(t *testing.T) {
	s := &server{}

	template, err := s.parseScenarioTemplateUpsertRequest(nil, scenarioTemplateUpsertRequest{
		ZoneKind:  " Ancient Forest ",
		OpenEnded: true,
		Prompt:    "A hedge witch asks you to stop a poacher before the glade spirits retaliate.",
	}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if template.ZoneKind != "ancient-forest" {
		t.Fatalf("expected normalized zone kind, got %q", template.ZoneKind)
	}
}
