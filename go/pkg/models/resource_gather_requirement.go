package models

import (
	"time"

	"github.com/google/uuid"
)

type ResourceGatherRequirement struct {
	ID                      uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt               time.Time      `json:"createdAt"`
	UpdatedAt               time.Time      `json:"updatedAt"`
	ResourceID              *uuid.UUID     `json:"resourceId,omitempty" gorm:"-"`
	ResourceTypeID          *uuid.UUID     `json:"resourceTypeId,omitempty" gorm:"column:resource_type_id"`
	MinLevel                int            `json:"minLevel" gorm:"column:min_level"`
	MaxLevel                int            `json:"maxLevel" gorm:"column:max_level"`
	RequiredInventoryItemID int            `json:"requiredInventoryItemId" gorm:"column:required_inventory_item_id"`
	RequiredInventoryItem   *InventoryItem `json:"requiredInventoryItem,omitempty" gorm:"foreignKey:RequiredInventoryItemID"`
}

func (ResourceGatherRequirement) TableName() string {
	return "resource_gather_requirements"
}
