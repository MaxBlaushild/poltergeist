package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestChildren struct {
	ID                               uuid.UUID                  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PointOfInterestGroupMemberID     uuid.UUID                  `json:"pointOfInterestGroupMemberId"`
	PointOfInterestGroupMember       PointOfInterestGroupMember `json:"pointOfInterestGroupMember" gorm:"foreignKey:PointOfInterestGroupMemberID"`
	CreatedAt                        time.Time                  `json:"createdAt"`
	UpdatedAt                        time.Time                  `json:"updatedAt"`
	PointOfInterestChallengeID       uuid.UUID                  `json:"pointOfInterestChallengeId"`
	PointOfInterestChallenge         PointOfInterestChallenge   `json:"pointOfInterestChallenge" gorm:"foreignKey:PointOfInterestChallengeID"`
	NextPointOfInterestGroupMemberID uuid.UUID                  `json:"nextPointOfInterestGroupMemberId"`
	NextPointOfInterestGroupMember   PointOfInterestGroupMember `json:"nextPointOfInterestGroupMember" gorm:"foreignKey:NextPointOfInterestGroupMemberID"`
}

func (PointOfInterestChildren) TableName() string {
	return "point_of_interest_children"
}
