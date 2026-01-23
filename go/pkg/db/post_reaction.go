package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postReactionHandle struct {
	db *gorm.DB
}

func (h *postReactionHandle) CreateOrUpdate(ctx context.Context, postID uuid.UUID, userID uuid.UUID, emoji string) (*models.PostReaction, error) {
	var reaction models.PostReaction
	
	// Try to find existing reaction
	err := h.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postID, userID).First(&reaction).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new reaction
		reaction = models.PostReaction{
			PostID:    postID,
			UserID:    userID,
			Emoji:     emoji,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := h.db.WithContext(ctx).Create(&reaction).Error; err != nil {
			return nil, err
		}
		return &reaction, nil
	} else if err != nil {
		return nil, err
	}
	
	// Update existing reaction
	reaction.Emoji = emoji
	reaction.UpdatedAt = time.Now()
	if err := h.db.WithContext(ctx).Save(&reaction).Error; err != nil {
		return nil, err
	}
	
	return &reaction, nil
}

func (h *postReactionHandle) Delete(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postID, userID).Delete(&models.PostReaction{}).Error
}

func (h *postReactionHandle) FindByPostID(ctx context.Context, postID uuid.UUID) ([]models.PostReaction, error) {
	var reactions []models.PostReaction
	if err := h.db.WithContext(ctx).Where("post_id = ?", postID).Find(&reactions).Error; err != nil {
		return nil, err
	}
	return reactions, nil
}

func (h *postReactionHandle) FindByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]models.PostReaction, error) {
	var reactions []models.PostReaction
	if err := h.db.WithContext(ctx).Where("post_id IN (?)", postIDs).Find(&reactions).Error; err != nil {
		return nil, err
	}
	return reactions, nil
}

func (h *postReactionHandle) FindByPostIDAndUserID(ctx context.Context, postID uuid.UUID, userID uuid.UUID) (*models.PostReaction, error) {
	var reaction models.PostReaction
	if err := h.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postID, userID).First(&reaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &reaction, nil
}
