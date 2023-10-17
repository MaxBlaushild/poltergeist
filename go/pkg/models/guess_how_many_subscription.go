package models

import "gorm.io/gorm"

type GuessHowManySubscription struct {
	gorm.Model
	User             User `json:"user"`
	UserID           uint `json:"userId"`
	Subscribed       bool `gorm:"default:false" json:"subscribed"`
	NumFreeQuestions uint `gorm:"default:0" json:"numFreeQuestions"`
}
