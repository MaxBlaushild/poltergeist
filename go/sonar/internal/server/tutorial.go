package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	dungeonmasterruntime "github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
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
	CharacterID                       *string                      `json:"characterId"`
	BaseQuestArchetypeID              *string                      `json:"baseQuestArchetypeId"`
	BaseQuestGiverCharacterID         *string                      `json:"baseQuestGiverCharacterId"`
	BaseQuestGiverCharacterTemplateID *string                      `json:"baseQuestGiverCharacterTemplateId"`
	Dialogue                          []models.DialogueMessage     `json:"dialogue"`
	PostWelcomeDialogue               []models.DialogueMessage     `json:"postWelcomeDialogue"`
	ScenarioObjectiveCopy             string                       `json:"scenarioObjectiveCopy"`
	PostScenarioDialogue              []models.DialogueMessage     `json:"postScenarioDialogue"`
	LoadoutDialogue                   []models.DialogueMessage     `json:"loadoutDialogue"`
	LoadoutObjectiveCopy              string                       `json:"loadoutObjectiveCopy"`
	PostMonsterDialogue               []models.DialogueMessage     `json:"postMonsterDialogue"`
	BaseKitDialogue                   []models.DialogueMessage     `json:"baseKitDialogue"`
	BaseKitObjectiveCopy              string                       `json:"baseKitObjectiveCopy"`
	PostBasePlacementDialogue         []models.DialogueMessage     `json:"postBasePlacementDialogue"`
	HearthObjectiveCopy               string                       `json:"hearthObjectiveCopy"`
	PostBaseDialogue                  []models.DialogueMessage     `json:"postBaseDialogue"`
	ScenarioPrompt                    string                       `json:"scenarioPrompt"`
	ScenarioImageURL                  string                       `json:"scenarioImageUrl"`
	Options                           []tutorialOptionPayload      `json:"options"`
	MonsterEncounterID                *string                      `json:"monsterEncounterId"`
	MonsterObjectiveCopy              string                       `json:"monsterObjectiveCopy"`
	MonsterRewardExperience           int                          `json:"monsterRewardExperience"`
	MonsterRewardGold                 int                          `json:"monsterRewardGold"`
	MonsterItemRewards                []scenarioRewardItemPayload  `json:"monsterItemRewards"`
	RewardExperience                  int                          `json:"rewardExperience"`
	RewardGold                        int                          `json:"rewardGold"`
	ItemRewards                       []scenarioRewardItemPayload  `json:"itemRewards"`
	SpellRewards                      []scenarioRewardSpellPayload `json:"spellRewards"`
}

type tutorialStatusResponse struct {
	ShowWelcomeDialogue       bool                     `json:"showWelcomeDialogue"`
	Stage                     string                   `json:"stage"`
	ActivatedAt               interface{}              `json:"activatedAt"`
	CompletedAt               interface{}              `json:"completedAt"`
	ScenarioID                *uuid.UUID               `json:"scenarioId,omitempty"`
	MonsterEncounterID        *uuid.UUID               `json:"monsterEncounterId,omitempty"`
	Character                 *models.Character        `json:"character,omitempty"`
	Dialogue                  []models.DialogueMessage `json:"dialogue"`
	PostWelcomeDialogue       []models.DialogueMessage `json:"postWelcomeDialogue"`
	ScenarioObjectiveCopy     string                   `json:"scenarioObjectiveCopy"`
	PostScenarioDialogue      []models.DialogueMessage `json:"postScenarioDialogue"`
	LoadoutDialogue           []models.DialogueMessage `json:"loadoutDialogue"`
	LoadoutObjectiveCopy      string                   `json:"loadoutObjectiveCopy"`
	PostMonsterDialogue       []models.DialogueMessage `json:"postMonsterDialogue"`
	BaseKitDialogue           []models.DialogueMessage `json:"baseKitDialogue"`
	BaseKitObjectiveCopy      string                   `json:"baseKitObjectiveCopy"`
	PostBasePlacementDialogue []models.DialogueMessage `json:"postBasePlacementDialogue"`
	HearthObjectiveCopy       string                   `json:"hearthObjectiveCopy"`
	PostBaseDialogue          []models.DialogueMessage `json:"postBaseDialogue"`
	MonsterObjectiveCopy      string                   `json:"monsterObjectiveCopy"`
	RequiredEquipItemIDs      []int                    `json:"requiredEquipItemIds"`
	CompletedEquipItemIDs     []int                    `json:"completedEquipItemIds"`
	RequiredUseItemIDs        []int                    `json:"requiredUseItemIds"`
	CompletedUseItemIDs       []int                    `json:"completedUseItemIds"`
}

