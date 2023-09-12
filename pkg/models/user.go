package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserID             string `json:"userId" gorm:"unique;index"`
	PhoneNumber        string `json:"phoneNumber" gorm:"unique;index"`
	HowManyQuestionSub bool   `json:"howManyQuestionsSub"`
}
