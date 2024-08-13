package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestChallenge struct {
	ID                uuid.UUID       `json:"id"`
	PointOfInterestID uuid.UUID       `json:"pointOfInterestId"`
	PointOfInterest   PointOfInterest `json:"pointOfInterest"`
	Question          string          `json:"question"`
	Tier              int             `json:"tier"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}
