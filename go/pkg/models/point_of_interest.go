package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterest struct {
	ID               uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	Name             string    `json:"name"`
	Clue             string    `json:"clue"`
	CaptureChallenge string    `json:"captureChallenge"`
	AttuneChallenge  string    `json:"attuneChallenge"`
	Lat              string    `json:"lat"`
	Lng              string    `json:"lng"`
	ImageUrl         string    `json:"imageURL"`
}

func (p *PointOfInterest) TableName() string {
	return "points_of_interest"
}
