package models

import "time"

type InventoryItem struct {
	ID            int       `gorm:"primary_key" json:"id"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Name          string    `json:"name"`
	ImageURL      string    `gorm:"column:image_url" json:"imageUrl"`
	FlavorText    string    `gorm:"column:flavor_text" json:"flavorText"`
	EffectText    string    `gorm:"column:effect_text" json:"effectText"`
	RarityTier    string    `gorm:"column:rarity_tier" json:"rarityTier"`
	IsCaptureType bool      `gorm:"column:is_capture_type" json:"isCaptureType"`
	ItemType      string    `gorm:"column:item_type" json:"itemType"`
	EquipmentSlot *string   `gorm:"column:equipment_slot" json:"equipmentSlot,omitempty"`

	// Relationship to stats
	Stats *InventoryItemStats `json:"stats,omitempty" gorm:"foreignKey:InventoryItemID"`
}

func (InventoryItem) TableName() string {
	return "inventory_items"
}