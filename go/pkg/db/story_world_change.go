package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type storyWorldChangeHandle struct {
	db *gorm.DB
}

func (h *storyWorldChangeHandle) CreateBatch(
	ctx context.Context,
	changes []models.StoryWorldChange,
) error {
	if len(changes) == 0 {
		return nil
	}
	return h.db.WithContext(ctx).Create(&changes).Error
}

func (h *storyWorldChangeHandle) FindAll(
	ctx context.Context,
) ([]models.StoryWorldChange, error) {
	var changes []models.StoryWorldChange
	if err := h.db.WithContext(ctx).
		Order("priority DESC").
		Order("beat_order DESC").
		Order("created_at DESC").
		Find(&changes).Error; err != nil {
		return nil, err
	}
	return changes, nil
}

func (h *storyWorldChangeHandle) FindByMainStoryTemplateID(
	ctx context.Context,
	templateID uuid.UUID,
) ([]models.StoryWorldChange, error) {
	var changes []models.StoryWorldChange
	if err := h.db.WithContext(ctx).
		Where("main_story_template_id = ?", templateID).
		Order("priority DESC").
		Order("beat_order ASC").
		Find(&changes).Error; err != nil {
		return nil, err
	}
	return changes, nil
}
