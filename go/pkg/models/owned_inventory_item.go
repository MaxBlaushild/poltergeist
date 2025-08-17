package models

import (
	"time"

	"github.com/google/uuid"
)

type OwnedInventoryItem struct {
	ID              uuid.UUID  `json:"id"`
	TeamID          *uuid.UUID `json:"teamId"`
	Team            *Team      `json:"team"`
	UserID          *uuid.UUID `json:"userId"`
	User            *User      `json:"user"`
	InventoryItemID uuid.UUID  `json:"inventoryItemId"`
	Quantity        int        `json:"quantity"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (o *OwnedInventoryItem) IsTeamItem() bool {
	return o.TeamID != nil
}

func (o *OwnedInventoryItem) IsUserItem() bool {
	return o.UserID != nil
}
