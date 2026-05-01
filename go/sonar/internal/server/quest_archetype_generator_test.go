package server

import (
	"strings"
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestBuildGeneratedChallengeTemplateQuestionStaysConcrete(t *testing.T) {
	step := normalizedQuestTemplateGeneratorStep{
		Content: questTemplateGeneratorContentChallenge,
		LocationArchetype: &models.LocationArchetype{
			Name: "Bustling City Square",
		},
	}
	nextStep := &normalizedQuestTemplateGeneratorStep{
		Content: questTemplateGeneratorContentScenario,
	}

	question := buildGeneratedChallengeTemplateQuestion("a missing courier trail", step, nextStep)
	lower := strings.ToLower(question)

	if !strings.Contains(lower, "photograph") {
		t.Fatalf("expected concrete photo task, got %q", question)
	}
	if strings.Contains(lower, "clue") || strings.Contains(lower, "search") || strings.Contains(lower, "investigate") {
		t.Fatalf("expected question to avoid impossible clue-hunting language, got %q", question)
	}
}

func TestSelectQuestMonsterTemplateMatchesPrefersQuestThemedExistingTemplates(t *testing.T) {
	matches := selectQuestMonsterTemplateMatches(
		[]models.MonsterTemplate{
			{
				ID:          uuid.New(),
				MonsterType: models.MonsterTemplateTypeMonster,
				Name:        "Shadow Courier",
				Description: "A nimble alley runner who strikes from rooftops and smoke.",
			},
			{
				ID:          uuid.New(),
				MonsterType: models.MonsterTemplateTypeMonster,
				Name:        "Forest Boar",
				Description: "A tusked beast that tramples anything in its path.",
			},
		},
		questMonsterTemplateRequest{
			Count:            1,
			MonsterType:      models.MonsterTemplateTypeMonster,
			ThemePrompt:      "smuggler ring in moonlit alleys",
			EncounterConcept: "rooftop courier lookout",
			LocationConcept:  "shadowed alley",
		},
		1,
	)

	if len(matches) != 1 {
		t.Fatalf("expected one match, got %d", len(matches))
	}
	if matches[0].template.Name != "Shadow Courier" {
		t.Fatalf("expected Shadow Courier to be selected, got %q", matches[0].template.Name)
	}
}

func TestSelectQuestMonsterTemplateMatchesPrefersMatchingZoneKind(t *testing.T) {
	preferredZoneKind := &models.ZoneKind{Slug: "forest", Name: "Forest"}
	matches := selectQuestMonsterTemplateMatches(
		[]models.MonsterTemplate{
			{
				ID:          uuid.New(),
				MonsterType: models.MonsterTemplateTypeMonster,
				ZoneKind:    "desert",
				Name:        "Trail Stalker",
				Description: "A hunter that shadows travelers between waystones.",
			},
			{
				ID:          uuid.New(),
				MonsterType: models.MonsterTemplateTypeMonster,
				ZoneKind:    "forest",
				Name:        "Canopy Stalker",
				Description: "A hunter that shadows travelers between waystones.",
			},
		},
		questMonsterTemplateRequest{
			Count:             1,
			MonsterType:       models.MonsterTemplateTypeMonster,
			PreferredZoneKind: preferredZoneKind,
			ThemePrompt:       "ambush on a caravan trail",
			EncounterConcept:  "predator stalking from cover",
			LocationConcept:   "mossy woodland path",
		},
		1,
	)

	if len(matches) != 1 {
		t.Fatalf("expected one match, got %d", len(matches))
	}
	if matches[0].template.Name != "Canopy Stalker" {
		t.Fatalf("expected forest-matched template to be selected, got %q", matches[0].template.Name)
	}
}

func TestSelectQuestMonsterTemplateMatchesPenalizesVeryRecentMatches(t *testing.T) {
	now := time.Now()
	matches := selectQuestMonsterTemplateMatches(
		[]models.MonsterTemplate{
			{
				ID:          uuid.New(),
				CreatedAt:   now.Add(-2 * 24 * time.Hour),
				MonsterType: models.MonsterTemplateTypeMonster,
				ZoneKind:    "harbor",
				Name:        "Harbor Stalker",
				Description: "A hunter that shadows smugglers through dockside lanes.",
			},
			{
				ID:          uuid.New(),
				CreatedAt:   now.Add(-180 * 24 * time.Hour),
				MonsterType: models.MonsterTemplateTypeMonster,
				ZoneKind:    "harbor",
				Name:        "Dockside Stalker",
				Description: "A hunter that shadows smugglers through dockside lanes.",
			},
		},
		questMonsterTemplateRequest{
			Count:             1,
			MonsterType:       models.MonsterTemplateTypeMonster,
			PreferredZoneKind: &models.ZoneKind{Slug: "harbor", Name: "Harbor"},
			ThemePrompt:       "dockside smuggling trail",
			EncounterConcept:  "predator stalking smugglers through the lanes",
			LocationConcept:   "dockside lane",
		},
		1,
	)

	if len(matches) != 1 {
		t.Fatalf("expected one match, got %d", len(matches))
	}
	if matches[0].template.Name != "Dockside Stalker" {
		t.Fatalf("expected older harbor match to win, got %q", matches[0].template.Name)
	}
}

func TestBuildQuestMonsterFallbackSpecsFromRequestUsesContextualRoleSeeds(t *testing.T) {
	specs := buildQuestMonsterFallbackSpecsFromRequest(
		2,
		map[string]struct{}{},
		questMonsterTemplateRequest{
			Count:             2,
			MonsterType:       models.MonsterTemplateTypeMonster,
			PreferredZoneKind: &models.ZoneKind{Slug: "harbor", Name: "Harbor"},
			ThemePrompt:       "shadow trade along the docks",
			EncounterConcept:  "ambushers controlling the tide gates",
			LocationConcept:   "dockside canal",
			EncounterTone:     []string{"shadowed", "territorial"},
		},
	)

	if len(specs) != 2 {
		t.Fatalf("expected two fallback specs, got %d", len(specs))
	}
	for _, spec := range specs {
		if spec.ZoneKind != "harbor" {
			t.Fatalf("expected harbor zone kind, got %q", spec.ZoneKind)
		}
		if strings.Contains(strings.ToLower(spec.Name), "goblin") {
			t.Fatalf("expected contextual role-seed fallback, got %q", spec.Name)
		}
	}
	if !strings.Contains(strings.ToLower(specs[0].Name), "harbor") {
		t.Fatalf("expected first fallback name to carry a contextual prefix, got %q", specs[0].Name)
	}
}
