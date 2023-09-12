package models

import "gorm.io/gorm"

type UserTeam struct {
	gorm.Model
	TeamID uint
	Team   Team
	UserID uint
}
