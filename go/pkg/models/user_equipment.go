package models

import (
	"time"

	"github.com/google/uuid"
)

type UserEquipment struct {
	ID                    uuid.UUID  `json:"id"`
	UserID                uuid.UUID  `json:"userId"`
	User                  *User      `json:"user"`
	EquipmentSlot         string     `json:"equipmentSlot"`
	OwnedInventoryItemID  uuid.UUID  `json:"ownedInventoryItemId"`
	OwnedInventoryItem    *OwnedInventoryItem `json:"ownedInventoryItem"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}