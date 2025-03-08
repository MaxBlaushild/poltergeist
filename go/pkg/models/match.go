package models

import (
	"time"

	"github.com/google/uuid"
)

type Match struct {
	ID                   uuid.UUID                  `json:"id" db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt            time.Time                  `json:"createdAt" db:"created_at"`
	UpdatedAt            time.Time                  `json:"updatedAt" db:"updated_at"`
	CreatorID            uuid.UUID                  `json:"creatorId" db:"creator_id"`
	Creator              User                       `json:"creator" db:"creator" gorm:"foreignKey:CreatorID"`
	VerificationCodes    []VerificationCode         `json:"verificationCodes" gorm:"many2many:match_verification_codes;"`
	StartedAt            *time.Time                 `json:"startedAt" db:"started_at"`
	EndedAt              *time.Time                 `json:"endedAt" db:"ended_at"`
	PointsOfInterest     []PointOfInterest          `json:"pointsOfInterest" gorm:"many2many:match_points_of_interest;"`
	Teams                []Team                     `json:"teams" gorm:"many2many:team_matches;"`
	InventoryItemEffects []MatchInventoryItemEffect `json:"inventoryItemEffects"`
	Users                []User                     `json:"users" gorm:"many2many:match_users;"`
}
