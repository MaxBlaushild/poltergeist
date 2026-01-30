package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postFlagHandle struct {
	db *gorm.DB
}

func (h *postFlagHandle) Create(ctx context.Context, postID, userID uuid.UUID) error {
	f := &models.PostFlag{PostID: postID, UserID: userID}
	return h.db.WithContext(ctx).Create(f).Error
}

func (h *postFlagHandle) DeleteByPostID(ctx context.Context, postID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("post_id = ?", postID).Delete(&models.PostFlag{}).Error
}

func (h *postFlagHandle) FindFlaggedPostIDs(ctx context.Context) ([]uuid.UUID, error) {
	var postIDs []uuid.UUID
	err := h.db.WithContext(ctx).Model(&models.PostFlag{}).
		Distinct("post_id").
		Pluck("post_id", &postIDs).Error
	return postIDs, err
}

func (h *postFlagHandle) GetFlagCount(ctx context.Context, postID uuid.UUID) (int64, error) {
	var count int64
	err := h.db.WithContext(ctx).Model(&models.PostFlag{}).
		Where("post_id = ?", postID).
		Count(&count).Error
	return count, err
}

func (h *postFlagHandle) IsFlaggedByUser(ctx context.Context, postID, userID uuid.UUID) (bool, error) {
	var count int64
	err := h.db.WithContext(ctx).Model(&models.PostFlag{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error
	return count > 0, err
}
