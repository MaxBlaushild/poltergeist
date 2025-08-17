package models

import (
	"time"

	"github.com/google/uuid"
)

type UserEquipment struct {
	ID                       uuid.UUID  `json:"id"`
	UserID                   uuid.UUID  `json:"userId"`
	User                     *User      `json:"user"`
	HelmInventoryItemID      *uuid.UUID `json:"helmInventoryItemId"`
	ChestInventoryItemID     *uuid.UUID `json:"chestInventoryItemId"`
	LeftHandInventoryItemID  *uuid.UUID `json:"leftHandInventoryItemId"`
	RightHandInventoryItemID *uuid.UUID `json:"rightHandInventoryItemId"`
	FeetInventoryItemID      *uuid.UUID `json:"feetInventoryItemId"`
	GlovesInventoryItemID    *uuid.UUID `json:"glovesInventoryItemId"`
	NeckInventoryItemID      *uuid.UUID `json:"neckInventoryItemId"`
	LeftRingInventoryItemID  *uuid.UUID `json:"leftRingInventoryItemId"`
	RightRingInventoryItemID *uuid.UUID `json:"rightRingInventoryItemId"`
	LegInventoryItemID       *uuid.UUID `json:"legInventoryItemId"`
	CreatedAt                time.Time  `json:"createdAt"`
	UpdatedAt                time.Time  `json:"updatedAt"`
}
