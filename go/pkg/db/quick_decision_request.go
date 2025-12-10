package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type quickDecisionRequestHandler struct {
	db *gorm.DB
}

func (h *quickDecisionRequestHandler) Create(
	ctx context.Context,
	request *models.QuickDecisionRequest,
) (*models.QuickDecisionRequest, error) {
	// Set ID and timestamps if not set
	if request.ID == uuid.Nil {
		request.ID = uuid.New()
	}
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now()
	}
	if request.UpdatedAt.IsZero() {
		request.UpdatedAt = time.Now()
	}

	// Create the request
	if err := h.db.WithContext(ctx).Create(request).Error; err != nil {
		return nil, err
	}

	// Return the created request (GORM will populate the ID and timestamps)
	return request, nil
}

func (h *quickDecisionRequestHandler) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.QuickDecisionRequest, error) {
	var requests []models.QuickDecisionRequest
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}
