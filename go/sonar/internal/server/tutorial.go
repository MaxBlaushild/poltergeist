package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const tutorialScenarioFallbackImageURL = "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/scenario-undiscovered.png"

type tutorialOptionPayload struct {
	OptionText string `json:"optionText"`
	StatTag    string `json:"statTag"`
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

	updated, err := s.dbClient.Tutorial().UpsertConfig(ctx, config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) getTutorialStatus(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
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
			Dialogue:            []string{},
		})
		return
	}

	config, err := s.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	itemRewards, optionItemRewards := tutorialScenarioItemRewards(config.ItemRewards)
	spellRewards, optionSpellRewards := tutorialScenarioSpellRewards(config.SpellRewards)
	options := tutorialScenarioOptions(
		config.Options,
		config.RewardExperience,
		config.RewardGold,
		optionItemRewards,
		optionSpellRewards,
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
			RewardExperience:   config.RewardExperience,
			RewardGold:         config.RewardGold,
			OpenEnded:          false,
		},
		options,
		itemRewards,
		spellRewards,
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
		config.Options = append(config.Options, models.TutorialScenarioOption{
			OptionText: optionText,
			StatTag:    statTag,
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

func tutorialScenarioItemRewards(input []models.TutorialItemReward) ([]models.ScenarioItemReward, []models.ScenarioItemReward) {
	scenarioRewards := make([]models.ScenarioItemReward, 0, len(input))
	optionRewards := make([]models.ScenarioItemReward, 0, len(input))
	for _, reward := range input {
		scenarioRewards = append(scenarioRewards, models.ScenarioItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
		optionRewards = append(optionRewards, models.ScenarioItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return scenarioRewards, optionRewards
}

func tutorialScenarioSpellRewards(input []models.TutorialSpellReward) ([]models.ScenarioSpellReward, []models.ScenarioSpellReward) {
	scenarioRewards := make([]models.ScenarioSpellReward, 0, len(input))
	optionRewards := make([]models.ScenarioSpellReward, 0, len(input))
	for _, reward := range input {
		spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
		if err != nil {
			continue
		}
		scenarioRewards = append(scenarioRewards, models.ScenarioSpellReward{
			SpellID: spellID,
		})
		optionRewards = append(optionRewards, models.ScenarioSpellReward{
			SpellID: spellID,
		})
	}
	return scenarioRewards, optionRewards
}

func tutorialScenarioOptions(
	input []models.TutorialScenarioOption,
	rewardExperience int,
	rewardGold int,
	itemRewards []models.ScenarioItemReward,
	spellRewards []models.ScenarioSpellReward,
) []models.ScenarioOption {
	options := make([]models.ScenarioOption, 0, len(input))
	for _, option := range input {
		options = append(options, models.ScenarioOption{
			OptionText:       option.OptionText,
			SuccessText:      "You step out ready for what comes next.",
			FailureText:      "You hesitate, and the moment slips away.",
			StatTag:          option.StatTag,
			RewardExperience: rewardExperience,
			RewardGold:       rewardGold,
			ItemRewards:      cloneScenarioOptionItemRewards(itemRewards),
			SpellRewards:     cloneScenarioOptionSpellRewards(spellRewards),
		})
	}
	return options
}

func cloneScenarioOptionItemRewards(input []models.ScenarioItemReward) []models.ScenarioOptionItemReward {
	out := make([]models.ScenarioOptionItemReward, 0, len(input))
	for _, reward := range input {
		out = append(out, models.ScenarioOptionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func cloneScenarioOptionSpellRewards(input []models.ScenarioSpellReward) []models.ScenarioOptionSpellReward {
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
