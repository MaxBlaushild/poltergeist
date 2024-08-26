package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID                   uuid.UUID `json:"id" db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt            time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt            time.Time `json:"updatedAt" db:"updated_at"`
	Name                 string    `json:"name"`
	UserTeams            []UserTeam
	Users                []User                `json:"users" gorm:"many2many:user_teams;"`
	PointOfInterestTeams []PointOfInterestTeam `json:"pointOfInterestTeams"`
	TeamInventoryItems   []TeamInventoryItem   `json:"teamInventoryItems"`
}
