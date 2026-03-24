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

func (s *server) createAndEnqueueDistrictSeedJob(
	ctx *gin.Context,
	districtID uuid.UUID,
	questArchetypes []*models.QuestArchetype,
) (*models.DistrictSeedJob, error) {
	questArchetypeIDs := make(models.StringArray, 0, len(questArchetypes))
	results := make(models.DistrictSeedResults, 0, len(questArchetypes))
	for _, questArchetype := range questArchetypes {
		if questArchetype == nil {
			continue
		}
		id := questArchetype.ID.String()
		questArchetypeIDs = append(questArchetypeIDs, id)
		results = append(results, models.DistrictSeedResult{
			QuestArchetypeID:   id,
			QuestArchetypeName: questArchetype.Name,
			Status:             models.DistrictSeedResultStatusQueued,
		})
	}
	if len(questArchetypeIDs) == 0 {
		return nil, fmt.Errorf("at least one quest template is required")
	}

	job := &models.DistrictSeedJob{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		DistrictID:        districtID,
		Status:            models.DistrictSeedJobStatusQueued,
		QuestArchetypeIDs: questArchetypeIDs,
		Results:           results,
	}
	if err := s.dbClient.DistrictSeedJob().Create(ctx, job); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(jobs.SeedDistrictTaskPayload{JobID: job.ID})
	if err != nil {
		return nil, err
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.SeedDistrictTaskType, payload)); err != nil {
		return nil, err
	}

	return job, nil
}

func normalizeDistrictSeedJobStatuses(rawStatuses []string) ([]string, error) {
	if len(rawStatuses) == 0 {
		return nil, nil
	}
	seen := map[string]struct{}{}
	statuses := make([]string, 0, len(rawStatuses))
	for _, rawStatus := range rawStatuses {
		status := strings.TrimSpace(rawStatus)
		if status == "" {
			continue
		}
		if !models.IsValidDistrictSeedJobStatus(status) {
			return nil, fmt.Errorf("invalid district seed job status: %s", status)
		}
		if _, exists := seen[status]; exists {
			continue
		}
		seen[status] = struct{}{}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func (s *server) loadDistrictSeedQuestArchetypes(
	ctx *gin.Context,
	rawIDs []string,
) ([]*models.QuestArchetype, error) {
	if len(rawIDs) == 0 {
		return nil, fmt.Errorf("questArchetypeIds array cannot be empty")
	}

	seen := map[uuid.UUID]struct{}{}
	questArchetypes := make([]*models.QuestArchetype, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		questArchetypeID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil {
			return nil, fmt.Errorf("invalid quest archetype ID: %s", rawID)
		}
		if _, exists := seen[questArchetypeID]; exists {
			continue
		}
		seen[questArchetypeID] = struct{}{}

		questArchetype, err := s.dbClient.QuestArchetype().FindByID(ctx, questArchetypeID)
		if err != nil {
			return nil, err
		}
		if questArchetype == nil {
			return nil, fmt.Errorf("quest archetype not found: %s", questArchetypeID)
		}
		questArchetypes = append(questArchetypes, questArchetype)
	}

	if len(questArchetypes) == 0 {
		return nil, fmt.Errorf("questArchetypeIds array cannot be empty")
	}

	return questArchetypes, nil
}

func (s *server) createDistrictSeedJob(ctx *gin.Context) {
	var requestBody struct {
		DistrictID        string   `json:"districtId"`
		QuestArchetypeIDs []string `json:"questArchetypeIds"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	districtID, err := uuid.Parse(strings.TrimSpace(requestBody.DistrictID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid districtId"})
		return
	}

	district, err := s.dbClient.District().FindByID(ctx, districtID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if district == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "district not found"})
		return
	}
	if len(district.Zones) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "district must include at least one zone"})
		return
	}

	questArchetypes, err := s.loadDistrictSeedQuestArchetypes(ctx, requestBody.QuestArchetypeIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := s.createAndEnqueueDistrictSeedJob(ctx, districtID, questArchetypes)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, job)
}

func (s *server) getDistrictSeedJobs(ctx *gin.Context) {
	districtIDParam := strings.TrimSpace(ctx.Query("districtId"))
	statusesParam := strings.TrimSpace(ctx.Query("statuses"))
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var districtID *uuid.UUID
	if districtIDParam != "" {
		parsedDistrictID, err := uuid.Parse(districtIDParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid districtId"})
			return
		}
		districtID = &parsedDistrictID
	}

	statuses, err := normalizeDistrictSeedJobStatuses(strings.Split(statusesParam, ","))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobsList, err := s.dbClient.DistrictSeedJob().FindFiltered(ctx, districtID, statuses, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getDistrictSeedJob(ctx *gin.Context) {
	jobID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid district seed job ID"})
		return
	}

	job, err := s.dbClient.DistrictSeedJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "district seed job not found"})
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (s *server) retryDistrictSeedJob(ctx *gin.Context) {
	jobID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid district seed job ID"})
		return
	}

	job, err := s.dbClient.DistrictSeedJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "district seed job not found"})
		return
	}
	if job.Status != models.DistrictSeedJobStatusFailed {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "district seed job is not in a failed state"})
		return
	}

	job.Status = models.DistrictSeedJobStatusQueued
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := s.dbClient.DistrictSeedJob().Update(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.SeedDistrictTaskPayload{JobID: job.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.SeedDistrictTaskType, payload)); err != nil {
		errMsg := err.Error()
		job.Status = models.DistrictSeedJobStatusFailed
		job.ErrorMessage = &errMsg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.DistrictSeedJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (s *server) deleteDistrictSeedJob(ctx *gin.Context) {
	jobID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid district seed job ID"})
		return
	}

	job, err := s.dbClient.DistrictSeedJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "district seed job not found"})
		return
	}
	if job.Status == models.DistrictSeedJobStatusInProgress {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete a district seed job while it is in progress"})
		return
	}

	if err := s.dbClient.DistrictSeedJob().DeleteByID(ctx, job.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}