type activateTutorialRequest struct {
	Force bool `json:"force"`
}

type advanceTutorialRequest struct {
	Action string `json:"action"`
}

type generateTutorialImageRequest struct {
	ScenarioPrompt string `json:"scenarioPrompt"`
}

type adminInstantiateTutorialBaseQuestRequest struct {
	UserID string `json:"userId"`
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

func (s *server) adminInstantiateTutorialBaseQuest(ctx *gin.Context) {
	var requestBody adminInstantiateTutorialBaseQuestRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(strings.TrimSpace(requestBody.UserID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userId must be a valid UUID"})
		return
	}

	config, err := s.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if config == nil || config.BaseQuestArchetypeID == nil ||
		(config.BaseQuestGiverCharacterID == nil && config.BaseQuestGiverCharacterTemplateID == nil) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "tutorial home base quest is not fully configured"})
		return
	}

	user, err := s.dbClient.User().FindByID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	base, err := s.dbClient.Base().FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if base == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "selected user does not have a home base"})
		return
	}

	s.instantiateTutorialBaseQuestAsync(userID, base, config)
	ctx.JSON(http.StatusAccepted, gin.H{"queued": true})
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
		ctx.JSON(http.StatusOK, buildTutorialStatusResponse(config, nil))
		return
	}

	state, err = s.maybeAdvanceTutorialProgress(ctx, user.ID, config, state)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, buildTutorialStatusResponse(config, state))
}

