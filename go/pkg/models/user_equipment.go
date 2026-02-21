package models

import (
	"time"

	"github.com/google/uuid"
)

type UserEquipment struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	UserID             uuid.UUID `json:"userId"`
	Slot               string    `json:"slot"`
	OwnedInventoryItemID uuid.UUID `json:"ownedInventoryItemId" gorm:"column:owned_inventory_item_id"`
}

func (UserEquipment) TableName() string {
	return "user_equipment"
}
