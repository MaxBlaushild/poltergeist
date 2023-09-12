package models

import "gorm.io/gorm"

type Challenge struct {
	gorm.Model
	AuthUser   User
	AuthUserID uint
	Challenge  string
}
