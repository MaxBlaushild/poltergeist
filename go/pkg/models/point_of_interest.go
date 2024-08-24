package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterest struct {
	ID                        uuid.UUID                  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                 time.Time                  `json:"createdAt"`
	UpdatedAt                 time.Time                  `json:"updatedAt"`
	Name                      string                     `json:"name"`
	Clue                      string                     `json:"clue"`
	Lat                       string                     `json:"lat"`
	Lng                       string                     `json:"lng"`
	ImageUrl                  string                     `json:"imageURL"`
	Description               string                     `json:"description"`
	PointOfInterestChallenges []PointOfInterestChallenge `json:"pointOfInterestChallenges"`
}

func (p *PointOfInterest) TableName() string {
	return "points_of_interest"
}
