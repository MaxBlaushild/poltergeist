package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestChallengeSubmission struct {
	ID                         uuid.UUID                `json:"id"`
	PointOfInterestChallengeID uuid.UUID                `json:"pointOfInterestChallengeId"`
	PointOfInterestChallenge   PointOfInterestChallenge `json:"pointOfInterestChallenge"`
	TeamID                     *uuid.UUID               `json:"teamId"`
	Team                       *Team                    `json:"team"`
	UserID                     *uuid.UUID               `json:"userId"`
	User                       *User                    `json:"user"`
	CreatedAt                  time.Time                `json:"createdAt"`
	UpdatedAt                  time.Time                `json:"updatedAt"`
	Text                       string                   `json:"question"`
	ImageURL                   string                   `json:"imageUrl"`
	IsCorrect                  *bool                    `json:"isCorrect"`
}