func (s *server) resetTutorial(ctx *gin.Context) {
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
	if err := s.forceResetTutorialReplayState(ctx, user.ID, config); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	state, err := s.dbClient.Tutorial().FindStateByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, buildTutorialStatusResponse(config, state))
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
	if requestBody.Force {
		if err := s.forceResetTutorialReplayState(ctx, user.ID, config); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if err := s.dbClient.Tutorial().InitializeForNewUser(ctx, user.ID); err != nil {
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
		false,
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

func (s *server) advanceTutorial(ctx *gin.Context) {
	var requestBody advanceTutorialRequest
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

	action := strings.TrimSpace(strings.ToLower(requestBody.Action))
	switch action {
	case "welcome_dialogue_closed":
		if err := s.dbClient.Tutorial().AdvanceToPostWelcomeDialogue(ctx, user.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "post_scenario_dialogue_closed":
		if _, err := s.advanceTutorialToMonsterOrComplete(ctx, user.ID, config); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "post_monster_dialogue_closed":
		requiredUseItemIDs, err := s.tutorialBaseKitRequirementItemIDs(ctx, config)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if len(requiredUseItemIDs) > 0 {
			if err := s.dbClient.Tutorial().AdvanceToBaseKit(ctx, user.ID, requiredUseItemIDs); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			if err := s.dbClient.Tutorial().AdvanceToBaseKit(ctx, user.ID, []int{}); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if err := s.dbClient.Tutorial().AdvanceToPostBaseDialogue(ctx, user.ID); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	case "post_base_placement_dialogue_closed":
		if err := s.dbClient.Tutorial().AdvanceToPostBaseDialogue(ctx, user.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "post_base_dialogue_closed":
		if err := s.dbClient.Tutorial().MarkCompleted(ctx, user.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported tutorial action"})
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
	ctx.JSON(http.StatusOK, buildTutorialStatusResponse(config, state))
}

func buildTutorialStatusResponse(
	config *models.TutorialConfig,
	state *models.UserTutorialState,
) tutorialStatusResponse {
	if config == nil {
		config = &models.TutorialConfig{}
	}
	if state == nil {
		return tutorialStatusResponse{
			ShowWelcomeDialogue:       false,
			Stage:                     models.TutorialStageCompleted,
			Character:                 config.Character,
			Dialogue:                  append([]models.DialogueMessage{}, config.Dialogue...),
			PostWelcomeDialogue:       append([]models.DialogueMessage{}, config.PostWelcomeDialogue...),
			ScenarioObjectiveCopy:     config.ScenarioObjectiveCopy,
			PostScenarioDialogue:      append([]models.DialogueMessage{}, config.PostScenarioDialogue...),
			LoadoutDialogue:           append([]models.DialogueMessage{}, config.LoadoutDialogue...),
			LoadoutObjectiveCopy:      config.LoadoutObjectiveCopy,
			PostMonsterDialogue:       append([]models.DialogueMessage{}, config.PostMonsterDialogue...),
			BaseKitDialogue:           append([]models.DialogueMessage{}, config.BaseKitDialogue...),
			BaseKitObjectiveCopy:      config.BaseKitObjectiveCopy,
			PostBasePlacementDialogue: append([]models.DialogueMessage{}, config.PostBasePlacementDialogue...),
			HearthObjectiveCopy:       config.HearthObjectiveCopy,
			PostBaseDialogue:          append([]models.DialogueMessage{}, config.PostBaseDialogue...),
			MonsterObjectiveCopy:      config.MonsterObjectiveCopy,
			RequiredEquipItemIDs:      []int{},
			CompletedEquipItemIDs:     []int{},
			RequiredUseItemIDs:        []int{},
			CompletedUseItemIDs:       []int{},
		}
	}

	showWelcomeDialogue := state.Stage == models.TutorialStageWelcome &&
		state.ActivatedAt == nil &&
		state.CompletedAt == nil &&
		config.IsConfigured() &&
		config.Character != nil

	return tutorialStatusResponse{
		ShowWelcomeDialogue:       showWelcomeDialogue,
		Stage:                     state.Stage,
		ActivatedAt:               state.ActivatedAt,
		CompletedAt:               state.CompletedAt,
		ScenarioID:                activeTutorialScenarioID(state),
		MonsterEncounterID:        activeTutorialMonsterEncounterID(state),
		Character:                 config.Character,
		Dialogue:                  append([]models.DialogueMessage{}, config.Dialogue...),
		PostWelcomeDialogue:       append([]models.DialogueMessage{}, config.PostWelcomeDialogue...),
		ScenarioObjectiveCopy:     config.ScenarioObjectiveCopy,
		PostScenarioDialogue:      append([]models.DialogueMessage{}, config.PostScenarioDialogue...),
		LoadoutDialogue:           append([]models.DialogueMessage{}, config.LoadoutDialogue...),
		LoadoutObjectiveCopy:      config.LoadoutObjectiveCopy,
		PostMonsterDialogue:       append([]models.DialogueMessage{}, config.PostMonsterDialogue...),
		BaseKitDialogue:           append([]models.DialogueMessage{}, config.BaseKitDialogue...),
		BaseKitObjectiveCopy:      config.BaseKitObjectiveCopy,
		PostBasePlacementDialogue: append([]models.DialogueMessage{}, config.PostBasePlacementDialogue...),
		HearthObjectiveCopy:       config.HearthObjectiveCopy,
		PostBaseDialogue:          append([]models.DialogueMessage{}, config.PostBaseDialogue...),
		MonsterObjectiveCopy:      config.MonsterObjectiveCopy,
		RequiredEquipItemIDs:      append([]int{}, state.RequiredEquipItemIDs...),
		CompletedEquipItemIDs:     append([]int{}, state.CompletedEquipItemIDs...),
		RequiredUseItemIDs:        append([]int{}, state.RequiredUseItemIDs...),
		CompletedUseItemIDs:       append([]int{}, state.CompletedUseItemIDs...),
	}
}

func parseTutorialConfigRequest(body tutorialConfigRequest) (*models.TutorialConfig, error) {
	config := &models.TutorialConfig{
		Dialogue:                  models.DialogueSequence{},
		PostWelcomeDialogue:       models.DialogueSequence{},
		ScenarioObjectiveCopy:     strings.TrimSpace(body.ScenarioObjectiveCopy),
		PostScenarioDialogue:      models.DialogueSequence{},
		LoadoutDialogue:           models.DialogueSequence{},
		LoadoutObjectiveCopy:      strings.TrimSpace(body.LoadoutObjectiveCopy),
		PostMonsterDialogue:       models.DialogueSequence{},
		BaseKitDialogue:           models.DialogueSequence{},
		BaseKitObjectiveCopy:      strings.TrimSpace(body.BaseKitObjectiveCopy),
		PostBasePlacementDialogue: models.DialogueSequence{},
		HearthObjectiveCopy:       strings.TrimSpace(body.HearthObjectiveCopy),
		PostBaseDialogue:          models.DialogueSequence{},
		ScenarioPrompt:            strings.TrimSpace(body.ScenarioPrompt),
		ScenarioImageURL:          strings.TrimSpace(body.ScenarioImageURL),
		Options:                   []models.TutorialScenarioOption{},
		MonsterObjectiveCopy:      strings.TrimSpace(body.MonsterObjectiveCopy),
		MonsterItemRewards:        []models.TutorialItemReward{},
		ItemRewards:               []models.TutorialItemReward{},
		SpellRewards:              []models.TutorialSpellReward{},
		MonsterRewardExperience:   body.MonsterRewardExperience,
		MonsterRewardGold:         body.MonsterRewardGold,
		RewardExperience:          body.RewardExperience,
		RewardGold:                body.RewardGold,
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

	if body.BaseQuestArchetypeID != nil {
		trimmed := strings.TrimSpace(*body.BaseQuestArchetypeID)
		if trimmed != "" {
			questArchetypeID, err := uuid.Parse(trimmed)
			if err != nil {
				return nil, fmt.Errorf("baseQuestArchetypeId must be a valid UUID")
			}
			config.BaseQuestArchetypeID = &questArchetypeID
		}
	}

	if body.BaseQuestGiverCharacterID != nil {
		trimmed := strings.TrimSpace(*body.BaseQuestGiverCharacterID)
		if trimmed != "" {
			characterID, err := uuid.Parse(trimmed)
			if err != nil {
				return nil, fmt.Errorf("baseQuestGiverCharacterId must be a valid UUID")
			}
			config.BaseQuestGiverCharacterID = &characterID
		}
	}
	if body.BaseQuestGiverCharacterTemplateID != nil {
		trimmed := strings.TrimSpace(*body.BaseQuestGiverCharacterTemplateID)
		if trimmed != "" {
			templateID, err := uuid.Parse(trimmed)
			if err != nil {
				return nil, fmt.Errorf("baseQuestGiverCharacterTemplateId must be a valid UUID")
			}
			config.BaseQuestGiverCharacterTemplateID = &templateID
		}
	}
	if config.BaseQuestGiverCharacterID != nil && config.BaseQuestGiverCharacterTemplateID != nil {
		return nil, fmt.Errorf("tutorial home base quest must use either baseQuestGiverCharacterId or baseQuestGiverCharacterTemplateId")
	}

	config.Dialogue = models.DialogueSequence(body.Dialogue)
	config.PostWelcomeDialogue = models.DialogueSequence(body.PostWelcomeDialogue)
	config.PostScenarioDialogue = models.DialogueSequence(body.PostScenarioDialogue)
	config.LoadoutDialogue = models.DialogueSequence(body.LoadoutDialogue)
	config.PostMonsterDialogue = models.DialogueSequence(body.PostMonsterDialogue)
	config.BaseKitDialogue = models.DialogueSequence(body.BaseKitDialogue)
	config.PostBasePlacementDialogue = models.DialogueSequence(body.PostBasePlacementDialogue)
	config.PostBaseDialogue = models.DialogueSequence(body.PostBaseDialogue)

	if body.MonsterEncounterID != nil {
		trimmed := strings.TrimSpace(*body.MonsterEncounterID)
		if trimmed != "" {
			monsterEncounterID, err := uuid.Parse(trimmed)
			if err != nil {
				return nil, fmt.Errorf("monsterEncounterId must be a valid UUID")
			}
			config.MonsterEncounterID = &monsterEncounterID
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
	if config.MonsterRewardExperience < 0 {
		config.MonsterRewardExperience = 0
	}
	if config.MonsterRewardGold < 0 {
		config.MonsterRewardGold = 0
	}

	for _, reward := range body.MonsterItemRewards {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		config.MonsterItemRewards = append(config.MonsterItemRewards, models.TutorialItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
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

func (s *server) maybeAdvanceTutorialProgress(
	ctx *gin.Context,
	userID uuid.UUID,
	config *models.TutorialConfig,
	state *models.UserTutorialState,
) (*models.UserTutorialState, error) {
	if state == nil {
		return nil, nil
	}
	switch state.Stage {
	case models.TutorialStageLoadout:
		if state.HasOutstandingLoadoutRequirements() {
			return state, nil
		}
		if config != nil && len(config.PostScenarioDialogue) > 0 {
			if err := s.dbClient.Tutorial().AdvanceToPostScenarioDialogue(ctx, userID); err != nil {
				return nil, err
			}
			return s.dbClient.Tutorial().FindStateByUserID(ctx, userID)
		}
		return s.advanceTutorialToMonsterOrComplete(ctx, userID, config)
	case models.TutorialStageBaseKit:
		if state.HasOutstandingLoadoutRequirements() {
			return state, nil
		}
		base, err := s.dbClient.Base().FindByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if base == nil {
			return state, nil
		}
		if err := s.dbClient.Tutorial().AdvanceToPostBaseDialogue(ctx, userID); err != nil {
			return nil, err
		}
		return s.dbClient.Tutorial().FindStateByUserID(ctx, userID)
	default:
		return state, nil
	}
}

func (s *server) advanceTutorialToMonsterOrComplete(
	ctx context.Context,
	userID uuid.UUID,
	config *models.TutorialConfig,
) (*models.UserTutorialState, error) {
	if config == nil || config.MonsterEncounterID == nil {
		if err := s.dbClient.Tutorial().MarkCompleted(ctx, userID); err != nil {
			return nil, err
		}
		return s.dbClient.Tutorial().FindStateByUserID(ctx, userID)
	}

	templateEncounter, err := s.dbClient.MonsterEncounter().FindByID(ctx, *config.MonsterEncounterID)
	if err != nil || templateEncounter == nil {
		if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
			err = s.dbClient.Tutorial().MarkCompleted(ctx, userID)
			if err != nil {
				return nil, err
			}
			return s.dbClient.Tutorial().FindStateByUserID(ctx, userID)
		}
		return nil, err
	}

	userLat, userLng, err := s.getUserLatLng(ctx, userID)
	if err != nil {
		return s.dbClient.Tutorial().FindStateByUserID(ctx, userID)
	}
	zones, err := s.dbClient.Zone().FindAll(ctx)
	if err != nil {
		return nil, err
	}
	zone, err := selectZoneForCoordinates(zones, userLat, userLng)
	if err != nil {
		return nil, err
	}

	encounter, members, monsters := buildTutorialMonsterEncounter(
		userID,
		zone.ID,
		userLat,
		userLng,
		templateEncounter,
		config,
	)
	if _, _, err := s.dbClient.Tutorial().ActivateMonsterForUser(
		ctx,
		userID,
		encounter,
		members,
		monsters,
	); err != nil {
		return nil, err
	}
	return s.dbClient.Tutorial().FindStateByUserID(ctx, userID)
}

func (s *server) forceResetTutorialReplayState(
	ctx *gin.Context,
	userID uuid.UUID,
	config *models.TutorialConfig,
) error {
	if config != nil && config.MonsterEncounterID != nil {
		if err := s.dbClient.UserMonsterEncounterVictory().Delete(
			ctx,
			userID,
			*config.MonsterEncounterID,
		); err != nil {
			return err
		}
	}
	if err := s.dbClient.Tutorial().InitializeForNewUser(ctx, userID); err != nil {
		return err
	}
	if err := s.dbClient.Tutorial().ResetForReplay(ctx, userID); err != nil {
		return err
	}
	return s.purgeTutorialReplayArtifacts(ctx, userID, config)
}

func (s *server) purgeTutorialReplayArtifacts(
	ctx context.Context,
	userID uuid.UUID,
	config *models.TutorialConfig,
) error {
	var baseQuestArchetypeID *uuid.UUID
	if config != nil {
		baseQuestArchetypeID = config.BaseQuestArchetypeID
	}
	return dungeonmasterruntime.PurgeTutorialReplayArtifacts(
		ctx,
		s.dbClient,
		userID,
		baseQuestArchetypeID,
	)
}

func (s *server) maybeAdvanceTutorialAfterHomeBaseKitUse(
	ctx *gin.Context,
	userID uuid.UUID,
) *tutorialStatusResponse {
	config, err := s.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		log.Printf("[tutorial] failed to load config after home base use user=%s err=%v", userID, err)
		return nil
	}
	state, err := s.dbClient.Tutorial().FindStateByUserID(ctx, userID)
	if err != nil {
		log.Printf("[tutorial] failed to load state after home base use user=%s err=%v", userID, err)
		return nil
	}
	if state == nil {
		return nil
	}
	nextState, err := s.maybeAdvanceTutorialProgress(ctx, userID, config, state)
	if err != nil {
		log.Printf("[tutorial] failed to advance after home base use user=%s err=%v", userID, err)
		nextState = state
	}
	response := buildTutorialStatusResponse(config, nextState)
	return &response
}

func (s *server) instantiateTutorialBaseQuestAsync(
	userID uuid.UUID,
	base *models.Base,
	config *models.TutorialConfig,
) {
	if base == nil || config == nil || config.BaseQuestArchetypeID == nil ||
		(config.BaseQuestGiverCharacterID == nil && config.BaseQuestGiverCharacterTemplateID == nil) {
		return
	}

	baseLatitude := base.Latitude
	baseLongitude := base.Longitude
	questArchetypeID := *config.BaseQuestArchetypeID
	questGiverCharacterID := cloneOptionalTutorialUUID(config.BaseQuestGiverCharacterID)
	questGiverCharacterTemplateID := cloneOptionalTutorialUUID(config.BaseQuestGiverCharacterTemplateID)

	fallback := func() {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()

			asyncBase := &models.Base{
				Latitude:  baseLatitude,
				Longitude: baseLongitude,
			}
			asyncConfig := &models.TutorialConfig{
				BaseQuestArchetypeID:              &questArchetypeID,
				BaseQuestGiverCharacterID:         questGiverCharacterID,
				BaseQuestGiverCharacterTemplateID: questGiverCharacterTemplateID,
			}

			if err := s.instantiateTutorialBaseQuest(
				bgCtx,
				userID,
				asyncBase,
				asyncConfig,
			); err != nil {
				log.Printf(
					"[tutorial] failed to instantiate tutorial base quest asynchronously user=%s err=%v",
					userID,
					err,
				)
			}
		}()
	}

	if s.asyncClient == nil {
		fallback()
		return
	}

	payloadBytes, err := json.Marshal(jobs.InstantiateTutorialBaseQuestTaskPayload{
		UserID:                            userID,
		BaseLatitude:                      baseLatitude,
		BaseLongitude:                     baseLongitude,
		BaseQuestArchetypeID:              questArchetypeID,
		BaseQuestGiverCharacterID:         questGiverCharacterID,
		BaseQuestGiverCharacterTemplateID: questGiverCharacterTemplateID,
	})
	if err != nil {
		log.Printf("[tutorial] failed to marshal tutorial base quest task payload user=%s err=%v", userID, err)
		fallback()
		return
	}

	if _, err := s.asyncClient.Enqueue(
		asynq.NewTask(jobs.InstantiateTutorialBaseQuestTaskType, payloadBytes),
	); err != nil {
		log.Printf("[tutorial] failed to enqueue tutorial base quest task user=%s err=%v", userID, err)
		fallback()
	}
}

func activeTutorialScenarioID(state *models.UserTutorialState) *uuid.UUID {
	if state == nil || state.Stage != models.TutorialStageScenario {
		return nil
	}
	return state.TutorialScenarioID
}

func cloneOptionalTutorialUUID(input *uuid.UUID) *uuid.UUID {
	if input == nil {
		return nil
	}
	value := *input
	return &value
}

func activeTutorialMonsterEncounterID(state *models.UserTutorialState) *uuid.UUID {
	if state == nil || state.Stage != models.TutorialStageMonster {
		return nil
	}
	return state.TutorialMonsterEncounterID
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

func buildTutorialMonsterEncounter(
	userID uuid.UUID,
	zoneID uuid.UUID,
	latitude float64,
	longitude float64,
	template *models.MonsterEncounter,
	config *models.TutorialConfig,
) (*models.MonsterEncounter, []models.MonsterEncounterMember, []models.Monster) {
	ownerUserID := userID
	encounter := &models.MonsterEncounter{
		Name:               strings.TrimSpace(template.Name),
		Description:        strings.TrimSpace(template.Description),
		ImageURL:           strings.TrimSpace(template.ImageURL),
		ThumbnailURL:       strings.TrimSpace(template.ThumbnailURL),
		EncounterType:      template.EncounterType,
		OwnerUserID:        &ownerUserID,
		Ephemeral:          true,
		ScaleWithUserLevel: template.ScaleWithUserLevel,
		ZoneID:             zoneID,
		Latitude:           latitude,
		Longitude:          longitude,
		RewardMode:         template.RewardMode,
		RandomRewardSize:   template.RandomRewardSize,
		RewardExperience:   maxInt(0, template.RewardExperience),
		RewardGold:         maxInt(0, template.RewardGold),
		MaterialRewards:    cloneBaseMaterialRewards(template.MaterialRewards),
		ItemRewards:        cloneMonsterEncounterRewardItems(template.ItemRewards),
	}

	useConfiguredRewards := tutorialMonsterRewardsConfigured(config)
	if useConfiguredRewards {
		encounter.RewardMode = models.RewardModeExplicit
		encounter.RandomRewardSize = models.RandomRewardSizeSmall
		encounter.RewardExperience = maxInt(0, config.MonsterRewardExperience)
		encounter.RewardGold = maxInt(0, config.MonsterRewardGold)
		encounter.ItemRewards = tutorialMonsterEncounterItemRewards(config.MonsterItemRewards)
	}
	members := make([]models.MonsterEncounterMember, 0, len(template.Members))
	monsters := make([]models.Monster, 0, len(template.Members))
	for index, member := range template.Members {
		source := member.Monster
		monster := models.Monster{
			Name:                        strings.TrimSpace(source.Name),
			Description:                 strings.TrimSpace(source.Description),
			ImageURL:                    strings.TrimSpace(source.ImageURL),
			ThumbnailURL:                strings.TrimSpace(source.ThumbnailURL),
			OwnerUserID:                 &ownerUserID,
			Ephemeral:                   true,
			ZoneID:                      zoneID,
			GenreID:                     source.GenreID,
			Genre:                       source.Genre,
			Latitude:                    latitude,
			Longitude:                   longitude,
			TemplateID:                  source.TemplateID,
			DominantHandInventoryItemID: source.DominantHandInventoryItemID,
			OffHandInventoryItemID:      source.OffHandInventoryItemID,
			WeaponInventoryItemID:       source.WeaponInventoryItemID,
			Level:                       source.Level,
			RewardMode:                  models.RewardModeExplicit,
			RandomRewardSize:            models.RandomRewardSizeSmall,
			RewardExperience:            source.RewardExperience,
			RewardGold:                  source.RewardGold,
			ItemRewards:                 cloneMonsterItemRewards(source.ItemRewards),
		}
		if useConfiguredRewards {
			if index == 0 {
				monster.RewardExperience = maxInt(0, config.MonsterRewardExperience)
				monster.RewardGold = maxInt(0, config.MonsterRewardGold)
				monster.ItemRewards = tutorialMonsterItemRewards(config.MonsterItemRewards)
			} else {
				monster.RewardExperience = 0
				monster.RewardGold = 0
				monster.ItemRewards = []models.MonsterItemReward{}
			}
		}
		monsters = append(monsters, monster)
		members = append(members, models.MonsterEncounterMember{Slot: member.Slot})
	}

	return encounter, members, monsters
}

func tutorialMonsterRewardsConfigured(config *models.TutorialConfig) bool {
	if config == nil {
		return false
	}
	return config.MonsterRewardExperience > 0 ||
		config.MonsterRewardGold > 0 ||
		len(config.MonsterItemRewards) > 0
}

func tutorialMonsterItemRewards(input []models.TutorialItemReward) []models.MonsterItemReward {
	rewards := make([]models.MonsterItemReward, 0, len(input))
	for _, reward := range input {
		rewards = append(rewards, models.MonsterItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func tutorialMonsterEncounterItemRewards(input []models.TutorialItemReward) models.MonsterEncounterRewardItems {
	rewards := make(models.MonsterEncounterRewardItems, 0, len(input))
	for _, reward := range input {
		rewards = append(rewards, models.MonsterEncounterRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func (s *server) tutorialBaseKitRequirementItemIDs(
	ctx *gin.Context,
	config *models.TutorialConfig,
) ([]int, error) {
	if config == nil || config.MonsterEncounterID == nil {
		return []int{}, nil
	}

	candidateRewards := []models.TutorialItemReward{}
	if len(config.MonsterItemRewards) > 0 {
		candidateRewards = append(candidateRewards, config.MonsterItemRewards...)
	} else {
		templateEncounter, err := s.dbClient.MonsterEncounter().FindByID(ctx, *config.MonsterEncounterID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return []int{}, nil
			}
			return nil, err
		}
		if templateEncounter == nil {
			return []int{}, nil
		}
		for _, member := range templateEncounter.Members {
			if member.MonsterID == uuid.Nil {
				continue
			}
			for _, reward := range member.Monster.ItemRewards {
				candidateRewards = append(candidateRewards, models.TutorialItemReward{
					InventoryItemID: reward.InventoryItemID,
					Quantity:        reward.Quantity,
				})
			}
		}
	}

	requiredUseItemIDs := []int{}
	seen := map[int]struct{}{}
	for _, reward := range candidateRewards {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, reward.InventoryItemID)
		if err != nil {
			return nil, err
		}
		if item == nil || !item.ConsumeCreateBase {
			continue
		}
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		requiredUseItemIDs = append(requiredUseItemIDs, item.ID)
	}
	return requiredUseItemIDs, nil
}

func (s *server) instantiateTutorialBaseQuest(
	ctx context.Context,
	userID uuid.UUID,
	base *models.Base,
	config *models.TutorialConfig,
) error {
	if base == nil || config == nil || config.BaseQuestArchetypeID == nil ||
		(config.BaseQuestGiverCharacterID == nil && config.BaseQuestGiverCharacterTemplateID == nil) {
		return nil
	}
	return dungeonmasterruntime.InstantiateTutorialBaseQuest(
		ctx,
		s.dbClient,
		s.dungeonmaster,
		userID,
		base,
		*config.BaseQuestArchetypeID,
		config.BaseQuestGiverCharacterID,
		config.BaseQuestGiverCharacterTemplateID,
	)
}

func (s *server) findTutorialBaseQuestForUser(
	ctx context.Context,
	userID uuid.UUID,
	questArchetypeID uuid.UUID,
) (*models.Quest, error) {
	quests, err := s.dbClient.Quest().FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range quests {
		quest := &quests[i]
		if quest.OwnerUserID == nil || *quest.OwnerUserID != userID {
			continue
		}
		if quest.QuestArchetypeID == nil || *quest.QuestArchetypeID != questArchetypeID {
			continue
		}
		return quest, nil
	}
	return nil, nil
}

func cloneMonsterItemRewards(input []models.MonsterItemReward) []models.MonsterItemReward {
	rewards := make([]models.MonsterItemReward, 0, len(input))
	for _, reward := range input {
		rewards = append(rewards, models.MonsterItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func cloneMonsterEncounterRewardItems(input models.MonsterEncounterRewardItems) models.MonsterEncounterRewardItems {
	rewards := make(models.MonsterEncounterRewardItems, 0, len(input))
	for _, reward := range input {
		rewards = append(rewards, models.MonsterEncounterRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func cloneBaseMaterialRewards(input models.BaseMaterialRewards) models.BaseMaterialRewards {
	rewards := make(models.BaseMaterialRewards, 0, len(input))
	for _, reward := range input {
		rewards = append(rewards, models.BaseResourceDelta{
			ResourceKey: reward.ResourceKey,
			Amount:      reward.Amount,
		})
	}
	return rewards
}

func (s *server) tutorialLoadoutRequirementItemIDs(
	ctx *gin.Context,
	rewards []scenarioRewardItem,
) ([]int, []int, error) {
	requiredEquipItemIDs := []int{}
	requiredUseItemIDs := []int{}
	seenEquip := map[int]struct{}{}
	seenUse := map[int]struct{}{}

	for _, reward := range rewards {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, reward.InventoryItemID)
		if err != nil {
			return nil, nil, err
		}
		if item == nil {
			continue
		}
		equipSlot := ""
		if item.EquipSlot != nil {
			equipSlot = strings.TrimSpace(*item.EquipSlot)
		}
		if equipSlot != "" {
			if _, ok := seenEquip[item.ID]; !ok {
				seenEquip[item.ID] = struct{}{}
				requiredEquipItemIDs = append(requiredEquipItemIDs, item.ID)
			}
		}
		if inventoryItemHasConsumableEffects(item) {
			if _, ok := seenUse[item.ID]; !ok {
				seenUse[item.ID] = struct{}{}
				requiredUseItemIDs = append(requiredUseItemIDs, item.ID)
			}
		}
	}

	return requiredEquipItemIDs, requiredUseItemIDs, nil
}

func cloneScenarioRewardItems(input []scenarioRewardItem) []scenarioRewardItem {
	out := make([]scenarioRewardItem, 0, len(input))
	for _, reward := range input {
		out = append(out, scenarioRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func cloneScenarioRewardSpells(input []scenarioRewardSpell) []scenarioRewardSpell {
	out := make([]scenarioRewardSpell, 0, len(input))
	for _, reward := range input {
		out = append(out, scenarioRewardSpell{SpellID: reward.SpellID})
	}
	return out
}

func scenarioVisibleToUser(userID uuid.UUID, scenario *models.Scenario) bool {
	if scenario == nil || scenario.OwnerUserID == nil {
		return true
	}
	return *scenario.OwnerUserID == userID
}

func characterVisibleToUser(userID uuid.UUID, character *models.Character) bool {
	if character == nil || character.OwnerUserID == nil {
		return true
	}
	return *character.OwnerUserID == userID
}

func monsterVisibleToUser(userID uuid.UUID, monster *models.Monster) bool {
	if monster == nil || monster.OwnerUserID == nil {
		return true
	}
	return *monster.OwnerUserID == userID
}

func questVisibleToUser(userID uuid.UUID, quest *models.Quest) bool {
	if quest == nil || quest.OwnerUserID == nil {
		return true
	}
	return *quest.OwnerUserID == userID
}

func monsterEncounterVisibleToUser(userID uuid.UUID, encounter *models.MonsterEncounter) bool {
	if encounter == nil || encounter.OwnerUserID == nil {
		return true
	}
	return *encounter.OwnerUserID == userID
}
