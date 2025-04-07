package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tagGroupHandle struct {
	db *gorm.DB
}

func (h *tagGroupHandle) FindAll(ctx context.Context) ([]*models.TagGroup, error) {
	var tagGroups []*models.TagGroup
	if err := h.db.WithContext(ctx).Preload("Tags").Find(&tagGroups).Error; err != nil {
		return nil, err
	}
	return tagGroups, nil
}

func (h *tagGroupHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.TagGroup, error) {
	var tagGroup models.TagGroup
	if err := h.db.WithContext(ctx).Preload("Tags").Where("id = ?", id).First(&tagGroup).Error; err != nil {
		return nil, err
	}
	return &tagGroup, nil
}

func (h *tagGroupHandle) Create(ctx context.Context, tagGroup *models.TagGroup) error {
	return h.db.WithContext(ctx).Create(tagGroup).Error
}

func (h *tagGroupHandle) Update(ctx context.Context, tagGroup *models.TagGroup) error {
	return h.db.WithContext(ctx).Save(tagGroup).Error
}

func (h *tagGroupHandle) Delete(ctx context.Context, tagGroup *models.TagGroup) error {
	return h.db.WithContext(ctx).Delete(tagGroup).Error
}
