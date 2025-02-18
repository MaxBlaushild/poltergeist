package models

import (
	"time"

	"github.com/google/uuid"
)

type MatchUser struct {
	ID        uuid.UUID `json:"id"`
	MatchID   uuid.UUID `json:"matchId"`
	UserID    uuid.UUID `json:"userId"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
