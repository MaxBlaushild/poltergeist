package models

import "gorm.io/gorm"

type TextVerificationCode struct {
	gorm.Model
	PhoneNumber string
	Code        string
	Used        bool
}
