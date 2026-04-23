package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type challengeTemplateUpsertRequest struct {
	LocationArchetypeID string                      `json:"locationArchetypeId"`
	Question            string                      `json:"question"`
	Description         string                      `json:"description"`
	ImageURL            string                      `json:"imageUrl"`
	ThumbnailURL        string                      `json:"thumbnailUrl"`
	ScaleWithUserLevel  bool                        `json:"scaleWithUserLevel"`
	RewardMode          string                      `json:"rewardMode"`
	RandomRewardSize    string                      `json:"randomRewardSize"`
	RewardExperience    int                         `json:"rewardExperience"`
	Reward              int                         `json:"reward"`
	InventoryItemID     *int                        `json:"inventoryItemId"`
	ItemChoiceRewards   []scenarioRewardItemPayload `json:"itemChoiceRewards"`
	SubmissionType      string                      `json:"submissionType"`
	Difficulty          int                         `json:"difficulty"`
	StatTags            []string                    `json:"statTags"`
	Proficiency         string                      `json:"proficiency"`
}

type challengeTemplateGenerationJobRequest struct {
	LocationArchetypeID string `json:"locationArchetypeId"`
	Count               int    `json:"count"`
	ZoneKind            string `json:"zoneKind"`
}

func (s *server) parseChallengeTemplateUpsertRequest(body challengeTemplateUpsertRequest) (*models.ChallengeTemplate, error) {
	locationArchetypeID, err := uuid.Parse(strings.TrimSpace(body.LocationArchetypeID))
	if err != nil {
		return nil, fmt.Errorf("invalid locationArchetypeId")
	}
	question := strings.TrimSpace(body.Question)
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}
	if body.RewardExperience < 0 {
		return nil, fmt.Errorf("rewardExperience must be zero or greater")
	}
	if body.Reward < 0 {
		return nil, fmt.Errorf("reward must be zero or greater")
	}
	if body.Difficulty < 0 {
		return nil, fmt.Errorf("difficulty must be zero or greater")
	}

	submissionTypeRaw := strings.TrimSpace(body.SubmissionType)
	if submissionTypeRaw == "" {
		submissionTypeRaw = string(models.DefaultQuestNodeSubmissionType())
	}
	submissionType := models.QuestNodeSubmissionType(submissionTypeRaw)
	if !submissionType.IsValid() {
		return nil, fmt.Errorf("invalid submissionType")
	}

	proficiency := strings.TrimSpace(body.Proficiency)
	var proficiencyPtr *string
	if proficiency != "" {
		proficiencyPtr = &proficiency
	}
	imageURL := strings.TrimSpace(body.ImageURL)
	thumbnailURL := strings.TrimSpace(body.ThumbnailURL)
	if thumbnailURL == "" && imageURL != "" {
		thumbnailURL = imageURL
	}

	itemChoiceRewards := make(models.ChallengeTemplateItemChoiceRewards, 0, len(body.ItemChoiceRewards))
	for _, reward := range body.ItemChoiceRewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			return nil, fmt.Errorf("itemChoiceRewards require inventoryItemId and positive quantity")
		}
		itemChoiceRewards = append(itemChoiceRewards, models.ChallengeTemplateItemChoiceReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	if len(itemChoiceRewards) == 1 {
		return nil, fmt.Errorf("itemChoiceRewards must include at least 2 options when provided")
	}
	rewardMode := models.NormalizeRewardMode(body.RewardMode)
	if strings.TrimSpace(body.RewardMode) == "" {
		if body.RewardExperience > 0 || body.Reward > 0 || body.InventoryItemID != nil || len(itemChoiceRewards) > 0 {
			rewardMode = models.RewardModeExplicit
		}
	}
	if rewardMode == models.RewardModeRandom && len(itemChoiceRewards) > 0 {
		return nil, fmt.Errorf("itemChoiceRewards require explicit rewardMode")
	}

	return &models.ChallengeTemplate{
		LocationArchetypeID: locationArchetypeID,
		Question:            question,
		Description:         strings.TrimSpace(body.Description),
		ImageURL:            imageURL,
		ThumbnailURL:        thumbnailURL,
		ScaleWithUserLevel:  body.ScaleWithUserLevel,
		RewardMode:          rewardMode,
		RandomRewardSize:    models.NormalizeRandomRewardSize(body.RandomRewardSize),
		RewardExperience:    body.RewardExperience,
		Reward:              body.Reward,
		InventoryItemID:     body.InventoryItemID,
		ItemChoiceRewards:   itemChoiceRewards,
		SubmissionType:      submissionType,
		Difficulty:          body.Difficulty,
		StatTags:            parseChallengeStatTags(body.StatTags),
		Proficiency:         proficiencyPtr,
	}, nil
}

