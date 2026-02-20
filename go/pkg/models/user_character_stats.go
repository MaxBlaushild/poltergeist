package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	CharacterStatBaseValue     = 10
	CharacterStatPointsPerLevel = 5
)

type UserCharacterStats struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	UserID          uuid.UUID `json:"userId"`
	Strength        int       `json:"strength" gorm:"default:10"`
	Dexterity       int       `json:"dexterity" gorm:"default:10"`
	Constitution    int       `json:"constitution" gorm:"default:10"`
	Intelligence    int       `json:"intelligence" gorm:"default:10"`
	Wisdom          int       `json:"wisdom" gorm:"default:10"`
	Charisma        int       `json:"charisma" gorm:"default:10"`
	UnspentPoints   int       `json:"unspentPoints" gorm:"default:0"`
	LastLevelAwarded int      `json:"lastLevelAwarded" gorm:"default:1"`
}

func (u *UserCharacterStats) TableName() string {
	return "user_character_stats"
}
