package models

import (
	"time"

	"github.com/google/uuid"
)

type MatchPointOfInterest struct {
	ID                uuid.UUID       `json:"id"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	MatchID           uuid.UUID       `json:"matchId"`
	Match             Match           `json:"match"`
	PointOfInterestID uuid.UUID       `json:"pointOfInterestId"`
	PointOfInterest   PointOfInterest `json:"pointOfInterest"`
}
