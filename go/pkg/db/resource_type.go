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

func (h *resourceTypeHandle) preload(db *gorm.DB) *gorm.DB {
	return db.
		Preload("GatherRequirements", func(db *gorm.DB) *gorm.DB {
			return db.Order("min_level ASC, max_level ASC, required_inventory_item_id ASC")
		}).
		Preload("GatherRequirements.RequiredInventoryItem")
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
	requirements := resourceType.GatherRequirements
	resourceType.GatherRequirements = nil
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("GatherRequirements").Create(resourceType).Error; err != nil {
			return err
		}
		return syncResourceTypeGatherRequirementsTx(tx, resourceType.ID, requirements)
	})
}

func (h *resourceTypeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ResourceType, error) {
	var resourceType models.ResourceType
	if err := h.preload(h.db.WithContext(ctx)).First(&resourceType, id).Error; err != nil {
		return nil, err
	}
	return &resourceType, nil
}

func (h *resourceTypeHandle) FindBySlug(ctx context.Context, slug string) (*models.ResourceType, error) {
	var resourceType models.ResourceType
	err := h.preload(h.db.WithContext(ctx)).
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
	if err := h.preload(h.db.WithContext(ctx)).
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
	requirements := updates.GatherRequirements
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&models.ResourceType{}).
			Where("id = ?", id).
			Updates(payload).Error; err != nil {
			return err
		}
		return syncResourceTypeGatherRequirementsTx(tx, id, requirements)
	})
}

func (h *resourceTypeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.ResourceType{}, "id = ?", id).Error
}

func syncResourceTypeGatherRequirementsTx(
	tx *gorm.DB,
	resourceTypeID uuid.UUID,
	requirements []models.ResourceGatherRequirement,
) error {
	if err := tx.Where("resource_type_id = ?", resourceTypeID).Delete(&models.ResourceGatherRequirement{}).Error; err != nil {
		return err
	}
	if len(requirements) == 0 {
		return nil
	}

	now := time.Now()
	records := make([]models.ResourceGatherRequirement, 0, len(requirements))
	for _, requirement := range requirements {
		resourceTypeIDCopy := resourceTypeID
		records = append(records, models.ResourceGatherRequirement{
			ID:                      uuid.New(),
			CreatedAt:               now,
			UpdatedAt:               now,
			ResourceTypeID:          &resourceTypeIDCopy,
			MinLevel:                requirement.MinLevel,
			MaxLevel:                requirement.MaxLevel,
			RequiredInventoryItemID: requirement.RequiredInventoryItemID,
		})
	}

	return tx.Create(&records).Error
}
