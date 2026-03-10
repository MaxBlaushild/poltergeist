package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type feedbackItemHandle struct {
	db *gorm.DB
}

func (h *feedbackItemHandle) Create(ctx context.Context, item *models.FeedbackItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	item.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Create(item).Error
}

func (h *feedbackItemHandle) ListRecent(ctx context.Context, limit int) ([]models.FeedbackItem, error) {
	if limit <= 0 {
		limit = 100
	}
	var items []models.FeedbackItem
	err := h.db.WithContext(ctx).
		Preload("User").
		Preload("Zone").
		Order("created_at DESC").
		Limit(limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
