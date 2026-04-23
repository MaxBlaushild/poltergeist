package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type paginatedScenarioTemplateResponse struct {
	Items    []models.ScenarioTemplate `json:"items"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"pageSize"`
}

type scenarioTemplateUpsertRequest struct {
	GenreID                   string                         `json:"genreId"`
	ZoneKind                  string                         `json:"zoneKind"`
	Prompt                    string                         `json:"prompt"`
	ImageURL                  string                         `json:"imageUrl"`
	ThumbnailURL              string                         `json:"thumbnailUrl"`
	RewardMode                string                         `json:"rewardMode"`
	RandomRewardSize          string                         `json:"randomRewardSize"`
	Difficulty                *int                           `json:"difficulty"`
	RewardExperience          int                            `json:"rewardExperience"`
	RewardGold                int                            `json:"rewardGold"`
	OpenEnded                 bool                           `json:"openEnded"`
	SuccessHandoffText        string                         `json:"successHandoffText"`
	FailureHandoffText        string                         `json:"failureHandoffText"`
	ScaleWithUserLevel        bool                           `json:"scaleWithUserLevel"`
	FailurePenaltyMode        string                         `json:"failurePenaltyMode"`
	FailureHealthDrainType    string                         `json:"failureHealthDrainType"`
	FailureHealthDrainValue   int                            `json:"failureHealthDrainValue"`
	FailureManaDrainType      string                         `json:"failureManaDrainType"`
	FailureManaDrainValue     int                            `json:"failureManaDrainValue"`
	FailureStatuses           []scenarioFailureStatusPayload `json:"failureStatuses"`
	SuccessRewardMode         string                         `json:"successRewardMode"`
	SuccessHealthRestoreType  string                         `json:"successHealthRestoreType"`
	SuccessHealthRestoreValue int                            `json:"successHealthRestoreValue"`
	SuccessManaRestoreType    string                         `json:"successManaRestoreType"`
	SuccessManaRestoreValue   int                            `json:"successManaRestoreValue"`
	SuccessStatuses           []scenarioFailureStatusPayload `json:"successStatuses"`
	Options                   []scenarioOptionPayload        `json:"options"`
	ItemRewards               []scenarioRewardItemPayload    `json:"itemRewards"`
	ItemChoiceRewards         []scenarioRewardItemPayload    `json:"itemChoiceRewards"`
	SpellRewards              []scenarioRewardSpellPayload   `json:"spellRewards"`
}

type scenarioTemplateGenerationJobRequest struct {
	Count     int    `json:"count"`
	OpenEnded bool   `json:"openEnded"`
	GenreID   string `json:"genreId"`
	ZoneKind  string `json:"zoneKind"`
	YeetIt    bool   `json:"yeetIt"`
}

