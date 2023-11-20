package models

import (
	"time"

	"github.com/google/uuid"
)

type HowManyAnswer struct {
	ID                uuid.UUID       `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time       `db:"created_at"`
	UpdatedAt         time.Time       `db:"updated_at"`
	HowManyQuestion   HowManyQuestion `json:"howManyQuestion"`
	HowManyQuestionID uuid.UUID       `json:"howManyQuestionId"`
	Answer            int             `json:"answer"`
	Guess             int             `json:"guess"`
	OffBy             int             `json:"offBy"`
	Correctness       float64         `json:"correctness"`
	User              User            `json:"user"`
	UserID            uuid.UUID       `json:"userId"`
}
