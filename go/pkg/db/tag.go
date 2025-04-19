package db

import (
	"context"
	"errors"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (h *tagHandle) Upsert(ctx context.Context, tag *models.Tag) error {
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "value"}},
			DoNothing: true,
		}).
		Create(tag).
		Where("value = ?", tag.Value).
		First(tag).Error
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
		ID:                uuid.New(),
	}).Error
}

func (h *tagHandle) RemoveTagFromPointOfInterest(ctx context.Context, tagID uuid.UUID, pointOfInterestID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("tag_id = ? AND point_of_interest_id = ?", tagID, pointOfInterestID).Delete(&models.TagEntity{}).Error
}

func (h *tagHandle) FindByValue(ctx context.Context, value string) (*models.Tag, error) {
	var tag models.Tag
	if err := h.db.WithContext(ctx).Where("value = ?", value).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

func (h *tagHandle) MoveTagToTagGroup(ctx context.Context, tagID uuid.UUID, tagGroupID uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.Tag{}).Where("id = ?", tagID).Update("tag_group_id", tagGroupID).Error
}

func (h *tagHandle) CreateTagGroup(ctx context.Context, tagGroup *models.TagGroup) error {
	return h.db.WithContext(ctx).Create(tagGroup).Error
}
