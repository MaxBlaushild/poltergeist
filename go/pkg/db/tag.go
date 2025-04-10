package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tagHandle struct {
	db *gorm.DB
}

func (h *tagHandle) FindAll(ctx context.Context) ([]*models.Tag, error) {
	var tags []*models.Tag
	if err := h.db.WithContext(ctx).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (h *tagHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	var tag models.Tag
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (h *tagHandle) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*models.Tag, error) {
	var tags []*models.Tag
	if err := h.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (h *tagHandle) Create(ctx context.Context, tag *models.Tag) error {
	return h.db.WithContext(ctx).Create(tag).Error
}

func (h *tagHandle) Update(ctx context.Context, tag *models.Tag) error {
	return h.db.WithContext(ctx).Save(tag).Error
}

func (h *tagHandle) AddTagToPointOfInterest(ctx context.Context, tagID uuid.UUID, pointOfInterestID uuid.UUID) error {
	return h.db.WithContext(ctx).Create(&models.TagEntity{
		TagID:             tagID,
		PointOfInterestID: &pointOfInterestID,
	}).Error
}

func (h *tagHandle) RemoveTagFromPointOfInterest(ctx context.Context, tagID uuid.UUID, pointOfInterestID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("tag_id = ? AND point_of_interest_id = ?", tagID, pointOfInterestID).Delete(&models.TagEntity{}).Error
}