func (s *server) parseScenarioTemplateUpsertRequest(
	ctx context.Context,
	body scenarioTemplateUpsertRequest,
	existing *models.ScenarioTemplate,
) (*models.ScenarioTemplate, error) {
	prompt := strings.TrimSpace(body.Prompt)
	if len(prompt) > 1200 {
		prompt = strings.TrimSpace(prompt[:1200])
	}
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	thumbnailURL := strings.TrimSpace(body.ThumbnailURL)
	if thumbnailURL == "" && strings.TrimSpace(body.ImageURL) != "" {
		thumbnailURL = strings.TrimSpace(body.ImageURL)
	}

	difficulty, err := scenarioDifficultyValue(body.Difficulty)
	if err != nil {
		return nil, err
	}
	if body.RewardExperience < 0 {
		return nil, fmt.Errorf("rewardExperience must be zero or greater")
	}
	if body.RewardGold < 0 {
		return nil, fmt.Errorf("rewardGold must be zero or greater")
	}

	failurePenaltyMode, err := normalizeScenarioFailurePenaltyMode(body.FailurePenaltyMode, body.OpenEnded)
	if err != nil {
		return nil, err
	}
	failureHealthDrainType, err := normalizeScenarioFailureDrainType(body.FailureHealthDrainType)
	if err != nil {
		return nil, fmt.Errorf("invalid failureHealthDrainType")
	}
	failureHealthDrainValue, err := normalizeScenarioFailureDrainValue(
		failureHealthDrainType,
		body.FailureHealthDrainValue,
		"failureHealthDrainValue",
	)
	if err != nil {
		return nil, err
	}
	failureManaDrainType, err := normalizeScenarioFailureDrainType(body.FailureManaDrainType)
	if err != nil {
		return nil, fmt.Errorf("invalid failureManaDrainType")
	}
	failureManaDrainValue, err := normalizeScenarioFailureDrainValue(
		failureManaDrainType,
		body.FailureManaDrainValue,
		"failureManaDrainValue",
	)
	if err != nil {
		return nil, err
	}
	failureStatuses, err := parseScenarioFailureStatusTemplates(body.FailureStatuses, "failureStatuses")
	if err != nil {
		return nil, err
	}

	successRewardMode, err := normalizeScenarioSuccessRewardMode(body.SuccessRewardMode, body.OpenEnded)
	if err != nil {
		return nil, err
	}
	successHealthRestoreType, err := normalizeScenarioFailureDrainType(body.SuccessHealthRestoreType)
	if err != nil {
		return nil, fmt.Errorf("invalid successHealthRestoreType")
	}
	successHealthRestoreValue, err := normalizeScenarioFailureDrainValue(
		successHealthRestoreType,
		body.SuccessHealthRestoreValue,
		"successHealthRestoreValue",
	)
	if err != nil {
		return nil, err
	}
	successManaRestoreType, err := normalizeScenarioFailureDrainType(body.SuccessManaRestoreType)
	if err != nil {
		return nil, fmt.Errorf("invalid successManaRestoreType")
	}
	successManaRestoreValue, err := normalizeScenarioFailureDrainValue(
		successManaRestoreType,
		body.SuccessManaRestoreValue,
		"successManaRestoreValue",
	)
	if err != nil {
		return nil, err
	}
	successStatuses, err := parseScenarioFailureStatusTemplates(body.SuccessStatuses, "successStatuses")
	if err != nil {
		return nil, err
	}

	if body.OpenEnded && len(body.Options) > 0 {
		return nil, fmt.Errorf("openEnded templates cannot include options")
	}
	if !body.OpenEnded && len(body.Options) == 0 {
		return nil, fmt.Errorf("choice-based templates require at least one option")
	}

	options := make(models.ScenarioTemplateOptions, 0, len(body.Options))
	for _, optionPayload := range body.Options {
		optionText := strings.TrimSpace(optionPayload.OptionText)
		successText := strings.TrimSpace(optionPayload.SuccessText)
		failureText := strings.TrimSpace(optionPayload.FailureText)
		if optionText == "" || successText == "" || failureText == "" {
			return nil, fmt.Errorf("each option requires optionText, successText, and failureText")
		}
		statTag, ok := normalizeScenarioStatTag(optionPayload.StatTag)
		if !ok {
			return nil, fmt.Errorf("invalid option statTag")
		}
		optionFailureHealthDrainType, err := normalizeScenarioFailureDrainType(optionPayload.FailureHealthDrainType)
		if err != nil {
			return nil, fmt.Errorf("invalid option failureHealthDrainType")
		}
		optionFailureHealthDrainValue, err := normalizeScenarioFailureDrainValue(optionFailureHealthDrainType, optionPayload.FailureHealthDrainValue, "option failureHealthDrainValue")
		if err != nil {
			return nil, err
		}
		optionFailureManaDrainType, err := normalizeScenarioFailureDrainType(optionPayload.FailureManaDrainType)
		if err != nil {
			return nil, fmt.Errorf("invalid option failureManaDrainType")
		}
		optionFailureManaDrainValue, err := normalizeScenarioFailureDrainValue(optionFailureManaDrainType, optionPayload.FailureManaDrainValue, "option failureManaDrainValue")
		if err != nil {
			return nil, err
		}
		optionFailureStatuses, err := parseScenarioFailureStatusTemplates(optionPayload.FailureStatuses, "option failureStatuses")
		if err != nil {
			return nil, err
		}
		optionSuccessHealthRestoreType, err := normalizeScenarioFailureDrainType(optionPayload.SuccessHealthRestoreType)
		if err != nil {
			return nil, fmt.Errorf("invalid option successHealthRestoreType")
		}
		optionSuccessHealthRestoreValue, err := normalizeScenarioFailureDrainValue(optionSuccessHealthRestoreType, optionPayload.SuccessHealthRestoreValue, "option successHealthRestoreValue")
		if err != nil {
			return nil, err
		}
		optionSuccessManaRestoreType, err := normalizeScenarioFailureDrainType(optionPayload.SuccessManaRestoreType)
		if err != nil {
			return nil, fmt.Errorf("invalid option successManaRestoreType")
		}
		optionSuccessManaRestoreValue, err := normalizeScenarioFailureDrainValue(optionSuccessManaRestoreType, optionPayload.SuccessManaRestoreValue, "option successManaRestoreValue")
		if err != nil {
			return nil, err
		}
		optionSuccessStatuses, err := parseScenarioFailureStatusTemplates(optionPayload.SuccessStatuses, "option successStatuses")
		if err != nil {
			return nil, err
		}
		var optionDifficulty *int
		if optionPayload.Difficulty != nil {
			if *optionPayload.Difficulty < 0 {
				return nil, fmt.Errorf("option difficulty must be zero or greater")
			}
			value := *optionPayload.Difficulty
			optionDifficulty = &value
		}
		options = append(options, models.ScenarioTemplateOption{
			OptionText:                optionText,
			SuccessText:               successText,
			FailureText:               failureText,
			SuccessHandoffText:        strings.TrimSpace(optionPayload.SuccessHandoffText),
			FailureHandoffText:        strings.TrimSpace(optionPayload.FailureHandoffText),
			StatTag:                   statTag,
			Proficiencies:             models.StringArray(normalizeScenarioProficiencies(optionPayload.Proficiencies)),
			Difficulty:                optionDifficulty,
			RewardExperience:          optionPayload.RewardExperience,
			RewardGold:                optionPayload.RewardGold,
			FailureHealthDrainType:    optionFailureHealthDrainType,
			FailureHealthDrainValue:   optionFailureHealthDrainValue,
			FailureManaDrainType:      optionFailureManaDrainType,
			FailureManaDrainValue:     optionFailureManaDrainValue,
			FailureStatuses:           optionFailureStatuses,
			SuccessHealthRestoreType:  optionSuccessHealthRestoreType,
			SuccessHealthRestoreValue: optionSuccessHealthRestoreValue,
			SuccessManaRestoreType:    optionSuccessManaRestoreType,
			SuccessManaRestoreValue:   optionSuccessManaRestoreValue,
			SuccessStatuses:           optionSuccessStatuses,
			ItemRewards:               scenarioRewardPayloadsToTemplateRewards(optionPayload.ItemRewards),
			ItemChoiceRewards:         scenarioRewardPayloadsToTemplateRewards(optionPayload.ItemChoiceRewards),
			SpellRewards:              scenarioSpellPayloadsToTemplateRewards(optionPayload.SpellRewards),
		})
	}

	itemRewards := scenarioRewardPayloadsToTemplateRewards(body.ItemRewards)
	itemChoiceRewards := scenarioRewardPayloadsToTemplateRewards(body.ItemChoiceRewards)
	if len(itemChoiceRewards) == 1 {
		return nil, fmt.Errorf("itemChoiceRewards must include at least 2 options when provided")
	}
	spellRewards := scenarioSpellPayloadsToTemplateRewards(body.SpellRewards)

	rawGenreID := strings.TrimSpace(body.GenreID)
	if rawGenreID == "" && existing != nil && existing.GenreID != uuid.Nil {
		rawGenreID = existing.GenreID.String()
	}
	genre, err := s.resolveZoneGenre(ctx, rawGenreID)
	if err != nil {
		return nil, err
	}
	zoneKind, err := s.resolveOptionalZoneKind(ctx, body.ZoneKind)
	if err != nil {
		return nil, err
	}

	rewardMode := models.NormalizeRewardMode(body.RewardMode)
	if strings.TrimSpace(body.RewardMode) == "" {
		if body.RewardExperience > 0 || body.RewardGold > 0 || len(itemRewards) > 0 || len(itemChoiceRewards) > 0 || len(spellRewards) > 0 || scenarioOptionsHaveExplicitRewards(body.Options) {
			rewardMode = models.RewardModeExplicit
		}
	}

	template := &models.ScenarioTemplate{
		GenreID:                   genre.ID,
		Genre:                     genre,
		ZoneKind:                  models.NormalizeZoneKind(body.ZoneKind),
		Prompt:                    prompt,
		ImageURL:                  strings.TrimSpace(body.ImageURL),
		ThumbnailURL:              thumbnailURL,
		ScaleWithUserLevel:        body.ScaleWithUserLevel,
		RewardMode:                rewardMode,
		RandomRewardSize:          models.NormalizeRandomRewardSize(body.RandomRewardSize),
		Difficulty:                difficulty,
		RewardExperience:          body.RewardExperience,
		RewardGold:                body.RewardGold,
		OpenEnded:                 body.OpenEnded,
		SuccessHandoffText:        strings.TrimSpace(body.SuccessHandoffText),
		FailureHandoffText:        strings.TrimSpace(body.FailureHandoffText),
		FailurePenaltyMode:        failurePenaltyMode,
		FailureHealthDrainType:    failureHealthDrainType,
		FailureHealthDrainValue:   failureHealthDrainValue,
		FailureManaDrainType:      failureManaDrainType,
		FailureManaDrainValue:     failureManaDrainValue,
		FailureStatuses:           failureStatuses,
		SuccessRewardMode:         successRewardMode,
		SuccessHealthRestoreType:  successHealthRestoreType,
		SuccessHealthRestoreValue: successHealthRestoreValue,
		SuccessManaRestoreType:    successManaRestoreType,
		SuccessManaRestoreValue:   successManaRestoreValue,
		SuccessStatuses:           successStatuses,
		Options:                   options,
		ItemRewards:               itemRewards,
		ItemChoiceRewards:         itemChoiceRewards,
		SpellRewards:              spellRewards,
	}
	if zoneKind != nil {
		template.ZoneKind = zoneKind.Slug
	}
	return template, nil
}

