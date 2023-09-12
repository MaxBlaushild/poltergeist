package models

import "gorm.io/gorm"

type Challenge struct {
	gorm.Model
	AuthUser   AuthUser
	AuthUserID uint
	Challenge  string
}
