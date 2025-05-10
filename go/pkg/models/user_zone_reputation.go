package models

import (
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserZoneReputationName string

const (
	UserZoneReputationNameNeutral   UserZoneReputationName = "neutral"
	UserZoneReputationNameFriendly  UserZoneReputationName = "friendly"
	UserZoneReputationNameHonored   UserZoneReputationName = "honored"
	UserZoneReputationNameRevered   UserZoneReputationName = "revered"
	UserZoneReputationNameExalted   UserZoneReputationName = "exalted"
	UserZoneReputationNameLegendary UserZoneReputationName = "legendary"

	BaseReputationPoints   int = 100
	ReputationGrowthFactor int = 3
)

var UserZoneReputationNames = []UserZoneReputationName{
	UserZoneReputationNameNeutral,
	UserZoneReputationNameFriendly,
	UserZoneReputationNameHonored,
	UserZoneReputationNameRevered,
	UserZoneReputationNameExalted,
	UserZoneReputationNameLegendary,
}

type UserZoneReputation struct {
	ID                    uuid.UUID              `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt             time.Time              `json:"createdAt"`
	UpdatedAt             time.Time              `json:"updatedAt"`
	UserID                uuid.UUID              `json:"userId"`
	ZoneID                uuid.UUID              `json:"zoneId"`
	Level                 int                    `json:"level"`
	TotalReputation       int                    `json:"totalReputation"`
	ReputationOnLevel     int                    `json:"reputationOnLevel"`
	LevelsGained          int                    `json:"levelsGained" gorm:"-"`
	Name                  UserZoneReputationName `json:"name" gorm:"-"`
	ReputationToNextLevel int                    `json:"reputationToNextLevel" gorm:"-"`
}

func (u *UserZoneReputation) TableName() string {
	return "user_zone_reputations"
}

func (u *UserZoneReputation) AfterFind(tx *gorm.DB) (err error) {
	u.Name = u.GetReputationName()
	u.ReputationToNextLevel = u.GetReputationToNextLevel()
	return
}

func (u *UserZoneReputation) AddReputationPoints(points int) {
	u.TotalReputation += points
	u.ReputationOnLevel += points
	if u.ReputationOnLevel >= u.GetReputationToNextLevel() && u.Level < 6 {
		u.Level++
		u.LevelsGained++
		u.ReputationOnLevel = u.TotalReputation - u.GetReputationToNextLevel()
	}
}

func (u *UserZoneReputation) GetReputationToNextLevel() int {
	return BaseReputationPoints * int(math.Pow(float64(GrowthFactor), float64(u.Level)))
}

func (u *UserZoneReputation) GetReputationName() UserZoneReputationName {
	return UserZoneReputationNames[u.Level-1]
}
