package models

import "gorm.io/gorm"

type Score struct {
	gorm.Model
	Username string `gorm:"unique;index"`
	Score    int
}
