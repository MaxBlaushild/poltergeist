package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type communityPollHandler struct {
	db *gorm.DB
}

func (h *communityPollHandler) Create(
	ctx context.Context,
	poll *models.CommunityPoll,
) (*models.CommunityPoll, error) {
	// Set ID and timestamps if not set
	if poll.ID == uuid.Nil {
		poll.ID = uuid.New()
	}
	if poll.CreatedAt.IsZero() {
		poll.CreatedAt = time.Now()
	}
	if poll.UpdatedAt.IsZero() {
		poll.UpdatedAt = time.Now()
	}

	// Create the poll
	if err := h.db.WithContext(ctx).Create(poll).Error; err != nil {
		return nil, err
	}

	// Return the created poll
	return poll, nil
}

func (h *communityPollHandler) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.CommunityPoll, error) {
	var polls []models.CommunityPoll
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&polls).Error; err != nil {
		return nil, err
	}
	return polls, nil
}
