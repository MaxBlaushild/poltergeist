package db

import (
	"context"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type scenarioTemplateGenerationDraftHandle struct {
	db *gorm.DB
}

func (h *scenarioTemplateGenerationDraftHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Genre").
		Preload("ScenarioTemplate").
		Preload("ScenarioTemplate.Genre")
}

func (h *scenarioTemplateGenerationDraftHandle) normalizeDraft(
	ctx context.Context,
	draft *models.ScenarioTemplateGenerationDraft,
) error {
	if draft == nil {
		return nil
	}

	draft.Status = models.NormalizeScenarioTemplateGenerationDraftStatus(draft.Status)
	draft.ZoneKind = models.NormalizeZoneKind(draft.ZoneKind)
	draft.Prompt = strings.TrimSpace(draft.Prompt)
	draft.Payload.ZoneKind = models.NormalizeZoneKind(draft.Payload.ZoneKind)
	draft.Payload.Prompt = strings.TrimSpace(draft.Payload.Prompt)

	if draft.Prompt == "" {
		draft.Prompt = draft.Payload.Prompt
	}
	if draft.Payload.Prompt == "" {
		draft.Payload.Prompt = draft.Prompt
	}

	if draft.ZoneKind == "" {
		draft.ZoneKind = draft.Payload.ZoneKind
	}
	if draft.Payload.ZoneKind == "" {
		draft.Payload.ZoneKind = draft.ZoneKind
	}

	resolvedGenreID, err := resolveScenarioTemplateGenerationDraftGenreID(ctx, h.db, draft)
	if err != nil {
		return err
	}
	draft.GenreID = resolvedGenreID
	if draft.Payload.GenreID == uuid.Nil {
		draft.Payload.GenreID = resolvedGenreID
	}

	draft.OpenEnded = draft.Payload.OpenEnded
	draft.Difficulty = draft.Payload.Difficulty

	return nil
}

func (h *scenarioTemplateGenerationDraftHandle) Create(
	ctx context.Context,
	draft *models.ScenarioTemplateGenerationDraft,
) error {
	if err := h.normalizeDraft(ctx, draft); err != nil {
		return err
	}
	return h.db.WithContext(ctx).Create(draft).Error
}

func (h *scenarioTemplateGenerationDraftHandle) Update(
	ctx context.Context,
	draft *models.ScenarioTemplateGenerationDraft,
) error {
	if err := h.normalizeDraft(ctx, draft); err != nil {
		return err
	}
	return h.db.WithContext(ctx).Save(draft).Error
}

func (h *scenarioTemplateGenerationDraftHandle) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.ScenarioTemplateGenerationDraft, error) {
	var draft models.ScenarioTemplateGenerationDraft
	if err := h.preloadBase(ctx).First(&draft, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &draft, nil
}

func (h *scenarioTemplateGenerationDraftHandle) FindByJobID(
	ctx context.Context,
	jobID uuid.UUID,
) ([]models.ScenarioTemplateGenerationDraft, error) {
	var drafts []models.ScenarioTemplateGenerationDraft
	if err := h.preloadBase(ctx).
		Where("job_id = ?", jobID).
		Order("created_at ASC").
		Find(&drafts).Error; err != nil {
		return nil, err
	}
	return drafts, nil
}

func (h *scenarioTemplateGenerationDraftHandle) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	return h.db.WithContext(ctx).
		Delete(&models.ScenarioTemplateGenerationDraft{}, "id = ?", id).
		Error
}
