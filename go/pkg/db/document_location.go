package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type documentLocationHandler struct {
	db *gorm.DB
}

func (h *documentLocationHandler) Create(ctx context.Context, location *models.DocumentLocation) error {
	// Set ID and timestamps if not set
	if location.ID == uuid.Nil {
		location.ID = uuid.New()
	}
	if location.CreatedAt.IsZero() {
		location.CreatedAt = time.Now()
	}
	if location.UpdatedAt.IsZero() {
		location.UpdatedAt = time.Now()
	}

	return h.db.WithContext(ctx).Create(location).Error
}

func (h *documentLocationHandler) FindByDocumentID(ctx context.Context, documentID uuid.UUID) ([]models.DocumentLocation, error) {
	var locations []models.DocumentLocation
	if err := h.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Find(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

func (h *documentLocationHandler) DeleteByDocumentID(ctx context.Context, documentID uuid.UUID) error {
	return h.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Delete(&models.DocumentLocation{}).Error
}
