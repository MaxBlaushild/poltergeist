package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type documentTagHandler struct {
	db *gorm.DB
}

func (h *documentTagHandler) FindOrCreateByText(ctx context.Context, text string) (*models.DocumentTag, error) {
	// Normalize text (trim and lowercase for case-insensitive matching)
	normalizedText := strings.TrimSpace(strings.ToLower(text))
	
	// Search for existing tag with case-insensitive comparison
	var tag models.DocumentTag
	err := h.db.WithContext(ctx).
		Where("LOWER(text) = ?", normalizedText).
		First(&tag).Error
	
	if err == nil {
		// Tag exists, return it
		return &tag, nil
	}
	
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Error other than not found
		return nil, err
	}
	
	// Tag doesn't exist, create it with original text (preserve case)
	tag = models.DocumentTag{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Text:      strings.TrimSpace(text),
	}
	
	if err := h.db.WithContext(ctx).Create(&tag).Error; err != nil {
		return nil, err
	}
	
	return &tag, nil
}

func (h *documentTagHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.DocumentTag, error) {
	var tag models.DocumentTag
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (h *documentTagHandler) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.DocumentTag, error) {
	if len(ids) == 0 {
		return []models.DocumentTag{}, nil
	}
	
	var tags []models.DocumentTag
	if err := h.db.WithContext(ctx).Where("id IN ?", ids).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

