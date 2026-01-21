package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postHandle struct {
	db *gorm.DB
}

func (h *postHandle) Create(ctx context.Context, userID uuid.UUID, imageURL string, caption *string) (*models.Post, error) {
	post := &models.Post{
		UserID:   userID,
		ImageURL: imageURL,
		Caption:  caption,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.db.WithContext(ctx).Create(post).Error; err != nil {
		return nil, err
	}

	return post, nil
}

func (h *postHandle) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Post, error) {
	var posts []models.Post
	if err := h.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (h *postHandle) FindByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]models.Post, error) {
	var posts []models.Post
	if err := h.db.WithContext(ctx).Where("user_id IN (?)", userIDs).Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (h *postHandle) FindAllFriendsPosts(ctx context.Context, userID uuid.UUID) ([]models.Post, error) {
	// First, find all friends
	var friends []models.Friend
	if err := h.db.WithContext(ctx).Find(&friends).Error; err != nil {
		return nil, err
	}

	// Extract friend user IDs
	friendIDs := []uuid.UUID{}
	for _, friend := range friends {
		if friend.FirstUserID == userID {
			friendIDs = append(friendIDs, friend.SecondUserID)
		} else if friend.SecondUserID == userID {
			friendIDs = append(friendIDs, friend.FirstUserID)
		}
	}

	// If no friends, return empty list
	if len(friendIDs) == 0 {
		return []models.Post{}, nil
	}

	// Get posts from all friends, ordered by created_at DESC (reverse chronological)
	var posts []models.Post
	if err := h.db.WithContext(ctx).Where("user_id IN (?)", friendIDs).Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}

func (h *postHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	var post models.Post
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (h *postHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Post{}, "id = ?", id).Error
}

