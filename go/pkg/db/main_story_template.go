package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mainStoryTemplateHandle struct {
	db *gorm.DB
}

func (h *mainStoryTemplateHandle) Create(ctx context.Context, template *models.MainStoryTemplate) error {
	return h.db.WithContext(ctx).Create(template).Error
}

func (h *mainStoryTemplateHandle) Update(ctx context.Context, template *models.MainStoryTemplate) error {
	return h.db.WithContext(ctx).Save(template).Error
}

func (h *mainStoryTemplateHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.MainStoryTemplate, error) {
	var template models.MainStoryTemplate
	if err := h.db.WithContext(ctx).First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

func (h *mainStoryTemplateHandle) FindAll(ctx context.Context) ([]models.MainStoryTemplate, error) {
	var templates []models.MainStoryTemplate
	if err := h.db.WithContext(ctx).Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}
