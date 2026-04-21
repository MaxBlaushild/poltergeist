package server

import (
	"strings"
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestNormalizeTutorialGuideSupportHistory(t *testing.T) {
	turns := []tutorialGuideSupportChatTurn{
		{Role: "system", Content: "ignore"},
		{Role: " user ", Content: "   First question   "},
		{Role: "assistant", Content: strings.Repeat("a", maxTutorialGuideSupportTurnCharacters+40)},
	}

	normalized := normalizeTutorialGuideSupportHistory(turns)
	if len(normalized) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(normalized))
	}
	if normalized[0].Role != "user" || normalized[0].Content != "First question" {
		t.Fatalf("unexpected first normalized turn: %+v", normalized[0])
	}
	if normalized[1].Role != "assistant" {
		t.Fatalf("expected assistant role, got %+v", normalized[1])
	}
	if len([]rune(normalized[1].Content)) != maxTutorialGuideSupportTurnCharacters {
		t.Fatalf("expected assistant content to be truncated to %d chars, got %d", maxTutorialGuideSupportTurnCharacters, len([]rune(normalized[1].Content)))
	}
}

func TestBuildTutorialGuideSupportPromptIncludesQuestContext(t *testing.T) {
	username := "streetwise"
	prompt := buildTutorialGuideSupportPrompt(
		&models.Character{
			Name:        "Mara",
			Description: "A calm scout who explains systems clearly.",
		},
		&models.User{
			ID:       uuid.New(),
			Name:     "Player One",
			Username: &username,
		},
		&models.UserTutorialState{
			Stage:       models.TutorialStageCompleted,
			CompletedAt: timePointer(time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)),
		},
		&models.TutorialConfig{
			GuideSupportPersonality: "Playful but reassuring. Speaks like a confident street mentor, never flippant.",
			GuideSupportBehavior:    "Lead with the clearest next step. Keep lore brief unless the player asks for it.",
			PostWelcomeDialogue: models.DialogueSequence{
				{Speaker: "character", Text: "Keep your quest log close.", Order: 0},
			},
		},
		&tutorialGuideSupportQuestContext{
			QuestName:        "Raise the Hearth",
			QuestDescription: "Claim your first real foothold.",
			Objective:        "Return to your base and upgrade the hearth.",
			IsAccepted:       true,
			IsTracked:        true,
		},
		[]tutorialGuideSupportChatTurn{
			{Role: "user", Content: "What should I do next?"},
		},
		"How do I upgrade the hearth?",
	)

	assertPromptContains(t, prompt, "You are Mara")
	assertPromptContains(t, prompt, "Player name: streetwise")
	assertPromptContains(t, prompt, "Guide personality:\nPlayful but reassuring. Speaks like a confident street mentor, never flippant.")
	assertPromptContains(t, prompt, "Guide support behavior:\nLead with the clearest next step. Keep lore brief unless the player asks for it.")
	assertPromptContains(t, prompt, "Follow-up quest: Raise the Hearth")
	assertPromptContains(t, prompt, "Current quest objective: Return to your base and upgrade the hearth.")
	assertPromptContains(t, prompt, "Player: What should I do next?")
	assertPromptContains(t, prompt, "Player question:\nHow do I upgrade the hearth?")
}

func TestFallbackTutorialGuideSupportAnswerUsesCharacterName(t *testing.T) {
	answer := fallbackTutorialGuideSupportAnswer(&models.Character{Name: "Mara"})
	if !strings.Contains(answer, "Mara pauses for a moment.") {
		t.Fatalf("expected character name in fallback answer, got %q", answer)
	}
}

func TestNormalizeTutorialGuideSupportAnswerExtractsResponseField(t *testing.T) {
	raw := `{"response":"Check your tracked quest and head back to your base hearth."}`
	answer := normalizeTutorialGuideSupportAnswer(raw)
	expected := "Check your tracked quest and head back to your base hearth."
	if answer != expected {
		t.Fatalf("expected %q, got %q", expected, answer)
	}
}

func assertPromptContains(t *testing.T, prompt string, expected string) {
	t.Helper()
	if !strings.Contains(prompt, expected) {
		t.Fatalf("expected prompt to contain %q, got:\n%s", expected, prompt)
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}
