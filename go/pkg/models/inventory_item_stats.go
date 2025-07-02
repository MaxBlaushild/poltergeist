package models

import (
	"time"

	"github.com/google/uuid"
)

type InventoryItemStats struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	InventoryItemID  int       `json:"inventoryItemId" gorm:"not null"`
	StrengthBonus    int       `json:"strengthBonus" gorm:"default:0"`
	DexterityBonus   int       `json:"dexterityBonus" gorm:"default:0"`
	ConstitutionBonus int      `json:"constitutionBonus" gorm:"default:0"`
	IntelligenceBonus int      `json:"intelligenceBonus" gorm:"default:0"`
	WisdomBonus      int       `json:"wisdomBonus" gorm:"default:0"`
	CharismaBonus    int       `json:"charismaBonus" gorm:"default:0"`

	// Foreign key relationship
	InventoryItem *InventoryItem `json:"inventoryItem,omitempty" gorm:"foreignKey:InventoryItemID"`
}

func (InventoryItemStats) TableName() string {
	return "inventory_item_stats"
}

// GetTotalStatBonus returns the total bonus for a specific stat
func (iis *InventoryItemStats) GetTotalStatBonus(statName string) int {
	switch statName {
	case "strength":
		return iis.StrengthBonus
	case "dexterity":
		return iis.DexterityBonus
	case "constitution":
		return iis.ConstitutionBonus
	case "intelligence":
		return iis.IntelligenceBonus
	case "wisdom":
		return iis.WisdomBonus
	case "charisma":
		return iis.CharismaBonus
	default:
		return 0
	}
}

// GetAllStatBonuses returns a map of all stat bonuses
func (iis *InventoryItemStats) GetAllStatBonuses() map[string]int {
	return map[string]int{
		"strength":     iis.StrengthBonus,
		"dexterity":    iis.DexterityBonus,
		"constitution": iis.ConstitutionBonus,
		"intelligence": iis.IntelligenceBonus,
		"wisdom":       iis.WisdomBonus,
		"charisma":     iis.CharismaBonus,
	}
}

// HasAnyStatBonuses returns true if the item provides any stat bonuses
func (iis *InventoryItemStats) HasAnyStatBonuses() bool {
	return iis.StrengthBonus != 0 || iis.DexterityBonus != 0 || iis.ConstitutionBonus != 0 ||
		   iis.IntelligenceBonus != 0 || iis.WisdomBonus != 0 || iis.CharismaBonus != 0
}