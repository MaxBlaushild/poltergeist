package server

import (
	"encoding/json"
	stdErrors "errors"
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
	"gorm.io/gorm"
)

var challengeValidStatTags = map[string]struct{}{
	"strength":     {},
	"dexterity":    {},
	"constitution": {},
	"intelligence": {},
	"wisdom":       {},
	"charisma":     {},
}

type challengeUpsertRequest struct {
	ZoneID          string   `json:"zoneId"`
	Latitude        float64  `json:"latitude"`
	Longitude       float64  `json:"longitude"`
	Question        string   `json:"question"`
	Description     string   `json:"description"`
	ImageURL        string   `json:"imageUrl"`
	ThumbnailURL    string   `json:"thumbnailUrl"`
	Reward          int      `json:"reward"`
	InventoryItemID *int     `json:"inventoryItemId"`
	SubmissionType  string   `json:"submissionType"`
	Difficulty      int      `json:"difficulty"`
	StatTags        []string `json:"statTags"`
	Proficiency     string   `json:"proficiency"`
}

type challengeGenerationJobRequest struct {
	ZoneID string `json:"zoneId"`
	Count  int    `json:"count"`
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

func parseChallengeUpsertRequest(body challengeUpsertRequest) (*models.Challenge, error) {
	zoneID, err := uuid.Parse(strings.TrimSpace(body.ZoneID))
	if err != nil {
		return nil, fmt.Errorf("invalid zoneId")
	}
	question := strings.TrimSpace(body.Question)
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}
	description := strings.TrimSpace(body.Description)
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

	challenge := &models.Challenge{
		ZoneID:          zoneID,
		Latitude:        body.Latitude,
		Longitude:       body.Longitude,
		Question:        question,
		Description:     description,
		ImageURL:        imageURL,
		ThumbnailURL:    thumbnailURL,
		Reward:          body.Reward,
		InventoryItemID: body.InventoryItemID,
		SubmissionType:  submissionType,
		Difficulty:      body.Difficulty,
		StatTags:        parseChallengeStatTags(body.StatTags),
		Proficiency:     proficiencyPtr,
	}
	return challenge, nil
}

func (s *server) getChallenges(ctx *gin.Context) {
	challenges, err := s.dbClient.Challenge().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, challenges)
}

func (s *server) getChallenge(ctx *gin.Context) {
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
	ctx.JSON(http.StatusOK, challenge)
}

func (s *server) getChallengesForZone(ctx *gin.Context) {
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
	ctx.JSON(http.StatusOK, challenges)
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

	challenge, err := parseChallengeUpsertRequest(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge payload"})
		return
	}
	challenge.ID = uuid.New()
	challenge.CreatedAt = time.Now()
	challenge.UpdatedAt = challenge.CreatedAt

	if err := s.dbClient.Challenge().Create(ctx, challenge); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, challenge)
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

	updates, err := parseChallengeUpsertRequest(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge payload"})
		return
	}
	updates.UpdatedAt = time.Now()

	if err := s.dbClient.Challenge().Update(ctx, challengeID, updates); err != nil {
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

	job := &models.ChallengeGenerationJob{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ZoneID:       zoneID,
		Status:       models.ChallengeGenerationStatusQueued,
		Count:        count,
		CreatedCount: 0,
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
