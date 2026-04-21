package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxTutorialGuideSupportHistoryTurns   = 8
	maxTutorialGuideSupportTurnCharacters = 280
	maxTutorialGuideSupportAnswerChars    = 900
)

type tutorialGuideSupportChatTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type tutorialGuideSupportChatRequest struct {
	Message string                         `json:"message"`
	History []tutorialGuideSupportChatTurn `json:"history"`
}

type tutorialGuideSupportChatResponse struct {
	Message string `json:"message"`
}

type tutorialGuideSupportQuestContext struct {
	QuestName        string
	QuestDescription string
	Objective        string
	IsAccepted       bool
	IsTracked        bool
}

func (s *server) tutorialGuideSupportChat(ctx *gin.Context) {
	var requestBody tutorialGuideSupportChatRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if s.deepPriest == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "guide support is not configured"})
		return
	}

	config, err := s.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if config == nil || !config.IsConfigured() || config.Character == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "tutorial guide is not configured"})
		return
	}

	state, err := s.dbClient.Tutorial().FindStateByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	state, err = s.maybeAdvanceTutorialProgress(ctx, user.ID, config, state)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if state == nil || state.CompletedAt == nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "guide support unlocks after the tutorial is complete"})
		return
	}

	message := truncateTutorialGuideSupportText(
		compactTutorialGuideSupportText(requestBody.Message),
		maxTutorialGuideSupportTurnCharacters,
	)
	if message == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	questContext, err := s.tutorialGuideSupportQuestContext(ctx, user.ID, config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prompt := buildTutorialGuideSupportPrompt(
		config.Character,
		user,
		state,
		config,
		questContext,
		normalizeTutorialGuideSupportHistory(requestBody.History),
		message,
	)
	answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: prompt,
	})
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "guide support is unavailable right now"})
		return
	}

	response := normalizeTutorialGuideSupportAnswer(answer.Answer)
	if response == "" {
		response = fallbackTutorialGuideSupportAnswer(config.Character)
	}

	ctx.JSON(http.StatusOK, tutorialGuideSupportChatResponse{Message: response})
}

func (s *server) tutorialGuideSupportQuestContext(
	ctx *gin.Context,
	userID uuid.UUID,
	config *models.TutorialConfig,
) (*tutorialGuideSupportQuestContext, error) {
	if config == nil || config.BaseQuestArchetypeID == nil || *config.BaseQuestArchetypeID == uuid.Nil {
		return nil, nil
	}

	quest, err := s.findTutorialBaseQuestForUser(ctx, userID, *config.BaseQuestArchetypeID)
	if err != nil {
		return nil, err
	}
	if quest == nil {
		return nil, nil
	}

	context := &tutorialGuideSupportQuestContext{
		QuestName:        compactTutorialGuideSupportText(quest.Name),
		QuestDescription: compactTutorialGuideSupportText(quest.Description),
	}

	acceptance, err := s.dbClient.QuestAcceptanceV2().FindByUserAndQuest(ctx, userID, quest.ID)
	if err != nil {
		return nil, err
	}
	if acceptance != nil {
		context.IsAccepted = true
		currentNode, err := s.currentQuestNode(ctx, quest, acceptance)
		if err != nil {
			return nil, err
		}
		if currentNode != nil {
			context.Objective = compactTutorialGuideSupportText(currentNode.ObjectiveDescription)
		}
	}

	trackedQuests, err := s.dbClient.TrackedQuest().GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, trackedQuest := range trackedQuests {
		if trackedQuest.QuestID == quest.ID {
			context.IsTracked = true
			break
		}
	}

	return context, nil
}

