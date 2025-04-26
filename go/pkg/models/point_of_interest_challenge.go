package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestChallenge struct {
	ID                                  uuid.UUID                            `json:"id"`
	PointOfInterestID                   uuid.UUID                            `json:"pointOfInterestId"`
	PointOfInterest                     PointOfInterest                      `json:"pointOfInterest"`
	PointOfInterestGroupID              *uuid.UUID                           `json:"pointOfInterestGroupId"`
	PointOfInterestGroup                *PointOfInterestGroup                `json:"pointOfInterestGroup"`
	Question                            string                               `json:"question"`
	Tier                                int                                  `json:"tier"`
	CreatedAt                           time.Time                            `json:"created_at"`
	UpdatedAt                           time.Time                            `json:"updated_at"`
	PointOfInterestChallengeSubmissions []PointOfInterestChallengeSubmission `json:"pointOfInterestChallengeSubmissions"`
	InventoryItemID                     int                                  `json:"inventoryItemId"`
}