func scenarioRewardPayloadsToTemplateRewards(payloads []scenarioRewardItemPayload) models.ScenarioTemplateRewards {
	rewards := make(models.ScenarioTemplateRewards, 0, len(payloads))
	for _, reward := range payloads {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		rewards = append(rewards, models.ScenarioTemplateReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func scenarioSpellPayloadsToTemplateRewards(payloads []scenarioRewardSpellPayload) models.ScenarioTemplateSpellRewards {
	rewards := make(models.ScenarioTemplateSpellRewards, 0, len(payloads))
	for _, reward := range payloads {
		spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
		if err != nil || spellID == uuid.Nil {
			continue
		}
		rewards = append(rewards, models.ScenarioTemplateSpellReward{SpellID: spellID})
	}
	return rewards
}

func scenarioTemplateFromGenerationDraft(
	draft *models.ScenarioTemplateGenerationDraft,
) *models.ScenarioTemplate {
	if draft == nil {
		return nil
	}
	template := models.ScenarioTemplateFromGenerationDraftPayload(draft.Payload)
	if draft.GenreID != uuid.Nil {
		template.GenreID = draft.GenreID
		template.Genre = draft.Genre
	}
	if zoneKind := models.NormalizeZoneKind(draft.ZoneKind); zoneKind != "" {
		template.ZoneKind = zoneKind
	}
	if prompt := strings.TrimSpace(draft.Prompt); prompt != "" {
		template.Prompt = prompt
	}
	template.OpenEnded = draft.OpenEnded
	template.Difficulty = draft.Difficulty
	return template
}

func (s *server) getScenarioTemplates(ctx *gin.Context) {
	templates, err := s.dbClient.ScenarioTemplate().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, templates)
}

func (s *server) getAdminScenarioTemplates(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	page := parseAdminMonsterListPage(ctx)
	pageSize := parseAdminMonsterListPageSize(ctx)
	genreID, err := parseOptionalGenreIDFilter(ctx.Query("genreId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := s.dbClient.ScenarioTemplate().ListAdmin(ctx, db.ScenarioTemplateAdminListParams{
		Page:     page,
		PageSize: pageSize,
		Query:    ctx.Query("query"),
		GenreID:  genreID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, paginatedScenarioTemplateResponse{
		Items:    result.Templates,
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *server) getScenarioTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario template ID"})
		return
	}
	template, err := s.dbClient.ScenarioTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if template == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario template not found"})
		return
	}
	ctx.JSON(http.StatusOK, template)
}

func (s *server) createScenarioTemplate(ctx *gin.Context) {
	var body scenarioTemplateUpsertRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := s.parseScenarioTemplateUpsertRequest(ctx, body, nil)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.ScenarioTemplate().Create(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	created, err := s.dbClient.ScenarioTemplate().FindByID(ctx, template.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateScenarioTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario template ID"})
		return
	}
	existing, err := s.dbClient.ScenarioTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario template not found"})
		return
	}
	var body scenarioTemplateUpsertRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := s.parseScenarioTemplateUpsertRequest(ctx, body, existing)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.ScenarioTemplate().Update(ctx, id, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := s.dbClient.ScenarioTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteScenarioTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario template ID"})
		return
	}
	if err := s.dbClient.ScenarioTemplate().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "scenario template deleted successfully"})
}

