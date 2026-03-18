package server

import (
	"encoding/json"
	"errors"
	"log"
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

func (s *server) enqueueBaseDescriptionGenerationJob(
	ctx *gin.Context,
	baseID uuid.UUID,
) (*models.BaseDescriptionGenerationJob, error) {
	if s.asyncClient == nil {
		return nil, errors.New("async job queue is unavailable")
	}

	job := &models.BaseDescriptionGenerationJob{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		BaseID:    baseID,
		Status:    models.BaseDescriptionGenerationStatusQueued,
	}
	if err := s.dbClient.BaseDescriptionGenerationJob().Create(ctx, job); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(jobs.GenerateBaseDescriptionTaskPayload{JobID: job.ID})
	if err != nil {
		return nil, err
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateBaseDescriptionTaskType, payload)); err != nil {
		errMsg := err.Error()
		job.Status = models.BaseDescriptionGenerationStatusFailed
		job.ErrorMessage = &errMsg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.BaseDescriptionGenerationJob().Update(ctx, job)
		return nil, err
	}

	return job, nil
}

func (s *server) queueBaseDescriptionGenerationFromItemUse(ctx *gin.Context, baseID uuid.UUID) {
	if s.asyncClient == nil {
		return
	}
	if _, err := s.enqueueBaseDescriptionGenerationJob(ctx, baseID); err != nil {
		log.Printf("failed to queue base description generation for base %s: %v", baseID, err)
	}
}

func (s *server) createBaseDescriptionGenerationJob(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if s.asyncClient == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "async job queue is unavailable"})
		return
	}

	baseID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil || baseID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid base ID"})
		return
	}

	base, err := s.dbClient.Base().FindByID(ctx, baseID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if base == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "base not found"})
		return
	}

	job, err := s.enqueueBaseDescriptionGenerationJob(ctx, baseID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, job)
}

func (s *server) getBaseDescriptionGenerationJobs(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	baseIDParam := strings.TrimSpace(ctx.Query("baseId"))
	limit := 50
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var (
		jobsList []models.BaseDescriptionGenerationJob
		err      error
	)
	if baseIDParam != "" {
		baseID, parseErr := uuid.Parse(baseIDParam)
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid baseId"})
			return
		}
		jobsList, err = s.dbClient.BaseDescriptionGenerationJob().FindByBaseID(ctx, baseID, limit)
	} else {
		jobsList, err = s.dbClient.BaseDescriptionGenerationJob().FindRecent(ctx, limit)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, jobsList)
}
