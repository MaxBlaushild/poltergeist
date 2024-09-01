package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditItem struct {
	ID        uuid.UUID `json:"id"`
	MatchID   uuid.UUID `json:"matchId"`
	Match     Match     `json:"match"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Message   string    `json:"message"`
}