func (s *server) createScenarioTemplateGenerationJob(ctx *gin.Context) {
	var body scenarioTemplateGenerationJobRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Count <= 0 {
		body.Count = 1
	}
	if body.Count > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "count must be between 1 and 100"})
		return
	}
	genre, err := s.resolveZoneGenre(ctx, body.GenreID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	zoneKind, err := s.resolveOptionalZoneKind(ctx, body.ZoneKind)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := &models.ScenarioTemplateGenerationJob{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		GenreID:      genre.ID,
		Genre:        genre,
		ZoneKind:     models.NormalizeZoneKind(body.ZoneKind),
		YeetIt:       body.YeetIt,
		Status:       models.ScenarioTemplateGenerationStatusQueued,
		Count:        body.Count,
		OpenEnded:    body.OpenEnded,
		CreatedCount: 0,
	}
	if zoneKind != nil {
		job.ZoneKind = zoneKind.Slug
	}
	if err := s.dbClient.ScenarioTemplateGenerationJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.GenerateScenarioTemplatesTaskPayload{JobID: job.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateScenarioTemplatesTaskType, payload)); err != nil {
		errMsg := err.Error()
		job.Status = models.ScenarioTemplateGenerationStatusFailed
		job.ErrorMessage = &errMsg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.ScenarioTemplateGenerationJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, job)
}

