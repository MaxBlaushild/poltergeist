package server

import "testing"

func TestNormalizeUserStatusEffectTypeSupportsResourceTicks(t *testing.T) {
	if got := normalizeUserStatusEffectType("health_over_time"); got != "health_over_time" {
		t.Fatalf("expected health_over_time, got %s", got)
	}
	if got := normalizeUserStatusEffectType("mana_over_time"); got != "mana_over_time" {
		t.Fatalf("expected mana_over_time, got %s", got)
	}
}

func TestParseScenarioFailureStatusTemplatesSupportsHealthAndManaTicks(t *testing.T) {
	templates, err := parseScenarioFailureStatusTemplates([]scenarioFailureStatusPayload{
		{
			Name:            "Regeneration",
			EffectType:      "health_over_time",
			HealthPerTick:   4,
			DurationSeconds: 30,
		},
		{
			Name:            "Meditation",
			EffectType:      "mana_over_time",
			ManaPerTick:     3,
			DurationSeconds: 45,
		},
	}, "statuses")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}
	if templates[0].HealthPerTick != 4 || templates[0].EffectType != "health_over_time" {
		t.Fatalf("unexpected health-over-time template: %+v", templates[0])
	}
	if templates[1].ManaPerTick != 3 || templates[1].EffectType != "mana_over_time" {
		t.Fatalf("unexpected mana-over-time template: %+v", templates[1])
	}
}

func TestParseScenarioFailureStatusTemplatesRejectsZeroResourceTicks(t *testing.T) {
	_, err := parseScenarioFailureStatusTemplates([]scenarioFailureStatusPayload{
		{
			Name:            "Flatline",
			EffectType:      "health_over_time",
			DurationSeconds: 30,
		},
	}, "statuses")
	if err == nil {
		t.Fatal("expected error for zero healthPerTick")
	}

	_, err = parseScenarioFailureStatusTemplates([]scenarioFailureStatusPayload{
		{
			Name:            "Silence",
			EffectType:      "mana_over_time",
			DurationSeconds: 30,
		},
	}, "statuses")
	if err == nil {
		t.Fatal("expected error for zero manaPerTick")
	}
}
