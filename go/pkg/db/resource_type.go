package db

import (
	"context"
	stdErrors "errors"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type resourceTypeHandle struct {
	db *gorm.DB
}

func (h *resourceTypeHandle) Create(ctx context.Context, resourceType *models.ResourceType) error {
	resourceType.ID = uuid.New()
	resourceType.CreatedAt = time.Now()
	resourceType.UpdatedAt = resourceType.CreatedAt
	resourceType.Name = strings.TrimSpace(resourceType.Name)
	resourceType.Slug = strings.ToLower(strings.TrimSpace(resourceType.Slug))
	resourceType.Description = strings.TrimSpace(resourceType.Description)
	resourceType.MapIconURL = strings.TrimSpace(resourceType.MapIconURL)
	resourceType.MapIconPrompt = strings.TrimSpace(resourceType.MapIconPrompt)
	return h.db.WithContext(ctx).Create(resourceType).Error
}

func (h *resourceTypeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ResourceType, error) {
	var resourceType models.ResourceType
	if err := h.db.WithContext(ctx).First(&resourceType, id).Error; err != nil {
		return nil, err
	}
	return &resourceType, nil
}

func (h *resourceTypeHandle) FindBySlug(ctx context.Context, slug string) (*models.ResourceType, error) {
	var resourceType models.ResourceType
	err := h.db.WithContext(ctx).
		Where("slug = ?", strings.ToLower(strings.TrimSpace(slug))).
		First(&resourceType).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &resourceType, nil
}

func (h *resourceTypeHandle) FindAll(ctx context.Context) ([]models.ResourceType, error) {
	var resourceTypes []models.ResourceType
	if err := h.db.WithContext(ctx).
		Order("LOWER(name) ASC, created_at ASC").
		Find(&resourceTypes).Error; err != nil {
		return nil, err
	}
	return resourceTypes, nil
}

func (h *resourceTypeHandle) Update(ctx context.Context, id uuid.UUID, updates *models.ResourceType) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	payload := map[string]interface{}{
		"name":            strings.TrimSpace(updates.Name),
		"slug":            strings.ToLower(strings.TrimSpace(updates.Slug)),
		"description":     strings.TrimSpace(updates.Description),
		"map_icon_url":    strings.TrimSpace(updates.MapIconURL),
		"map_icon_prompt": strings.TrimSpace(updates.MapIconPrompt),
		"updated_at":      updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).
		Model(&models.ResourceType{}).
		Where("id = ?", id).
		Updates(payload).Error
}

func (h *resourceTypeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.ResourceType{}, "id = ?", id).Error
}
