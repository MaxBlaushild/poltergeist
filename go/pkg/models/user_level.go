package models

import (
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	BaseExperiencePoints = 100
	GrowthFactor         = 10
)

type UserLevel struct {
	ID                      uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
	UserID                  uuid.UUID `json:"userId"`
	Level                   int       `json:"level" gorm:"default:1"`
	ExperiencePointsOnLevel int       `json:"experiencePointsOnLevel" gorm:"default:0"`
	TotalExperiencePoints   int       `json:"totalExperiencePoints" gorm:"default:0"`
	LevelsGained            int       `json:"levelsGained" gorm:"-"`
	ExperienceToNextLevel   int       `json:"experienceToNextLevel" gorm:"-"`
}

func (u *UserLevel) TableName() string {
	return "user_levels"
}

func (u *UserLevel) AfterFind(tx *gorm.DB) (err error) {
	u.ExperienceToNextLevel = u.XPToNextLevel()
	return
}

func (u *UserLevel) AddExperiencePoints(points int) {
	u.TotalExperiencePoints += points
	u.ExperiencePointsOnLevel += points
	if u.ExperiencePointsOnLevel >= u.XPToNextLevel() {
		u.Level++
		u.LevelsGained++
		u.ExperiencePointsOnLevel = u.TotalExperiencePoints - u.XPToNextLevel()
	}
}

func (u *UserLevel) XPToNextLevel() int {
	return BaseExperiencePoints * int(math.Log(float64(u.Level+1))) * GrowthFactor
}
