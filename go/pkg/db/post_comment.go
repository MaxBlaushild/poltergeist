package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postCommentHandle struct {
	db *gorm.DB
}

func (h *postCommentHandle) Create(ctx context.Context, postID uuid.UUID, userID uuid.UUID, text string) (*models.PostComment, error) {
	comment := &models.PostComment{
		PostID:    postID,
		UserID:    userID,
		Text:      text,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.db.WithContext(ctx).Create(comment).Error; err != nil {
		return nil, err
	}

	return comment, nil
}

func (h *postCommentHandle) Delete(ctx context.Context, commentID uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.PostComment{}, "id = ?", commentID).Error
}

func (h *postCommentHandle) FindByPostID(ctx context.Context, postID uuid.UUID) ([]models.PostComment, error) {
	var comments []models.PostComment
	if err := h.db.WithContext(ctx).Where("post_id = ?", postID).Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (h *postCommentHandle) FindByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]models.PostComment, error) {
	var comments []models.PostComment
	if err := h.db.WithContext(ctx).Where("post_id IN (?)", postIDs).Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (h *postCommentHandle) FindByID(ctx context.Context, commentID uuid.UUID) (*models.PostComment, error) {
	var comment models.PostComment
	if err := h.db.WithContext(ctx).Where("id = ?", commentID).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

func (h *postCommentHandle) GetCommentCount(ctx context.Context, postID uuid.UUID) (int64, error) {
	var count int64
	if err := h.db.WithContext(ctx).Model(&models.PostComment{}).Where("post_id = ?", postID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
