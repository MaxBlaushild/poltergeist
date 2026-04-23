package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterTemplateSuggestionJobHandle struct {
	db *gorm.DB
}

func (h *monsterTemplateSuggestionJobHandle) Create(ctx context.Context, job *models.MonsterTemplateSuggestionJob) error {
	if job != nil {
		job.Status = models.NormalizeMonsterTemplateSuggestionJobStatus(job.Status)
		job.MonsterType = models.NormalizeMonsterTemplateType(string(job.MonsterType))
		job.ZoneKind = models.NormalizeZoneKind(job.ZoneKind)
		if job.GenreID == uuid.Nil {
			resolvedGenreID, err := defaultMonsterGenreID(ctx, h.db)
			if err != nil {
				return err
			}
			job.GenreID = resolvedGenreID
		}
	}
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *monsterTemplateSuggestionJobHandle) Update(ctx context.Context, job *models.MonsterTemplateSuggestionJob) error {
	if job != nil {
		job.Status = models.NormalizeMonsterTemplateSuggestionJobStatus(job.Status)
		job.MonsterType = models.NormalizeMonsterTemplateType(string(job.MonsterType))
		job.ZoneKind = models.NormalizeZoneKind(job.ZoneKind)
		if job.GenreID == uuid.Nil {
			resolvedGenreID, err := defaultMonsterGenreID(ctx, h.db)
			if err != nil {
				return err
			}
			job.GenreID = resolvedGenreID
		}
	}
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *monsterTemplateSuggestionJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.MonsterTemplateSuggestionJob, error) {
	var job models.MonsterTemplateSuggestionJob
	if err := h.db.WithContext(ctx).
		Preload("Genre").
		First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *monsterTemplateSuggestionJobHandle) FindRecent(ctx context.Context, limit int) ([]models.MonsterTemplateSuggestionJob, error) {
	var jobs []models.MonsterTemplateSuggestionJob
	query := h.db.WithContext(ctx).
		Preload("Genre").
		Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
