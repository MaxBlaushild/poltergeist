package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/gin-gonic/gin"
)

func (s *server) generateQuestForQuestArchetype(ctx *gin.Context) {
	id := ctx.Param("id")
	questArchetypeID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype ID"})
		return
	}

	var requestBody struct {
		ZoneID      uuid.UUID  `json:"zoneId"`
		CharacterID *uuid.UUID `json:"characterId"`
	}

	if err := ctx.BindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestBody.ZoneID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "zoneId is required"})
		return
	}

	questArchetype, err := s.dbClient.QuestArchetype().FindByID(ctx, questArchetypeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if questArchetype == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest archetype not found"})
		return
	}

	zone, err := s.dbClient.Zone().FindByID(ctx, requestBody.ZoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if zone == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
		return
	}

	resolvedCharacterID := requestBody.CharacterID
	if resolvedCharacterID != nil {
		character, err := s.dbClient.Character().FindByID(ctx, *resolvedCharacterID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if character == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "character not found"})
			return
		}
	} else {
		resolvedCharacterID, err = s.resolveQuestTemplateCharacterID(ctx, zone.ID, questArchetype)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	job := &models.QuestGenerationJob{
		ID:                    uuid.New(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		ZoneQuestArchetypeID:  nil,
		ZoneID:                zone.ID,
		QuestArchetypeID:      questArchetype.ID,
		QuestGiverCharacterID: resolvedCharacterID,
		Status:                models.QuestGenerationStatusQueued,
		TotalCount:            1,
		CompletedCount:        0,
		FailedCount:           0,
		QuestIDs:              models.StringArray{},
	}

	if err := s.dbClient.QuestGenerationJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.GenerateQuestForZoneTaskPayload{
		ZoneID:                zone.ID,
		QuestArchetypeID:      questArchetype.ID,
		QuestGiverCharacterID: resolvedCharacterID,
		QuestGenerationJobID:  &job.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(
		asynq.NewTask(jobs.GenerateQuestForZoneTaskType, payload),
		asynq.TaskID(questGenerationTaskID(job.ID, 0)),
	); err != nil {
		msg := err.Error()
		job.Status = models.QuestGenerationStatusFailed
		job.ErrorMessage = &msg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.QuestGenerationJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, job)
}
