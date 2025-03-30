package models

import (
	"time"

	"github.com/google/uuid"
)

type TagEntity struct {
	ID                         uuid.UUID  `json:"id"`
	PointOfInterestID          *uuid.UUID `json:"pointOfInterestId"`
	PointOfInterestGroupID     *uuid.UUID `json:"pointOfInterestGroupId"`
	PointOfInterestChallengeID *uuid.UUID `json:"pointOfInterestChallengeId"`
	CreatedAt                  time.Time  `json:"createdAt"`
	UpdatedAt                  time.Time  `json:"updatedAt"`
}
