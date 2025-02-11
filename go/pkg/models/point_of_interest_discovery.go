package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestDiscovery struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	TeamID            *uuid.UUID      `json:"teamId"`
	Team              *Team           `json:"team"`
	UserID            *uuid.UUID      `json:"userId"`
	User              *User           `json:"user"`
	PointOfInterestID uuid.UUID       `json:"pointOfInterestId"`
	PointOfInterest   PointOfInterest `json:"pointOfInterest"`
}

func (PointOfInterestDiscovery) TableName() string {
	return "point_of_interest_discoveries"
}
