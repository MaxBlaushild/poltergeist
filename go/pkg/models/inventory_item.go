package models

import (
	"time"

	"github.com/google/uuid"
)

type InventoryItem struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	InventoryItemID int       `json:"inventoryItemId" gorm:"uniqueIndex;not null"`
	Name            string    `json:"name" gorm:"not null"`
	ImageURL        string    `json:"imageUrl"`
	FlavorText      string    `json:"flavorText"`
	EffectText      string    `json:"effectText"`
	RarityTier      string    `json:"rarityTier" gorm:"not null"`
	IsCaptureType   bool      `json:"isCaptureType" gorm:"not null;default:false"`
	ItemType        string    `json:"itemType" gorm:"not null"`
	EquipmentSlot   *string   `json:"equipmentSlot,omitempty"`
}