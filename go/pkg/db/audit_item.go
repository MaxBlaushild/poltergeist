package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type auditItemHandler struct {
	db *gorm.DB
}

func (h *auditItemHandler) Create(ctx context.Context, matchID *uuid.UUID, userID *uuid.UUID, message string) error {
	auditItem := &models.AuditItem{
		ID:        uuid.New(),
		MatchID:   matchID,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Message:   message,
	}
	return h.db.WithContext(ctx).Create(auditItem).Error
}

func (h *auditItemHandler) GetAuditItemsForMatch(ctx context.Context, matchID uuid.UUID) ([]*models.AuditItem, error) {
	var auditItems []*models.AuditItem
	if err := h.db.WithContext(ctx).Where("match_id = ?", matchID).Order("created_at DESC").Find(&auditItems).Error; err != nil {
		return nil, err
	}
	return auditItems, nil
}

func (h *auditItemHandler) GetAuditItemsForUser(ctx context.Context, userID uuid.UUID) ([]*models.AuditItem, error) {
	var auditItems []*models.AuditItem
	if err := h.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&auditItems).Error; err != nil {
		return nil, err
	}
	return auditItems, nil
}

func (h *auditItemHandler) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.AuditItem{}).Error
}
