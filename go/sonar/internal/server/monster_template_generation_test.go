package server

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestNextUniqueMonsterTemplateNamePrefersPrefixesBeforeNumericSuffix(t *testing.T) {
	used := map[string]struct{}{
		"goblin skirmisher":     {},
		"ash goblin skirmisher": {},
	}

	got := nextUniqueMonsterTemplateName(
		"Goblin Skirmisher",
		used,
		[]string{"Ash", "Briar"},
	)
	if got != "Briar Goblin Skirmisher" {
		t.Fatalf("expected prefixed fallback name, got %q", got)
	}
	if strings.Contains(got, " 2") {
		t.Fatalf("expected non-numeric fallback name, got %q", got)
	}
}

func TestBuildBulkMonsterTemplateSpecsFromSeedsAvoidsOverusedSeedFamilies(t *testing.T) {
	used := map[string]struct{}{
		"goblin skirmisher":       {},
		"ash goblin skirmisher":   {},
		"briar goblin skirmisher": {},
		"goblin skirmisher 2":     {},
	}

	specs := buildBulkMonsterTemplateSpecsFromSeeds(
		1,
		used,
		uuid.MustParse("00000000-0000-0000-0000-000000000123"),
		models.MonsterTemplateTypeMonster,
		nil,
		nil,
	)
	if len(specs) != 1 {
		t.Fatalf("expected one spec, got %d", len(specs))
	}

	if strings.Contains(strings.ToLower(specs[0].Name), "goblin skirmisher") {
		t.Fatalf("expected less-used fallback family, got %q", specs[0].Name)
	}
}
