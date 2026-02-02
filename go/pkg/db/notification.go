package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationHandle struct {
	db *gorm.DB
}

func (h *notificationHandle) Create(ctx context.Context, n *models.Notification) error {
	return h.db.WithContext(ctx).Create(n).Error
}

func (h *notificationHandle) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := h.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (h *notificationHandle) CountUnreadByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := h.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).Count(&count).Error
	return count, err
}

func (h *notificationHandle) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	now := time.Now()
	return h.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("read_at", now).Error
}

func (h *notificationHandle) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return h.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", now).Error
}

func (h *notificationHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	var n models.Notification
	err := h.db.WithContext(ctx).Where("id = ?", id).First(&n).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}