func (s *server) getChallengeTemplates(ctx *gin.Context) {
	templates, err := s.dbClient.ChallengeTemplate().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, templates)
}

func (s *server) getChallengeTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge template ID"})
		return
	}
	template, err := s.dbClient.ChallengeTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if template == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge template not found"})
		return
	}
	ctx.JSON(http.StatusOK, template)
}

func (s *server) createChallengeTemplate(ctx *gin.Context) {
	var body challengeTemplateUpsertRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := s.parseChallengeTemplateUpsertRequest(body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.ChallengeTemplate().Create(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	created, err := s.dbClient.ChallengeTemplate().FindByID(ctx, template.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateChallengeTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge template ID"})
		return
	}
	existing, err := s.dbClient.ChallengeTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge template not found"})
		return
	}
	var body challengeTemplateUpsertRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := s.parseChallengeTemplateUpsertRequest(body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.ChallengeTemplate().Update(ctx, id, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := s.dbClient.ChallengeTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteChallengeTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge template ID"})
		return
	}
	if err := s.dbClient.ChallengeTemplate().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "challenge template deleted successfully"})
}

func (s *server) generateChallengeTemplateImage(ctx *gin.Context) {
	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge template ID"})
		return
	}

	template, err := s.dbClient.ChallengeTemplate().FindByID(ctx, templateID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if template == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge template not found"})
		return
	}

	payload := jobs.GenerateChallengeTemplateImageTaskPayload{
		ChallengeTemplateID: templateID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateChallengeTemplateImageTaskType, payloadBytes)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"status":            "queued",
		"challengeTemplate": template,
	})
}

func (s *server) createChallengeTemplateGenerationJob(ctx *gin.Context) {
	var body challengeTemplateGenerationJobRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	locationArchetypeID, err := uuid.Parse(strings.TrimSpace(body.LocationArchetypeID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid locationArchetypeId"})
		return
	}
	if body.Count <= 0 {
		body.Count = 1
	}
	if body.Count > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "count must be between 1 and 100"})
		return
	}
	locationArchetype, err := s.dbClient.LocationArchetype().FindByID(ctx, locationArchetypeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if locationArchetype == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "location archetype not found"})
		return
	}
	zoneKind, err := s.resolveOptionalZoneKind(ctx, body.ZoneKind)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := &models.ChallengeTemplateGenerationJob{
		ID:                  uuid.New(),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		LocationArchetypeID: locationArchetypeID,
		ZoneKind:            models.NormalizeZoneKind(body.ZoneKind),
		Status:              models.ChallengeTemplateGenerationStatusQueued,
		Count:               body.Count,
		CreatedCount:        0,
	}
	if zoneKind != nil {
		job.ZoneKind = zoneKind.Slug
	}
	if err := s.dbClient.ChallengeTemplateGenerationJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	payload, err := json.Marshal(jobs.GenerateChallengeTemplatesTaskPayload{JobID: job.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateChallengeTemplatesTaskType, payload)); err != nil {
		errMsg := err.Error()
		job.Status = models.ChallengeTemplateGenerationStatusFailed
		job.ErrorMessage = &errMsg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.ChallengeTemplateGenerationJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, job)
}

func (s *server) getChallengeTemplateGenerationJobs(ctx *gin.Context) {
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	locationArchetypeIDParam := strings.TrimSpace(ctx.Query("locationArchetypeId"))
	var (
		jobsList []models.ChallengeTemplateGenerationJob
		err      error
	)
	if locationArchetypeIDParam != "" {
		locationArchetypeID, parseErr := uuid.Parse(locationArchetypeIDParam)
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid locationArchetypeId"})
			return
		}
		jobsList, err = s.dbClient.ChallengeTemplateGenerationJob().FindByLocationArchetypeID(ctx, locationArchetypeID, limit)
	} else {
		jobsList, err = s.dbClient.ChallengeTemplateGenerationJob().FindRecent(ctx, limit)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getChallengeTemplateGenerationJob(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge template generation job ID"})
		return
	}
	job, err := s.dbClient.ChallengeTemplateGenerationJob().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge template generation job not found"})
		return
	}
	ctx.JSON(http.StatusOK, job)
}
