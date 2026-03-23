package server

import (
	"encoding/json"
	stdErrors "errors"
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
	"gorm.io/gorm"
)

type paginatedChallengeResponse struct {
	Items    []models.Challenge `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

var challengeValidStatTags = map[string]struct{}{
	"strength":     {},
	"dexterity":    {},
	"constitution": {},
	"intelligence": {},
	"wisdom":       {},
	"charisma":     {},
}

type challengeUpsertRequest struct {
	ZoneID              string                      `json:"zoneId"`
	PointOfInterestID   string                      `json:"pointOfInterestId"`
	Latitude            float64                     `json:"latitude"`
	Longitude           float64                     `json:"longitude"`
	PolygonPoints       [][2]float64                `json:"polygonPoints"`
	Question            string                      `json:"question"`
	Description         string                      `json:"description"`
	ImageURL            string                      `json:"imageUrl"`
	ThumbnailURL        string                      `json:"thumbnailUrl"`
	ScaleWithUserLevel  bool                        `json:"scaleWithUserLevel"`
	RecurrenceFrequency *string                     `json:"recurrenceFrequency"`
	RewardMode          string                      `json:"rewardMode"`
	RandomRewardSize    string                      `json:"randomRewardSize"`
	RewardExperience    int                         `json:"rewardExperience"`
	Reward              int                         `json:"reward"`
	MaterialRewards     []baseMaterialRewardPayload `json:"materialRewards"`
	InventoryItemID     *int                        `json:"inventoryItemId"`
	ItemChoiceRewards   []scenarioRewardItemPayload `json:"itemChoiceRewards"`
	SubmissionType      string                      `json:"submissionType"`
	Difficulty          int                         `json:"difficulty"`
	StatTags            []string                    `json:"statTags"`
	Proficiency         string                      `json:"proficiency"`
}

type challengeGenerationJobRequest struct {
	ZoneID            string `json:"zoneId"`
	PointOfInterestID string `json:"pointOfInterestId"`
	Count             int    `json:"count"`
}

func parseChallengeStatTags(raw []string) models.StringArray {
	if len(raw) == 0 {
		return models.StringArray{}
	}
	out := models.StringArray{}
	seen := map[string]struct{}{}
	for _, tag := range raw {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		if _, ok := challengeValidStatTags[normalized]; !ok {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func (s *server) parseChallengeUpsertRequest(ctx *gin.Context, body challengeUpsertRequest) (*models.Challenge, []models.ChallengeItemChoiceReward, error) {
	zoneID, err := uuid.Parse(strings.TrimSpace(body.ZoneID))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid zoneId")
	}
	question := strings.TrimSpace(body.Question)
	if question == "" {
		return nil, nil, fmt.Errorf("question is required")
	}
	description := strings.TrimSpace(body.Description)
	if body.RewardExperience < 0 {
		return nil, nil, fmt.Errorf("rewardExperience must be zero or greater")
	}
	if body.Reward < 0 {
		return nil, nil, fmt.Errorf("reward must be zero or greater")
	}
	materialRewards, err := parseBaseMaterialRewards(body.MaterialRewards, "materialRewards")
	if err != nil {
		return nil, nil, err
	}
	if body.Difficulty < 0 {
		return nil, nil, fmt.Errorf("difficulty must be zero or greater")
	}

	submissionTypeRaw := strings.TrimSpace(body.SubmissionType)
	if submissionTypeRaw == "" {
		submissionTypeRaw = string(models.DefaultQuestNodeSubmissionType())
	}
	submissionType := models.QuestNodeSubmissionType(submissionTypeRaw)
	if !submissionType.IsValid() {
		return nil, nil, fmt.Errorf("invalid submissionType")
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
	pointOfInterestID, err := parseStandalonePointOfInterestID(body.PointOfInterestID)
	if err != nil {
		return nil, nil, err
	}
	hasPolygon := len(body.PolygonPoints) > 0
	if pointOfInterestID != nil && hasPolygon {
		return nil, nil, fmt.Errorf("challenge location must use either pointOfInterestId or polygonPoints, not both")
	}

	itemChoiceRewards := make([]models.ChallengeItemChoiceReward, 0, len(body.ItemChoiceRewards))
	for _, reward := range body.ItemChoiceRewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			return nil, nil, fmt.Errorf("itemChoiceRewards require inventoryItemId and positive quantity")
		}
		itemChoiceRewards = append(itemChoiceRewards, models.ChallengeItemChoiceReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	if len(itemChoiceRewards) == 1 {
		return nil, nil, fmt.Errorf("itemChoiceRewards must include at least 2 options when provided")
	}
	rewardMode := models.NormalizeRewardMode(body.RewardMode)
	if strings.TrimSpace(body.RewardMode) == "" {
		if body.RewardExperience > 0 || body.Reward > 0 || body.InventoryItemID != nil || len(itemChoiceRewards) > 0 || len(materialRewards) > 0 {
			rewardMode = models.RewardModeExplicit
		}
	}
	if rewardMode == models.RewardModeRandom && len(itemChoiceRewards) > 0 {
		return nil, nil, fmt.Errorf("itemChoiceRewards require explicit rewardMode")
	}
	randomRewardSize := models.NormalizeRandomRewardSize(body.RandomRewardSize)

	challenge := &models.Challenge{
		ZoneID:             zoneID,
		Question:           question,
		Description:        description,
		ImageURL:           imageURL,
		ThumbnailURL:       thumbnailURL,
		ScaleWithUserLevel: body.ScaleWithUserLevel,
		RewardMode:         rewardMode,
		RandomRewardSize:   randomRewardSize,
		RewardExperience:   body.RewardExperience,
		Reward:             body.Reward,
		MaterialRewards:    materialRewards,
		InventoryItemID:    body.InventoryItemID,
		SubmissionType:     submissionType,
		Difficulty:         body.Difficulty,
		StatTags:           parseChallengeStatTags(body.StatTags),
		Proficiency:        proficiencyPtr,
	}
	if hasPolygon {
		if err := challenge.SetPolygonPoints(body.PolygonPoints); err != nil {
			return nil, nil, err
		}
	} else {
		resolvedPointOfInterestID, resolvedLatitude, resolvedLongitude, err := s.resolveStandaloneLocation(
			ctx,
			&zoneID,
			pointOfInterestID,
			body.Latitude,
			body.Longitude,
		)
		if err != nil {
			return nil, nil, err
		}
		challenge.PointOfInterestID = resolvedPointOfInterestID
		challenge.Latitude = resolvedLatitude
		challenge.Longitude = resolvedLongitude
	}
	return challenge, itemChoiceRewards, nil
}

func (s *server) getChallenges(ctx *gin.Context) {
	challenges, err := s.dbClient.Challenge().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, challenges)
}

func (s *server) getAdminChallenges(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	page := parseAdminMonsterListPage(ctx)
	pageSize := parseAdminMonsterListPageSize(ctx)
	result, err := s.dbClient.Challenge().ListAdmin(ctx, db.ChallengeAdminListParams{
		Page:      page,
		PageSize:  pageSize,
		Query:     ctx.Query("query"),
		ZoneQuery: ctx.Query("zoneQuery"),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, paginatedChallengeResponse{
		Items:    result.Challenges,
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *server) getChallenge(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	challengeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge ID"})
		return
	}
	challenge, err := s.dbClient.Challenge().FindByID(ctx, challengeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if challenge == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
		return
	}
	userLevel, err := s.currentUserLevel(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	scaled := challengeWithScaledDifficulty(*challenge, userLevel)
	ctx.JSON(http.StatusOK, scaled)
}

func (s *server) getChallengesForZone(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	challenges, err := s.dbClient.Challenge().FindByZoneIDExcludingQuestNodes(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	challengeIDs := make([]uuid.UUID, 0, len(challenges))
	for _, challenge := range challenges {
		challengeIDs = append(challengeIDs, challenge.ID)
	}
	completedIDs, err := s.dbClient.Challenge().FindCompletedChallengeIDsByUser(ctx, user.ID, challengeIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	completedSet := make(map[uuid.UUID]struct{}, len(completedIDs))
	for _, id := range completedIDs {
		completedSet[id] = struct{}{}
	}
	userLevel, err := s.currentUserLevel(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]models.Challenge, 0, len(challenges))
	for i := range challenges {
		if _, completed := completedSet[challenges[i].ID]; completed {
			continue
		}
		response = append(response, challengeWithScaledDifficulty(challenges[i], userLevel))
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) createChallenge(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody challengeUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	challenge, itemChoiceRewards, err := s.parseChallengeUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	challenge.ID = uuid.New()
	challenge.CreatedAt = time.Now()
	challenge.UpdatedAt = challenge.CreatedAt
	if err := applyStandaloneRecurrenceForCreate(
		requestBody.RecurrenceFrequency,
		challenge.CreatedAt,
		&challenge.RecurringChallengeID,
		&challenge.RecurrenceFrequency,
		&challenge.NextRecurrenceAt,
	); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Challenge().Create(ctx, challenge); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Challenge().ReplaceItemChoiceRewards(ctx, challenge.ID, itemChoiceRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	created, err := s.dbClient.Challenge().FindByID(ctx, challenge.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateChallenge(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	challengeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge ID"})
		return
	}

	existing, err := s.dbClient.Challenge().FindByID(ctx, challengeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
		return
	}

	var requestBody challengeUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates, itemChoiceRewards, err := s.parseChallengeUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates.RecurringChallengeID = existing.RecurringChallengeID
	updates.RecurrenceFrequency = existing.RecurrenceFrequency
	updates.NextRecurrenceAt = existing.NextRecurrenceAt
	if err := applyStandaloneRecurrenceForUpdate(
		requestBody.RecurrenceFrequency,
		time.Now(),
		&updates.RecurringChallengeID,
		&updates.RecurrenceFrequency,
		&updates.NextRecurrenceAt,
	); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates.UpdatedAt = time.Now()

	if err := s.dbClient.Challenge().Update(ctx, challengeID, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Challenge().ReplaceItemChoiceRewards(ctx, challengeID, itemChoiceRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := s.dbClient.Challenge().FindByID(ctx, challengeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteChallenge(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	challengeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge ID"})
		return
	}
	if _, err := s.dbClient.Challenge().FindByID(ctx, challengeID); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	linkedToQuestNode, err := s.dbClient.Challenge().IsLinkedToQuestNode(ctx, challengeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if linkedToQuestNode {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "challenge is referenced by a quest node and cannot be deleted directly",
		})
		return
	}
	if err := s.dbClient.Challenge().Delete(ctx, challengeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "challenge deleted successfully"})
}

func (s *server) generateChallengeImage(ctx *gin.Context) {
	id := ctx.Param("id")
	challengeID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge ID"})
		return
	}

	challenge, err := s.dbClient.Challenge().FindByID(ctx, challengeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if challenge == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
		return
	}

	payload := jobs.GenerateChallengeImageTaskPayload{
		ChallengeID: challengeID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateChallengeImageTaskType, payloadBytes)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"status":    "queued",
		"challenge": challenge,
	})
}

func (s *server) createChallengeGenerationJob(ctx *gin.Context) {
	var requestBody challengeGenerationJobRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(strings.TrimSpace(requestBody.ZoneID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoneId"})
		return
	}

	count := requestBody.Count
	if count <= 0 {
		count = 1
	}
	if count > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "count must be between 1 and 100"})
		return
	}

	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if zone == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
		return
	}

	pointOfInterestID, err := parseStandalonePointOfInterestID(requestBody.PointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if pointOfInterestID != nil {
		if _, _, _, err := s.resolveStandaloneLocation(ctx, &zoneID, pointOfInterestID, 0, 0); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	job := &models.ChallengeGenerationJob{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		ZoneID:            zoneID,
		PointOfInterestID: pointOfInterestID,
		Status:            models.ChallengeGenerationStatusQueued,
		Count:             count,
		CreatedCount:      0,
	}
	if err := s.dbClient.ChallengeGenerationJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.GenerateChallengesTaskPayload{
		JobID: job.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateChallengesTaskType, payload)); err != nil {
		errMsg := err.Error()
		job.Status = models.ChallengeGenerationStatusFailed
		job.ErrorMessage = &errMsg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.ChallengeGenerationJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, job)
}

func (s *server) getChallengeGenerationJobs(ctx *gin.Context) {
	zoneIDParam := strings.TrimSpace(ctx.Query("zoneId"))
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var (
		jobsList []models.ChallengeGenerationJob
		err      error
	)
	if zoneIDParam != "" {
		zoneID, parseErr := uuid.Parse(zoneIDParam)
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoneId"})
			return
		}
		jobsList, err = s.dbClient.ChallengeGenerationJob().FindByZoneID(ctx, zoneID, limit)
	} else {
		jobsList, err = s.dbClient.ChallengeGenerationJob().FindRecent(ctx, limit)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getChallengeGenerationJob(ctx *gin.Context) {
	id := ctx.Param("id")
	jobID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge generation job ID"})
		return
	}

	job, err := s.dbClient.ChallengeGenerationJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge generation job not found"})
		return
	}

	ctx.JSON(http.StatusOK, job)
}
