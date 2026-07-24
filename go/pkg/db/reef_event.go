package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type reefEventHandle struct {
	db *gorm.DB
}

func (h *reefEventHandle) Create(ctx context.Context, event *models.ReefEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	return h.db.WithContext(ctx).Create(event).Error
}

func (h *reefEventHandle) CountByType(ctx context.Context, eventType string, since time.Time) (int64, error) {
	var count int64
	err := h.db.WithContext(ctx).Model(&models.ReefEvent{}).
		Where("event_type = ? AND created_at >= ?", eventType, since).
		Count(&count).Error
	return count, err
}

type RuleRejectionCount struct {
	Rule  string
	Count int64
}

func (h *reefEventHandle) CountRejectionsByRule(ctx context.Context, since time.Time) ([]RuleRejectionCount, error) {
	var rows []RuleRejectionCount
	err := h.db.WithContext(ctx).Model(&models.ReefEvent{}).
		Select("rule, count(*) as count").
		Where("event_type = ? AND created_at >= ?", models.ReefEventValidationRejected, since).
		Group("rule").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}
