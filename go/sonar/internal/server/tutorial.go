package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const tutorialScenarioFallbackImageURL = "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/scenario-undiscovered.png"

type tutorialOptionPayload struct {
	OptionText       string                       `json:"optionText"`
	StatTag          string                       `json:"statTag"`
	Difficulty       int                          `json:"difficulty"`
	RewardExperience int                          `json:"rewardExperience"`
	RewardGold       int                          `json:"rewardGold"`
	ItemRewards      []scenarioRewardItemPayload  `json:"itemRewards"`
	SpellRewards     []scenarioRewardSpellPayload `json:"spellRewards"`
}

type tutorialConfigRequest struct {
	CharacterID      *string                      `json:"characterId"`
	Dialogue         []string                     `json:"dialogue"`
	ScenarioPrompt   string                       `json:"scenarioPrompt"`
	ScenarioImageURL string                       `json:"scenarioImageUrl"`
	Options          []tutorialOptionPayload      `json:"options"`
	RewardExperience int                          `json:"rewardExperience"`
	RewardGold       int                          `json:"rewardGold"`
	ItemRewards      []scenarioRewardItemPayload  `json:"itemRewards"`
	SpellRewards     []scenarioRewardSpellPayload `json:"spellRewards"`
}

type tutorialStatusResponse struct {
	ShowWelcomeDialogue bool              `json:"showWelcomeDialogue"`
	ActivatedAt         interface{}       `json:"activatedAt"`
	CompletedAt         interface{}       `json:"completedAt"`
	ScenarioID          *uuid.UUID        `json:"scenarioId,omitempty"`
	Character           *models.Character `json:"character,omitempty"`
	Dialogue            []string          `json:"dialogue"`
}

type activateTutorialRequest struct {
	Force bool `json:"force"`
}

type generateTutorialImageRequest struct {
	ScenarioPrompt string `json:"scenarioPrompt"`
}

func (s *server) getTutorialConfig(ctx *gin.Context) {
	config, err := s.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, config)
}

func (s *server) updateTutorialConfig(ctx *gin.Context) {
	var requestBody tutorialConfigRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := parseTutorialConfigRequest(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := s.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	config.ImageGenerationStatus = existing.ImageGenerationStatus
	config.ImageGenerationError = existing.ImageGenerationError

	updated, err := s.dbClient.Tutorial().UpsertConfig(ctx, config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) generateTutorialImage(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if s.asyncClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "image generation is not configured"})
		return
	}

	var requestBody generateTutorialImageRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scenarioPrompt := strings.TrimSpace(requestBody.ScenarioPrompt)
	if scenarioPrompt == "" {
		config, err := s.dbClient.Tutorial().GetConfig(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		scenarioPrompt = strings.TrimSpace(config.ScenarioPrompt)
	}
	if scenarioPrompt == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "scenarioPrompt is required"})
		return
	}

	config, err := s.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	config.ImageGenerationStatus = models.TutorialImageGenerationStatusQueued
	config.ImageGenerationError = nil
	updated, err := s.dbClient.Tutorial().UpsertConfig(ctx, config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.GenerateTutorialImageTaskPayload{
		ScenarioPrompt: scenarioPrompt,
	})
	if err != nil {
		errMsg := err.Error()
		updated.ImageGenerationStatus = models.TutorialImageGenerationStatusFailed
		updated.ImageGenerationError = &errMsg
		_, _ = s.dbClient.Tutorial().UpsertConfig(ctx, updated)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateTutorialImageTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		updated.ImageGenerationStatus = models.TutorialImageGenerationStatusFailed
		updated.ImageGenerationError = &errMsg
		_, _ = s.dbClient.Tutorial().UpsertConfig(ctx, updated)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, updated)
}

func (s *server) getTutorialStatus(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	config, err := s.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	state, err := s.dbClient.Tutorial().FindStateByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if state == nil {
		ctx.JSON(http.StatusOK, tutorialStatusResponse{
			ShowWelcomeDialogue: false,
			Character:           config.Character,
			Dialogue:            append([]string{}, config.Dialogue...),
		})
		return
	}

	showWelcomeDialogue := state.ActivatedAt == nil &&
		state.CompletedAt == nil &&
		config.IsConfigured() &&
		config.Character != nil

	ctx.JSON(http.StatusOK, tutorialStatusResponse{
		ShowWelcomeDialogue: showWelcomeDialogue,
		ActivatedAt:         state.ActivatedAt,
		CompletedAt:         state.CompletedAt,
		ScenarioID:          state.TutorialScenarioID,
		Character:           config.Character,
		Dialogue:            append([]string{}, config.Dialogue...),
	})
}

