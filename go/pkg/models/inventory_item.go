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
	EquipSlot     *string   `json:"equipSlot" gorm:"column:equip_slot"`
	StrengthMod   int       `json:"strengthMod" gorm:"column:strength_mod"`
	DexterityMod  int       `json:"dexterityMod" gorm:"column:dexterity_mod"`
	ConstitutionMod int     `json:"constitutionMod" gorm:"column:constitution_mod"`
	IntelligenceMod int     `json:"intelligenceMod" gorm:"column:intelligence_mod"`
	WisdomMod     int       `json:"wisdomMod" gorm:"column:wisdom_mod"`
	CharismaMod   int       `json:"charismaMod" gorm:"column:charisma_mod"`
	ImageGenerationStatus string  `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError  *string `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
}

func (InventoryItem) TableName() string {
	return "inventory_items"
}

const (
	InventoryImageGenerationStatusNone       = "none"
	InventoryImageGenerationStatusQueued     = "queued"
	InventoryImageGenerationStatusInProgress = "in_progress"
	InventoryImageGenerationStatusComplete   = "complete"
	InventoryImageGenerationStatusFailed     = "failed"
)
