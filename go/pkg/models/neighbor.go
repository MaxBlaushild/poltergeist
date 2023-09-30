package models

import "gorm.io/gorm"

type Neighbor struct {
	gorm.Model
	CrystalOneID uint    `json:"crystalOneId" binding:"required"`
	CrystalOne   Crystal `json:"crystalOne"`
	CrystalTwoID uint    `json:"crystalTwoId" binding:"required"`
	CrystalTwo   Crystal `json:"crystalTwo"`
}

func (u *Neighbor) TableName() string {
	return "crystal_neighbors"
}
