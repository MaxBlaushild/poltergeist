package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mainStoryDistrictRunHandle struct {
	db *gorm.DB
}

func (h *mainStoryDistrictRunHandle) Create(ctx context.Context, run *models.MainStoryDistrictRun) error {
	if run != nil {
		run.Status = models.NormalizeMainStoryDistrictRunStatus(run.Status)
	}
	return h.db.WithContext(ctx).Create(run).Error
}

func (h *mainStoryDistrictRunHandle) Update(ctx context.Context, run *models.MainStoryDistrictRun) error {
	if run != nil {
		run.Status = models.NormalizeMainStoryDistrictRunStatus(run.Status)
	}
	return h.db.WithContext(ctx).Save(run).Error
}

func (h *mainStoryDistrictRunHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.MainStoryDistrictRun, error) {
	var run models.MainStoryDistrictRun
	if err := h.db.WithContext(ctx).First(&run, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (h *mainStoryDistrictRunHandle) FindAll(ctx context.Context) ([]models.MainStoryDistrictRun, error) {
	var runs []models.MainStoryDistrictRun
	if err := h.db.WithContext(ctx).
		Order("updated_at DESC").
		Order("created_at DESC").
		Find(&runs).Error; err != nil {
		return nil, err
	}
	return runs, nil
}

func (h *mainStoryDistrictRunHandle) FindByMainStoryTemplateID(ctx context.Context, templateID uuid.UUID) ([]models.MainStoryDistrictRun, error) {
	var runs []models.MainStoryDistrictRun
	if err := h.db.WithContext(ctx).
		Where("main_story_template_id = ?", templateID).
		Order("updated_at DESC").
		Order("created_at DESC").
		Find(&runs).Error; err != nil {
		return nil, err
	}
	return runs, nil
}
