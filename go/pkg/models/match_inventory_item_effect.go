package models

import (
	"time"

	"github.com/google/uuid"
)

type MatchInventoryItemEffect struct {
	ID              uuid.UUID     `json:"id"`
	MatchID         uuid.UUID     `json:"matchId"`
	InventoryItemID uuid.UUID     `json:"inventoryItemId"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	ExpiresAt       time.Time     `json:"expiresAt"`
	InventoryItem   InventoryItem `json:"inventoryItem"`
}
