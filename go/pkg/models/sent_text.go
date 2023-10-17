package models

import "gorm.io/gorm"

type SentText struct {
	gorm.Model
	TextType    string `gorm:"index"`
	PhoneNumber string `gorm:"index"`
	Text        string
}
