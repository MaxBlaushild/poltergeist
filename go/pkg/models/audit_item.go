package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditItem struct {
	ID        uuid.UUID `json:"id"`
	MatchID   uuid.UUID `json:"matchId"`
	Match     Match     `json:"match"`
	CreatedAT time.Time `json:"createdAt"`
	UpdatedAT time.Time `json:"updatedAt"`
	Message   string    `json:"message"`
}
