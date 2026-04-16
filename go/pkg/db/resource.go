package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type resourceHandle struct {
	db *gorm.DB
}

func (h *resourceHandle) preload(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Zone").
		Preload("ResourceType").
		Preload("GatherRequirements", func(db *gorm.DB) *gorm.DB {
			return db.Order("min_level ASC, max_level ASC, required_inventory_item_id ASC")
		}).
		Preload("GatherRequirements.RequiredInventoryItem")
}

func (h *resourceHandle) Create(ctx context.Context, resource *models.Resource) error {
	resource.ID = uuid.New()
	resource.CreatedAt = time.Now()
	resource.UpdatedAt = resource.CreatedAt
	if err := resource.SetGeometry(resource.Latitude, resource.Longitude); err != nil {
		return err
	}
	requirements := resource.GatherRequirements
	resource.GatherRequirements = nil
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("GatherRequirements").Create(resource).Error; err != nil {
			return err
		}
		return syncResourceGatherRequirementsTx(tx, resource.ID, requirements)
	})
}

func (h *resourceHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Resource, error) {
	var resource models.Resource
	if err := h.preload(h.db.WithContext(ctx)).First(&resource, id).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

func (h *resourceHandle) FindAll(ctx context.Context) ([]models.Resource, error) {
	var resources []models.Resource
	if err := h.preload(h.db.WithContext(ctx)).
		Order("created_at ASC").
		Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

func (h *resourceHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Resource, error) {
	var resources []models.Resource
	if err := h.preload(h.db.WithContext(ctx)).
		Where("zone_id = ? AND invalidated = false", zoneID).
		Order("created_at ASC").
		Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

func (h *resourceHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Resource) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	if updates.Latitude != 0 || updates.Longitude != 0 {
		if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
			return err
		}
	}
	payload := map[string]interface{}{
		"zone_id":          updates.ZoneID,
		"resource_type_id": updates.ResourceTypeID,
		"quantity":         updates.Quantity,
		"latitude":         updates.Latitude,
		"longitude":        updates.Longitude,
		"geometry":         updates.Geometry,
		"invalidated":      updates.Invalidated,
		"updated_at":       updates.UpdatedAt,
	}
	requirements := updates.GatherRequirements
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&models.Resource{}).
			Where("id = ?", id).
			Updates(payload).Error; err != nil {
			return err
		}
		return syncResourceGatherRequirementsTx(tx, id, requirements)
	})
}

func (h *resourceHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Resource{}, "id = ?", id).Error
}

func (h *resourceHandle) HasUserGathered(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).
		Model(&models.UserResourceGathering{}).
		Where("user_id = ? AND resource_id = ?", userID, resourceID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (h *resourceHandle) CreateUserGathering(ctx context.Context, gathering *models.UserResourceGathering) error {
	gathering.ID = uuid.New()
	gathering.CreatedAt = time.Now()
	gathering.UpdatedAt = gathering.CreatedAt
	if gathering.GatheredAt.IsZero() {
		gathering.GatheredAt = gathering.CreatedAt
	}
	return h.db.WithContext(ctx).Create(gathering).Error
}

func (h *resourceHandle) FindByIDWithUserStatus(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*models.Resource, bool, error) {
	resource, err := h.FindByID(ctx, id)
	if err != nil {
		return nil, false, err
	}
	gatheredByUser := false
	if userID != nil {
		gatheredByUser, err = h.HasUserGathered(ctx, *userID, id)
		if err != nil {
			return nil, false, err
		}
	}
	return resource, gatheredByUser, nil
}

func (h *resourceHandle) FindAllWithUserStatus(ctx context.Context, userID *uuid.UUID) ([]models.Resource, map[uuid.UUID]bool, error) {
	var resources []models.Resource
	if err := h.preload(h.db.WithContext(ctx)).
		Order("created_at ASC").
		Find(&resources).Error; err != nil {
		return nil, nil, err
	}
	gatheredMap := make(map[uuid.UUID]bool)
	if userID != nil {
		var gatherRows []models.UserResourceGathering
		if err := h.db.WithContext(ctx).
			Where("user_id = ?", *userID).
			Find(&gatherRows).Error; err != nil {
			return nil, nil, err
		}
		for _, gathering := range gatherRows {
			gatheredMap[gathering.ResourceID] = true
		}
	}
	return resources, gatheredMap, nil
}

func (h *resourceHandle) FindByZoneIDWithUserStatus(ctx context.Context, zoneID uuid.UUID, userID *uuid.UUID) ([]models.Resource, map[uuid.UUID]bool, error) {
	var resources []models.Resource
	if err := h.preload(h.db.WithContext(ctx)).
		Where("zone_id = ? AND invalidated = false", zoneID).
		Order("created_at ASC").
		Find(&resources).Error; err != nil {
		return nil, nil, err
	}
	gatheredMap := make(map[uuid.UUID]bool)
	if userID != nil {
		var gatherRows []models.UserResourceGathering
		if err := h.db.WithContext(ctx).
			Where("user_id = ?", *userID).
			Find(&gatherRows).Error; err != nil {
			return nil, nil, err
		}
		for _, gathering := range gatherRows {
			gatheredMap[gathering.ResourceID] = true
		}
	}
	return resources, gatheredMap, nil
}

func syncResourceGatherRequirementsTx(
	tx *gorm.DB,
	resourceID uuid.UUID,
	requirements []models.ResourceGatherRequirement,
) error {
	if err := tx.Where("resource_id = ?", resourceID).Delete(&models.ResourceGatherRequirement{}).Error; err != nil {
		return err
	}
	if len(requirements) == 0 {
		return nil
	}

	now := time.Now()
	records := make([]models.ResourceGatherRequirement, 0, len(requirements))
	for _, requirement := range requirements {
		records = append(records, models.ResourceGatherRequirement{
			ID:                      uuid.New(),
			CreatedAt:               now,
			UpdatedAt:               now,
			ResourceID:              resourceID,
			MinLevel:                requirement.MinLevel,
			MaxLevel:                requirement.MaxLevel,
			RequiredInventoryItemID: requirement.RequiredInventoryItemID,
		})
	}

	return tx.Create(&records).Error
}
