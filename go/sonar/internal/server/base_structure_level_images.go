package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

func (s *server) getAdminBaseStructures(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	definitions, err := s.dbClient.BaseStructureDefinition().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"structures": serializeBaseStructureDefinitions(definitions),
	})
}

func (s *server) generateBaseStructureLevelImage(ctx *gin.Context) {
	s.generateBaseStructureLevelVisual(ctx, "")
}

func (s *server) generateBaseStructureLevelTopDownImage(ctx *gin.Context) {
	s.generateBaseStructureLevelVisual(ctx, "top_down")
}

func (s *server) generateBaseStructureLevelVisual(ctx *gin.Context, view string) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if s.asyncClient == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "async job queue is unavailable"})
		return
	}

	definitionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil || definitionID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid base structure definition ID"})
		return
	}
	level, err := strconv.Atoi(ctx.Param("level"))
	if err != nil || level <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid level"})
		return
	}

	definition, err := s.dbClient.BaseStructureDefinition().FindByID(ctx, definitionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "base structure definition not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if definition == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "base structure definition not found"})
		return
	}
	if level > max(definition.MaxLevel, 1) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "level exceeds maxLevel"})
		return
	}

	now := time.Now()
	visual, err := s.dbClient.BaseStructureLevelVisual().FindByDefinitionIDAndLevel(ctx, definitionID, level)
	if err != nil && err != gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if visual == nil {
		visual = &models.BaseStructureLevelVisual{
			ID:                           uuid.New(),
			CreatedAt:                    now,
			UpdatedAt:                    now,
			StructureDefinitionID:        definitionID,
			Level:                        level,
			ImageGenerationStatus:        models.BaseStructureImageGenerationStatusNone,
			TopDownImageGenerationStatus: models.BaseStructureImageGenerationStatusNone,
		}
	}
	if view == "top_down" {
		visual.TopDownImageGenerationStatus = models.BaseStructureImageGenerationStatusQueued
		visual.TopDownImageGenerationError = nil
	} else {
		visual.ImageGenerationStatus = models.BaseStructureImageGenerationStatusQueued
		visual.ImageGenerationError = nil
	}
	if err := s.dbClient.BaseStructureLevelVisual().Upsert(ctx, visual); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.GenerateBaseStructureLevelImageTaskPayload{
		StructureDefinitionID: definitionID,
		Level:                 level,
		View:                  view,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	taskType := jobs.GenerateBaseStructureLevelImageTaskType
	if view == "top_down" {
		taskType = jobs.GenerateBaseStructureLevelTopDownImageTaskType
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(taskType, payload)); err != nil {
		errMsg := err.Error()
		if view == "top_down" {
			visual.TopDownImageGenerationStatus = models.BaseStructureImageGenerationStatusFailed
			visual.TopDownImageGenerationError = &errMsg
		} else {
			visual.ImageGenerationStatus = models.BaseStructureImageGenerationStatusFailed
			visual.ImageGenerationError = &errMsg
		}
		_ = s.dbClient.BaseStructureLevelVisual().Upsert(ctx, visual)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	definition, err = s.dbClient.BaseStructureDefinition().FindByID(ctx, definitionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, serializeBaseStructureDefinition(*definition))
}
