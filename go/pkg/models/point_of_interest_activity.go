package models

import "github.com/google/uuid"

type PointOfInterestActivity struct {
	ID                string          `json:"id"`
	CreatedAt         string          `json:"createdAt"`
	UpdatedAt         string          `json:"updatedAt"`
	PointOfInterestID uuid.UUID       `json:"pointOfInterestId" gorm:"foreignKey:PointOfInterestID"`
	PointOfInterest   PointOfInterest `json:"pointOfInterest"`
	SonarActivity     SonarActivity   `json:"sonarActivity" gorm:"foreignKey:SonarActivityID"`
	SonarActivityID   uuid.UUID       `json:"sonarActivityId"`
}
