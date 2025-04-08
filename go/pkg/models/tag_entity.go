package models

import (
	"time"

	"github.com/google/uuid"
)

type TagEntity struct {
	ID                         uuid.UUID                 `json:"id"`
	PointOfInterestID          *uuid.UUID                `json:"pointOfInterestId"`
	PointOfInterestGroupID     *uuid.UUID                `json:"pointOfInterestGroupId"`
	PointOfInterestChallengeID *uuid.UUID                `json:"pointOfInterestChallengeId"`
	PointOfInterest            *PointOfInterest          `json:"pointOfInterest"`
	PointOfInterestGroup       *PointOfInterestGroup     `json:"pointOfInterestGroup"`
	PointOfInterestChallenge   *PointOfInterestChallenge `json:"pointOfInterestChallenge"`
	Tag                        Tag                       `json:"tag"`
	TagID                      uuid.UUID                 `json:"tagId"`
	CreatedAt                  time.Time                 `json:"createdAt"`
	UpdatedAt                  time.Time                 `json:"updatedAt"`
}
