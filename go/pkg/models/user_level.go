package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	BaseExperiencePoints      = 100
	LinearExperienceGrowth    = 60
	QuadraticExperienceGrowth = 20
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
	if points <= 0 {
		u.ExperienceToNextLevel = u.XPToNextLevel()
		return
	}

	u.LevelsGained = 0
	u.TotalExperiencePoints += points
	u.ExperiencePointsOnLevel += points

	for u.ExperiencePointsOnLevel >= u.XPToNextLevel() {
		u.ExperiencePointsOnLevel -= u.XPToNextLevel()
		u.Level++
		u.LevelsGained++
	}

	u.ExperienceToNextLevel = u.XPToNextLevel()
}

func (u *UserLevel) XPToNextLevel() int {
	if u.Level <= 1 {
		return BaseExperiencePoints
	}

	levelOffset := u.Level - 1

	// Quadratic growth keeps early levels brisk while making later levels
	// meaningfully harder instead of flattening into long plateaus.
	return BaseExperiencePoints +
		(levelOffset * LinearExperienceGrowth) +
		(levelOffset * levelOffset * QuadraticExperienceGrowth)
}
