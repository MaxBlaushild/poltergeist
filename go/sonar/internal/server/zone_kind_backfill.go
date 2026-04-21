package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func (s *server) setZoneKindBackfillStatus(
	ctx context.Context,
	status jobs.ZoneKindBackfillStatus,
) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client unavailable")
	}
	payload, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return s.redisClient.Set(
		ctx,
		jobs.ZoneKindBackfillStatusKey(status.JobID),
		payload,
		jobs.ZoneKindBackfillStatusTTL,
	).Err()
}

func (s *server) getZoneKindBackfillStatus(
	ctx context.Context,
	jobID uuid.UUID,
) (*jobs.ZoneKindBackfillStatus, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client unavailable")
	}
	value, err := s.redisClient.Get(ctx, jobs.ZoneKindBackfillStatusKey(jobID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var status jobs.ZoneKindBackfillStatus
	if err := json.Unmarshal([]byte(value), &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *server) backfillContentZoneKinds(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if s.asyncClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "async client unavailable"})
		return
	}
	if s.redisClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "redis client unavailable"})
		return
	}

	jobID := uuid.New()
	queuedAt := time.Now().UTC()
	status := jobs.ZoneKindBackfillStatus{
		JobID:     jobID,
		Status:    jobs.ZoneKindBackfillStatusQueued,
		Summary:   jobs.ZoneKindBackfillSummary{Results: []jobs.ZoneKindBackfillResult{}},
		QueuedAt:  &queuedAt,
		UpdatedAt: queuedAt,
	}
	if err := s.setZoneKindBackfillStatus(ctx.Request.Context(), status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.BackfillContentZoneKindsTaskPayload{
		JobID: jobID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.BackfillContentZoneKindsTaskType, payloadBytes)); err != nil {
		failedAt := time.Now().UTC()
		status.Status = jobs.ZoneKindBackfillStatusFailed
		status.Error = err.Error()
		status.CompletedAt = &failedAt
		status.UpdatedAt = failedAt
		_ = s.setZoneKindBackfillStatus(ctx.Request.Context(), status)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, status)
}

func (s *server) getBackfillContentZoneKindsStatus(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	jobID, err := uuid.Parse(ctx.Param("jobId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	status, err := s.getZoneKindBackfillStatus(ctx.Request.Context(), jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if status == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone kind backfill job not found"})
		return
	}

	ctx.JSON(http.StatusOK, status)
}
