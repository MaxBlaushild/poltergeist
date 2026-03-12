package server

import (
	"encoding/json"
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

type zoneFlavorGenerationJobRequest struct {
	ZoneID string `json:"zoneId"`
}

func (s *server) createZoneFlavorGenerationJob(ctx *gin.Context) {
	var requestBody zoneFlavorGenerationJobRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(strings.TrimSpace(requestBody.ZoneID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoneId"})
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

	job := &models.ZoneFlavorGenerationJob{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ZoneID:    zoneID,
		Status:    models.ZoneFlavorGenerationStatusQueued,
	}
	if err := s.dbClient.ZoneFlavorGenerationJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.GenerateZoneFlavorTaskPayload{JobID: job.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateZoneFlavorTaskType, payload)); err != nil {
		errMsg := err.Error()
		job.Status = models.ZoneFlavorGenerationStatusFailed
		job.ErrorMessage = &errMsg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.ZoneFlavorGenerationJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, job)
}

func (s *server) getZoneFlavorGenerationJobs(ctx *gin.Context) {
	zoneIDParam := strings.TrimSpace(ctx.Query("zoneId"))
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var (
		jobsList []models.ZoneFlavorGenerationJob
		err      error
	)
	if zoneIDParam != "" {
		zoneID, parseErr := uuid.Parse(zoneIDParam)
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoneId"})
			return
		}
		jobsList, err = s.dbClient.ZoneFlavorGenerationJob().FindByZoneID(ctx, zoneID, limit)
	} else {
		jobsList, err = s.dbClient.ZoneFlavorGenerationJob().FindRecent(ctx, limit)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getZoneFlavorGenerationJob(ctx *gin.Context) {
	jobID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone flavor generation job ID"})
		return
	}

	job, err := s.dbClient.ZoneFlavorGenerationJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone flavor generation job not found"})
		return
	}

	ctx.JSON(http.StatusOK, job)
}
