package models

import "gorm.io/gorm"

type HowManyAnswer struct {
	gorm.Model
	HowManyQuestion   HowManyQuestion `json:"howManyQuestion"`
	HowManyQuestionID uint            `json:"howManyQuestionId"`
	Answer            int             `json:"answer"`
	Guess             int             `json:"guess"`
	OffBy             int             `json:"offBy"`
	Correctness       float64         `json:"correctness"`
	User              User            `json:"user"`
	UserID            *uint           `json:"userId"`
	EphemeralUserID   *string         `json:"ephemeralUserId"`
}

func (h *HowManyAnswer) TableName() string {
	return "how_many_as"
}