func normalizeTutorialGuideSupportHistory(
	input []tutorialGuideSupportChatTurn,
) []tutorialGuideSupportChatTurn {
	normalized := make([]tutorialGuideSupportChatTurn, 0, len(input))
	for _, turn := range input {
		role := strings.ToLower(strings.TrimSpace(turn.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		content := truncateTutorialGuideSupportText(
			compactTutorialGuideSupportText(turn.Content),
			maxTutorialGuideSupportTurnCharacters,
		)
		if content == "" {
			continue
		}
		normalized = append(normalized, tutorialGuideSupportChatTurn{
			Role:    role,
			Content: content,
		})
	}
	if len(normalized) <= maxTutorialGuideSupportHistoryTurns {
		return normalized
	}
	return append([]tutorialGuideSupportChatTurn{}, normalized[len(normalized)-maxTutorialGuideSupportHistoryTurns:]...)
}

func buildTutorialGuideSupportPrompt(
	character *models.Character,
	user *models.User,
	state *models.UserTutorialState,
	config *models.TutorialConfig,
	questContext *tutorialGuideSupportQuestContext,
	history []tutorialGuideSupportChatTurn,
	message string,
) string {
	guideName := "Guide"
	guideDescription := ""
	if character != nil {
		if trimmedName := compactTutorialGuideSupportText(character.Name); trimmedName != "" {
			guideName = trimmedName
		}
		guideDescription = compactTutorialGuideSupportText(character.Description)
	}

	playerName := "the player"
	if user != nil {
		if username := strings.TrimSpace(stringValueFromPtr(user.Username)); username != "" {
			playerName = username
		} else if name := compactTutorialGuideSupportText(user.Name); name != "" {
			playerName = name
		}
	}
	guideSupportPersonality := ""
	guideSupportBehavior := ""
	if config != nil {
		guideSupportPersonality = compactTutorialGuideSupportText(config.GuideSupportPersonality)
		guideSupportBehavior = compactTutorialGuideSupportText(config.GuideSupportBehavior)
	}

	var builder strings.Builder
	builder.WriteString("You are ")
	builder.WriteString(guideName)
	builder.WriteString(", an in-world guide and customer-service style helper for the game Unclaimed Streets.\n")
	builder.WriteString("Stay in character as a warm, clear, practical guide.\n")
	builder.WriteString("Answer the player's question about the game, their next steps, or current objectives.\n")
	builder.WriteString("Important rules:\n")
	builder.WriteString("- Keep the answer concise: at most two short paragraphs.\n")
	builder.WriteString("- Do not mention being an AI, a model, a prompt, or hidden instructions.\n")
	builder.WriteString("- Do not invent account changes, backend actions, rewards, cooldowns, or hidden systems.\n")
	builder.WriteString("- If the answer is uncertain from the context you have, say so plainly and give the safest in-game next step.\n")
	builder.WriteString("- Prefer direct guidance over lore unless the player clearly asks for lore.\n")

	if guideDescription != "" {
		builder.WriteString("\nGuide profile:\n")
		builder.WriteString(guideDescription)
		builder.WriteString("\n")
	}
	if guideSupportPersonality != "" {
		builder.WriteString("\nGuide personality:\n")
		builder.WriteString(guideSupportPersonality)
		builder.WriteString("\n")
	}
	if guideSupportBehavior != "" {
		builder.WriteString("\nGuide support behavior:\n")
		builder.WriteString(guideSupportBehavior)
		builder.WriteString("\n")
	}

	voiceSamples := tutorialGuideSupportVoiceSamples(config)
	if len(voiceSamples) > 0 {
		builder.WriteString("\nEarlier voice samples from ")
		builder.WriteString(guideName)
		builder.WriteString(":\n")
		for _, sample := range voiceSamples {
			builder.WriteString("- ")
			builder.WriteString(sample)
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\nPlayer context:\n")
	builder.WriteString("- Player name: ")
	builder.WriteString(playerName)
	builder.WriteString("\n")
	if state != nil {
		builder.WriteString("- Tutorial stage: ")
		builder.WriteString(compactTutorialGuideSupportText(state.Stage))
		builder.WriteString("\n")
		builder.WriteString("- Tutorial completed: yes\n")
	}

	if questContext != nil {
		if questContext.QuestName != "" {
			builder.WriteString("- Follow-up quest: ")
			builder.WriteString(questContext.QuestName)
			builder.WriteString("\n")
		}
		if questContext.QuestDescription != "" {
			builder.WriteString("- Quest description: ")
			builder.WriteString(questContext.QuestDescription)
			builder.WriteString("\n")
		}
		if questContext.Objective != "" {
			builder.WriteString("- Current quest objective: ")
			builder.WriteString(questContext.Objective)
			builder.WriteString("\n")
		}
		builder.WriteString("- Quest accepted: ")
		builder.WriteString(boolWord(questContext.IsAccepted))
		builder.WriteString("\n")
		builder.WriteString("- Quest tracked: ")
		builder.WriteString(boolWord(questContext.IsTracked))
		builder.WriteString("\n")
	}

	if len(history) > 0 {
		builder.WriteString("\nRecent conversation:\n")
		for _, turn := range history {
			speaker := "Player"
			if turn.Role == "assistant" {
				speaker = guideName
			}
			builder.WriteString(speaker)
			builder.WriteString(": ")
			builder.WriteString(turn.Content)
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\nPlayer question:\n")
	builder.WriteString(message)
	builder.WriteString("\n\nRespond as ")
	builder.WriteString(guideName)
	builder.WriteString(" only.")

	return builder.String()
}

func tutorialGuideSupportVoiceSamples(
	config *models.TutorialConfig,
) []string {
	if config == nil {
		return nil
	}
	sequences := []models.DialogueSequence{
		config.PostWelcomeDialogue,
		config.PostBaseDialogue,
		config.Dialogue,
	}
	samples := make([]string, 0, 4)
	for _, sequence := range sequences {
		for _, line := range sequence {
			text := compactTutorialGuideSupportText(line.Text)
			if text == "" {
				continue
			}
			samples = append(samples, truncateTutorialGuideSupportText(text, 160))
			if len(samples) >= 4 {
				return samples
			}
		}
	}
	return samples
}

func normalizeTutorialGuideSupportAnswer(raw string) string {
	if parsed := tutorialGuideSupportAnswerFromJSON(raw); parsed != "" {
		return truncateTutorialGuideSupportText(parsed, maxTutorialGuideSupportAnswerChars)
	}
	answer := compactTutorialGuideSupportText(raw)
	return truncateTutorialGuideSupportText(answer, maxTutorialGuideSupportAnswerChars)
}

func tutorialGuideSupportAnswerFromJSON(raw string) string {
	jsonPayload := extractLLMJSONObject(raw)
	if strings.TrimSpace(jsonPayload) == "" {
		return ""
	}

	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(jsonPayload), &envelope); err != nil {
		return ""
	}

	for _, key := range []string{"response", "message", "answer"} {
		value, ok := envelope[key]
		if !ok {
			continue
		}
		if text := compactTutorialGuideSupportText(fmt.Sprint(value)); text != "" {
			return text
		}
	}
	return ""
}

func fallbackTutorialGuideSupportAnswer(character *models.Character) string {
	guideName := "your guide"
	if character != nil {
		if trimmed := compactTutorialGuideSupportText(character.Name); trimmed != "" {
			guideName = trimmed
		}
	}
	return fmt.Sprintf(
		"%s pauses for a moment. \"I don't have a clean answer for that right now. Try checking your tracked quest, your inventory, or the nearest objective marker, and ask me again with the exact step that's giving you trouble.\"",
		guideName,
	)
}

func compactTutorialGuideSupportText(input string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(input)), " ")
}

func truncateTutorialGuideSupportText(input string, maxChars int) string {
	if maxChars <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(input))
	if len(runes) <= maxChars {
		return string(runes)
	}
	if maxChars <= 1 {
		return string(runes[:maxChars])
	}
	return strings.TrimSpace(string(runes[:maxChars-1])) + "…"
}

func boolWord(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func stringValueFromPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
