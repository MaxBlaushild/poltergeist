package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestChildren struct {
	ID                           uuid.UUID                  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PointOfInterestGroupMemberID uuid.UUID                  `json:"pointOfInterestGroupMemberId"`
	PointOfInterestGroupMember   PointOfInterestGroupMember `json:"pointOfInterestGroupMember" gorm:"foreignKey:PointOfInterestGroupMemberID"`
	PointOfInterestID            uuid.UUID                  `json:"pointOfInterestId"`
	PointOfInterest              PointOfInterest            `json:"pointOfInterest" gorm:"foreignKey:PointOfInterestID"`
	CreatedAt                    time.Time                  `json:"createdAt"`
	UpdatedAt                    time.Time                  `json:"updatedAt"`
	PointOfInterestChallengeID   uuid.UUID                  `json:"pointOfInterestChallengeId"`
	PointOfInterestChallenge     PointOfInterestChallenge   `json:"pointOfInterestChallenge" gorm:"foreignKey:PointOfInterestChallengeID"`
}
