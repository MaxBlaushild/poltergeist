package models

import (
	"time"

	"github.com/google/uuid"
)

type MatchInventoryItemEffect struct {
	ID              uuid.UUID `json:"id"`
	MatchID         uuid.UUID `json:"matchId"`
	TeamID          uuid.UUID `json:"teamId"`
	Team            Team      `json:"team"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	ExpiresAt       time.Time `json:"expiresAt"`
	InventoryItemID int       `json:"inventoryItemId"`
}

func (m MatchInventoryItemEffect) IsExpired() bool {
	return m.ExpiresAt.Before(time.Now())
}