func (s *server) activateTutorial(ctx *gin.Context) {
	var requestBody activateTutorialRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	config, err := s.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !config.IsConfigured() || config.Character == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "tutorial is not configured"})
		return
	}
	if err := s.dbClient.Tutorial().InitializeForNewUser(ctx, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userLat, userLng, err := s.getUserLatLng(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zones, err := s.dbClient.Zone().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	zone, err := selectZoneForCoordinates(zones, userLat, userLng)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ownerUserID := user.ID
	imageURL := strings.TrimSpace(config.ScenarioImageURL)
	if imageURL == "" {
		imageURL = tutorialScenarioFallbackImageURL
	}

	fallbackOptionItemRewards := tutorialScenarioItemRewards(config.ItemRewards)
	fallbackOptionSpellRewards := tutorialScenarioSpellRewards(config.SpellRewards)
	options := tutorialScenarioOptions(
		config.Options,
		config.RewardExperience,
		config.RewardGold,
		fallbackOptionItemRewards,
		fallbackOptionSpellRewards,
	)

	state, scenario, err := s.dbClient.Tutorial().ActivateForUser(
		ctx,
		user.ID,
		&models.Scenario{
			ZoneID:             zone.ID,
			OwnerUserID:        &ownerUserID,
			Latitude:           userLat,
			Longitude:          userLng,
			Prompt:             strings.TrimSpace(config.ScenarioPrompt),
			ImageURL:           imageURL,
			ThumbnailURL:       imageURL,
			Ephemeral:          true,
			ScaleWithUserLevel: false,
			RewardMode:         models.RewardModeExplicit,
			Difficulty:         0,
			RewardExperience:   0,
			RewardGold:         0,
			OpenEnded:          false,
			SuccessRewardMode:  models.ScenarioSuccessRewardModeIndividual,
		},
		options,
		nil,
		nil,
		requestBody.Force,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if state == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "tutorial is unavailable for this user"})
		return
	}
	if state.CompletedAt != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "tutorial already completed"})
		return
	}
	if scenario == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate tutorial scenario"})
		return
	}

	ctx.JSON(http.StatusOK, scenarioWithUserStatus{
		Scenario:        *scenario,
		AttemptedByUser: false,
	})
}

func parseTutorialConfigRequest(body tutorialConfigRequest) (*models.TutorialConfig, error) {
	config := &models.TutorialConfig{
		Dialogue:         []string{},
		ScenarioPrompt:   strings.TrimSpace(body.ScenarioPrompt),
		ScenarioImageURL: strings.TrimSpace(body.ScenarioImageURL),
		Options:          []models.TutorialScenarioOption{},
		ItemRewards:      []models.TutorialItemReward{},
		SpellRewards:     []models.TutorialSpellReward{},
		RewardExperience: body.RewardExperience,
		RewardGold:       body.RewardGold,
	}

	if body.CharacterID != nil {
		trimmed := strings.TrimSpace(*body.CharacterID)
		if trimmed != "" {
			characterID, err := uuid.Parse(trimmed)
			if err != nil {
				return nil, fmt.Errorf("characterId must be a valid UUID")
			}
			config.CharacterID = &characterID
		}
	}

	for _, line := range body.Dialogue {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			config.Dialogue = append(config.Dialogue, trimmed)
		}
	}

	for idx, option := range body.Options {
		optionText := strings.TrimSpace(option.OptionText)
		if optionText == "" {
			continue
		}
		statTag, ok := normalizeScenarioStatTag(option.StatTag)
		if !ok {
			return nil, fmt.Errorf("options[%d].statTag is invalid", idx)
		}
		rewardExperience := option.RewardExperience
		if rewardExperience < 0 {
			rewardExperience = 0
		}
		rewardGold := option.RewardGold
		if rewardGold < 0 {
			rewardGold = 0
		}
		difficulty := option.Difficulty
		if difficulty < 0 {
			difficulty = 0
		}
		config.Options = append(config.Options, models.TutorialScenarioOption{
			OptionText:       optionText,
			StatTag:          statTag,
			Difficulty:       difficulty,
			RewardExperience: rewardExperience,
			RewardGold:       rewardGold,
			ItemRewards:      parseTutorialItemRewardPayloads(option.ItemRewards),
			SpellRewards:     parseTutorialSpellRewardPayloads(option.SpellRewards),
		})
	}

	if config.RewardExperience < 0 {
		config.RewardExperience = 0
	}
	if config.RewardGold < 0 {
		config.RewardGold = 0
	}

	for _, reward := range body.ItemRewards {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		config.ItemRewards = append(config.ItemRewards, models.TutorialItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}

	for idx, reward := range body.SpellRewards {
		spellID := strings.TrimSpace(reward.SpellID)
		if spellID == "" {
			continue
		}
		if _, err := uuid.Parse(spellID); err != nil {
			return nil, fmt.Errorf("spellRewards[%d].spellId must be a valid UUID", idx)
		}
		config.SpellRewards = append(config.SpellRewards, models.TutorialSpellReward{
			SpellID: spellID,
		})
	}

	return config, nil
}

