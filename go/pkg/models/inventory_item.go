package models

import "time"

type InventoryItem struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	CreatedAt     time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"column:updated_at"`
	Name          string    `json:"name" gorm:"column:name"`
	ImageURL      string    `json:"imageUrl" gorm:"column:image_url"`
	FlavorText    string    `json:"flavorText" gorm:"column:flavor_text"`
	EffectText    string    `json:"effectText" gorm:"column:effect_text"`
	RarityTier    string    `json:"rarityTier" gorm:"column:rarity_tier"`
	IsCaptureType bool      `json:"isCaptureType" gorm:"column:is_capture_type"`
	SellValue     *int      `json:"sellValue" gorm:"column:sell_value"`
	UnlockTier    *int      `json:"unlockTier" gorm:"column:unlock_tier"`
}

func (InventoryItem) TableName() string {
	return "inventory_items"
}
