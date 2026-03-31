package server

import (
	"strings"
	"testing"

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