func tutorialScenarioItemRewards(input []models.TutorialItemReward) []models.ScenarioOptionItemReward {
	optionRewards := make([]models.ScenarioOptionItemReward, 0, len(input))
	for _, reward := range input {
		optionRewards = append(optionRewards, models.ScenarioOptionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return optionRewards
}

func tutorialScenarioSpellRewards(input []models.TutorialSpellReward) []models.ScenarioOptionSpellReward {
	optionRewards := make([]models.ScenarioOptionSpellReward, 0, len(input))
	for _, reward := range input {
		spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
		if err != nil {
			continue
		}
		optionRewards = append(optionRewards, models.ScenarioOptionSpellReward{
			SpellID: spellID,
		})
	}
	return optionRewards
}

func tutorialScenarioOptions(
	input []models.TutorialScenarioOption,
	fallbackRewardExperience int,
	fallbackRewardGold int,
	itemRewards []models.ScenarioOptionItemReward,
	spellRewards []models.ScenarioOptionSpellReward,
) []models.ScenarioOption {
	options := make([]models.ScenarioOption, 0, len(input))
	for _, option := range input {
		difficulty := option.Difficulty
		rewardExperience := option.RewardExperience
		rewardGold := option.RewardGold
		optionItemRewards := tutorialScenarioOptionItemRewards(option.ItemRewards)
		optionSpellRewards := tutorialScenarioOptionSpellRewards(option.SpellRewards)
		if tutorialOptionUsesLegacySharedRewards(option) {
			rewardExperience = fallbackRewardExperience
			rewardGold = fallbackRewardGold
			optionItemRewards = cloneScenarioOptionItemRewards(itemRewards)
			optionSpellRewards = cloneScenarioOptionSpellRewards(spellRewards)
		}
		options = append(options, models.ScenarioOption{
			OptionText:       option.OptionText,
			SuccessText:      "You step out ready for what comes next.",
			FailureText:      "You hesitate, and the moment slips away.",
			StatTag:          option.StatTag,
			Difficulty:       &difficulty,
			RewardExperience: rewardExperience,
			RewardGold:       rewardGold,
			ItemRewards:      optionItemRewards,
			SpellRewards:     optionSpellRewards,
		})
	}
	return options
}

func parseTutorialItemRewardPayloads(input []scenarioRewardItemPayload) []models.TutorialItemReward {
	out := make([]models.TutorialItemReward, 0, len(input))
	for _, reward := range input {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		out = append(out, models.TutorialItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func parseTutorialSpellRewardPayloads(input []scenarioRewardSpellPayload) []models.TutorialSpellReward {
	out := make([]models.TutorialSpellReward, 0, len(input))
	for _, reward := range input {
		spellID := strings.TrimSpace(reward.SpellID)
		if spellID == "" {
			continue
		}
		out = append(out, models.TutorialSpellReward{SpellID: spellID})
	}
	return out
}

func tutorialScenarioOptionItemRewards(input []models.TutorialItemReward) []models.ScenarioOptionItemReward {
	out := make([]models.ScenarioOptionItemReward, 0, len(input))
	for _, reward := range input {
		out = append(out, models.ScenarioOptionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func tutorialScenarioOptionSpellRewards(input []models.TutorialSpellReward) []models.ScenarioOptionSpellReward {
	out := make([]models.ScenarioOptionSpellReward, 0, len(input))
	for _, reward := range input {
		spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
		if err != nil {
			continue
		}
		out = append(out, models.ScenarioOptionSpellReward{
			SpellID: spellID,
		})
	}
	return out
}

func tutorialOptionUsesLegacySharedRewards(option models.TutorialScenarioOption) bool {
	return option.RewardExperience == 0 &&
		option.RewardGold == 0 &&
		len(option.ItemRewards) == 0 &&
		len(option.SpellRewards) == 0
}

func cloneScenarioOptionItemRewards(input []models.ScenarioOptionItemReward) []models.ScenarioOptionItemReward {
	out := make([]models.ScenarioOptionItemReward, 0, len(input))
	for _, reward := range input {
		out = append(out, models.ScenarioOptionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func cloneScenarioOptionSpellRewards(input []models.ScenarioOptionSpellReward) []models.ScenarioOptionSpellReward {
	out := make([]models.ScenarioOptionSpellReward, 0, len(input))
	for _, reward := range input {
		out = append(out, models.ScenarioOptionSpellReward{
			SpellID: reward.SpellID,
		})
	}
	return out
}

func scenarioVisibleToUser(userID uuid.UUID, scenario *models.Scenario) bool {
	if scenario == nil || scenario.OwnerUserID == nil {
		return true
	}
	return *scenario.OwnerUserID == userID
}
