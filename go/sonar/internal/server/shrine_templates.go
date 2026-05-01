package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type shrineTemplateUpsertRequest struct {
	ZoneKind          string `json:"zoneKind"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	BlessingName      string `json:"blessingName"`
	EffectDescription string `json:"effectDescription"`
	EffectKind        string `json:"effectKind"`
	BaseMagnitude     *int   `json:"baseMagnitude"`
}

type shrineTemplateGenerationJobRequest struct {
	Count    int    `json:"count"`
	ZoneKind string `json:"zoneKind"`
}

func (s *server) parseShrineTemplateUpsertRequest(
	ctx *gin.Context,
	body shrineTemplateUpsertRequest,
) (*models.ShrineTemplate, error) {
	zoneKind, err := s.resolveOptionalZoneKind(ctx, body.ZoneKind)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	blessingName := strings.TrimSpace(body.BlessingName)
	if blessingName == "" {
		return nil, fmt.Errorf("blessingName is required")
	}
	effectKind := models.NormalizeShrineEffectKind(body.EffectKind)
	if strings.TrimSpace(body.EffectKind) != "" && strings.TrimSpace(string(effectKind)) != strings.TrimSpace(strings.ToLower(body.EffectKind)) {
		return nil, fmt.Errorf("invalid effectKind")
	}
	baseMagnitude := 2
	if body.BaseMagnitude != nil {
		if *body.BaseMagnitude <= 0 {
			return nil, fmt.Errorf("baseMagnitude must be greater than zero")
		}
		baseMagnitude = *body.BaseMagnitude
	}

	template := &models.ShrineTemplate{
		Name:              name,
		Description:       strings.TrimSpace(body.Description),
		BlessingName:      blessingName,
		EffectDescription: strings.TrimSpace(body.EffectDescription),
		EffectKind:        effectKind,
		BaseMagnitude:     baseMagnitude,
	}
	if zoneKind != nil {
		template.ZoneKind = zoneKind.Slug
	}
	return template, nil
}

func (s *server) getShrineTemplates(ctx *gin.Context) {
	templates, err := s.dbClient.ShrineTemplate().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, templates)
}

func (s *server) getShrineTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid shrine template ID"})
		return
	}
	template, err := s.dbClient.ShrineTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if template == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "shrine template not found"})
		return
	}
	ctx.JSON(http.StatusOK, template)
}

func (s *server) createShrineTemplate(ctx *gin.Context) {
	var body shrineTemplateUpsertRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := s.parseShrineTemplateUpsertRequest(ctx, body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.ShrineTemplate().Create(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	created, err := s.dbClient.ShrineTemplate().FindByID(ctx, template.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateShrineTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid shrine template ID"})
		return
	}
	existing, err := s.dbClient.ShrineTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "shrine template not found"})
		return
	}

	var body shrineTemplateUpsertRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := s.parseShrineTemplateUpsertRequest(ctx, body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if template.ZoneKind == "" {
		template.ZoneKind = existing.ZoneKind
	}
	if err := s.dbClient.ShrineTemplate().Update(ctx, id, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := s.dbClient.ShrineTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteShrineTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid shrine template ID"})
		return
	}
	if err := s.dbClient.ShrineTemplate().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "shrine template deleted successfully"})
}

func (s *server) createShrineTemplateGenerationJob(ctx *gin.Context) {
	var body shrineTemplateGenerationJobRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Count <= 0 {
		body.Count = 1
	}
	if body.Count > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "count must be between 1 and 100"})
		return
	}
	zoneKind, err := s.resolveOptionalZoneKind(ctx, body.ZoneKind)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := &models.ShrineTemplateGenerationJob{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ZoneKind:     models.NormalizeZoneKind(body.ZoneKind),
		Status:       models.ShrineTemplateGenerationStatusQueued,
		Count:        body.Count,
		CreatedCount: 0,
	}
	if zoneKind != nil {
		job.ZoneKind = zoneKind.Slug
	}
	if err := s.dbClient.ShrineTemplateGenerationJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.GenerateShrineTemplatesTaskPayload{JobID: job.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateShrineTemplatesTaskType, payloadBytes)); err != nil {
		msg := err.Error()
		job.Status = models.ShrineTemplateGenerationStatusFailed
		job.ErrorMessage = &msg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.ShrineTemplateGenerationJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, job)
}

func (s *server) getShrineTemplateGenerationJobs(ctx *gin.Context) {
	limit := 20
	jobsList, err := s.dbClient.ShrineTemplateGenerationJob().FindRecent(ctx, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getShrineTemplateGenerationJob(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid shrine template generation job ID"})
		return
	}
	job, err := s.dbClient.ShrineTemplateGenerationJob().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "shrine template generation job not found"})
		return
	}
	ctx.JSON(http.StatusOK, job)
}
