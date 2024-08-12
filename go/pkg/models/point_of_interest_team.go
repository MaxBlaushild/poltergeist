package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestTeam struct {
	ID                 uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
	TeamID             uuid.UUID       `json:"teamId"`
	Team               Team            `json:"team"`
	PointOfInterestID  uuid.UUID       `json:"pointOfInterestId"`
	PointOfInterest    PointOfInterest `json:"pointOfInterest"`
	Unlocked           bool            `json:"unlocked"`
	FirstTierCaptured  bool            `json:"firstTierCaptured"`
	SecondTierCaptured bool            `json:"secondTierCaptured"`
	ThirdTierCaptured  bool            `json:"thirdTierCaptured"`
}
