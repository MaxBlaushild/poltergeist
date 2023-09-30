package models

import "gorm.io/gorm"

type CrystalUnlocking struct {
	gorm.Model
	TeamID    uint
	Team      Team
	CrystalID uint
	Crystal   Crystal
}

func (u *CrystalUnlocking) TableName() string {
	return "crisis_crystal_unlockings"
}
