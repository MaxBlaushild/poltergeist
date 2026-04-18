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

func (s *server) updateBaseStructurePrompts(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	definitionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil || definitionID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid base structure definition ID"})
		return
	}

	var body struct {
		ImagePrompt        string `json:"imagePrompt"`
		TopDownImagePrompt string `json:"topDownImagePrompt"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imagePrompt := strings.TrimSpace(body.ImagePrompt)
	topDownImagePrompt := strings.TrimSpace(body.TopDownImagePrompt)
	if len(imagePrompt) > 8000 || len(topDownImagePrompt) > 8000 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "prompts must be 8000 characters or fewer"})
		return
	}

	if err := s.dbClient.BaseStructureDefinition().UpdatePrompts(
		ctx,
		definitionID,
		imagePrompt,
		topDownImagePrompt,
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	definition, err := s.dbClient.BaseStructureDefinition().FindByID(ctx, definitionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, serializeBaseStructureDefinition(*definition))
}

func (s *server) updateBaseStructureHearthRecoveryConfig(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	definitionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil || definitionID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid base structure definition ID"})
		return
	}

	definition, err := s.dbClient.BaseStructureDefinition().FindByID(ctx, definitionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if definition == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "base structure definition not found"})
		return
	}
	if strings.TrimSpace(definition.Key) != "hearth" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "hearth recovery config can only be updated for the hearth"})
		return
	}

	var body struct {
		Level2Statuses []scenarioFailureStatusPayload `json:"level2Statuses"`
		Level3Statuses []scenarioFailureStatusPayload `json:"level3Statuses"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	level2Statuses, err := parseScenarioFailureStatusTemplates(body.Level2Statuses, "level2Statuses")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	level3Statuses, err := parseScenarioFailureStatusTemplates(body.Level3Statuses, "level3Statuses")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	effectConfig := definition.EffectConfig
	if effectConfig == nil {
		effectConfig = models.MetadataJSONB{}
	}
	effectConfig["hearthRecoveryStatusesByLevel"] = models.MetadataJSONB{
		"2": level2Statuses,
		"3": level3Statuses,
	}

	if err := s.dbClient.BaseStructureDefinition().UpdateEffectConfig(ctx, definitionID, effectConfig); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	definition, err = s.dbClient.BaseStructureDefinition().FindByID(ctx, definitionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, serializeBaseStructureDefinition(*definition))
}

func (s *server) updateBaseStructureChaosEngineConfig(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	definitionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil || definitionID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid base structure definition ID"})
		return
	}

	definition, err := s.dbClient.BaseStructureDefinition().FindByID(ctx, definitionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if definition == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "base structure definition not found"})
		return
	}
	if strings.TrimSpace(definition.Key) != "chaos_engine" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Chaos Engine config can only be updated for the Chaos Engine Room"})
		return
	}

	var body struct {
		RequiredInventoryItemID *int `json:"requiredInventoryItemId"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.RequiredInventoryItemID != nil && *body.RequiredInventoryItemID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "requiredInventoryItemId must be positive"})
		return
	}
	if body.RequiredInventoryItemID != nil {
		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, *body.RequiredInventoryItemID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "inventory item not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if item == nil || item.Archived {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "inventory item is unavailable"})
			return
		}
	}

	effectConfig := definition.EffectConfig
	if effectConfig == nil {
		effectConfig = models.MetadataJSONB{}
	}
	if body.RequiredInventoryItemID == nil {
		effectConfig["requiredInventoryItemId"] = nil
	} else {
		effectConfig["requiredInventoryItemId"] = *body.RequiredInventoryItemID
	}

	if err := s.dbClient.BaseStructureDefinition().UpdateEffectConfig(ctx, definitionID, effectConfig); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	definition, err = s.dbClient.BaseStructureDefinition().FindByID(ctx, definitionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, serializeBaseStructureDefinition(*definition))
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