func (s *server) getScenarioTemplateGenerationJobs(ctx *gin.Context) {
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	jobsList, err := s.dbClient.ScenarioTemplateGenerationJob().FindRecent(ctx, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getScenarioTemplateGenerationJob(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario template generation job ID"})
		return
	}
	job, err := s.dbClient.ScenarioTemplateGenerationJob().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario template generation job not found"})
		return
	}
	ctx.JSON(http.StatusOK, job)
}

func (s *server) getScenarioTemplateGenerationDrafts(ctx *gin.Context) {
	jobID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario template generation job ID"})
		return
	}
	drafts, err := s.dbClient.ScenarioTemplateGenerationDraft().FindByJobID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, drafts)
}

func (s *server) convertScenarioTemplateGenerationDraft(ctx *gin.Context) {
	draftID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario template generation draft ID"})
		return
	}
	draft, err := s.dbClient.ScenarioTemplateGenerationDraft().FindByID(ctx, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario template generation draft not found"})
		return
	}
	if draft.ScenarioTemplateID != nil && *draft.ScenarioTemplateID != uuid.Nil {
		existing, findErr := s.dbClient.ScenarioTemplate().FindByID(ctx, *draft.ScenarioTemplateID)
		if findErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": findErr.Error()})
			return
		}
		if existing != nil {
			ctx.JSON(http.StatusOK, existing)
			return
		}
	}

	template := scenarioTemplateFromGenerationDraft(draft)
	if err := s.dbClient.ScenarioTemplate().Create(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	convertedAt := time.Now().UTC()
	draft.Status = models.ScenarioTemplateGenerationDraftStatusConverted
	draft.ScenarioTemplateID = &template.ID
	draft.ConvertedAt = &convertedAt
	draft.UpdatedAt = convertedAt
	if err := s.dbClient.ScenarioTemplateGenerationDraft().Update(ctx, draft); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.dbClient.ScenarioTemplate().FindByID(ctx, template.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, created)
}

func (s *server) deleteScenarioTemplateGenerationDraft(ctx *gin.Context) {
	draftID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario template generation draft ID"})
		return
	}
	draft, err := s.dbClient.ScenarioTemplateGenerationDraft().FindByID(ctx, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario template generation draft not found"})
		return
	}
	if draft.ScenarioTemplateID != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "converted drafts cannot be deleted"})
		return
	}
	if err := s.dbClient.ScenarioTemplateGenerationDraft().Delete(ctx, draftID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "scenario template generation draft deleted"})
}
