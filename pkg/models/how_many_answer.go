package models

import "gorm.io/gorm"

type HowManyAnswer struct {
	gorm.Model
	HowManyQuestion   HowManyQuestion `json:"howManyQuestion"`
	HowManyQuestionID uint            `json:"howManyQuestionId"`
	Answer            int             `json:"answer"`
	Guess             int             `json:"guess"`
	Correctness       float64         `json:"correctness"`
	OffBy             int             `json:"offBy"`
	UserID            string          `json:"userId" gorm:"index"`
}
