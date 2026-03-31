package server

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
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
