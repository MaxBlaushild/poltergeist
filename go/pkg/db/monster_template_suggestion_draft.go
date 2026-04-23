package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterTemplateSuggestionDraftHandle struct {
	db *gorm.DB
}

func (h *monsterTemplateSuggestionDraftHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Genre").
		Preload("MonsterTemplate").
		Preload("MonsterTemplate.Genre")
}

func (h *monsterTemplateSuggestionDraftHandle) Create(ctx context.Context, draft *models.MonsterTemplateSuggestionDraft) error {
	if draft != nil {
		draft.Status = models.NormalizeMonsterTemplateSuggestionDraftStatus(draft.Status)
		draft.MonsterType = models.NormalizeMonsterTemplateType(string(draft.MonsterType))
		draft.ZoneKind = models.NormalizeZoneKind(draft.ZoneKind)
		if draft.GenreID == uuid.Nil {
			resolvedGenreID, err := defaultMonsterGenreID(ctx, h.db)
			if err != nil {
				return err
			}
			draft.GenreID = resolvedGenreID
		}
	}
	return h.db.WithContext(ctx).Create(draft).Error
}

func (h *monsterTemplateSuggestionDraftHandle) Update(ctx context.Context, draft *models.MonsterTemplateSuggestionDraft) error {
	if draft != nil {
		draft.Status = models.NormalizeMonsterTemplateSuggestionDraftStatus(draft.Status)
		draft.MonsterType = models.NormalizeMonsterTemplateType(string(draft.MonsterType))
		draft.ZoneKind = models.NormalizeZoneKind(draft.ZoneKind)
		if draft.GenreID == uuid.Nil {
			resolvedGenreID, err := defaultMonsterGenreID(ctx, h.db)
			if err != nil {
				return err
			}
			draft.GenreID = resolvedGenreID
		}
	}
	return h.db.WithContext(ctx).Save(draft).Error
}

func (h *monsterTemplateSuggestionDraftHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.MonsterTemplateSuggestionDraft, error) {
	var draft models.MonsterTemplateSuggestionDraft
	if err := h.preloadBase(ctx).First(&draft, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &draft, nil
}

func (h *monsterTemplateSuggestionDraftHandle) FindByJobID(ctx context.Context, jobID uuid.UUID) ([]models.MonsterTemplateSuggestionDraft, error) {
	var drafts []models.MonsterTemplateSuggestionDraft
	if err := h.preloadBase(ctx).
		Where("job_id = ?", jobID).
		Order("created_at ASC").
		Find(&drafts).Error; err != nil {
		return nil, err
	}
	return drafts, nil
}

func (h *monsterTemplateSuggestionDraftHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.MonsterTemplateSuggestionDraft{}, "id = ?", id).Error
}
