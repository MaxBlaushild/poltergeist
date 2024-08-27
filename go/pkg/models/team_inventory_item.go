package models

import (
	"time"

	"github.com/google/uuid"
)

type TeamInventoryItem struct {
	ID              uuid.UUID `json:"id"`
	TeamID          uuid.UUID `json:"teamId"`
	InventoryItemID int       `json:"inventoryItemId"`
	Quantity        int       `json:"quantity"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
