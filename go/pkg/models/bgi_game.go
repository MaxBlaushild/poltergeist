package models

import (
	"time"

	"github.com/google/uuid"
)

// BgiGame.Active gates catalog visibility (same meaning as ReefProduct.Active
// — is this shown/orderable at all), independent of whether its
// measurement data has been physically verified. See
// BgiBoxProfile/BgiSleeveProfile/BgiComponentManifest's own Verified fields
// for the actual data-trust signal, which the frontend must surface
// wherever that data is used rather than gating the whole game's liveness.
type BgiGame struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Slug          string    `json:"slug" gorm:"uniqueIndex"`
	Name          string    `json:"name"`
	Publisher     string    `json:"publisher"`
	YearPublished *int      `json:"yearPublished" gorm:"column:year_published"`
	Active        bool      `json:"active"`
}

func (BgiGame) TableName() string {
	return "bgi_games"
}

type BgiExpansion struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	GameID     uuid.UUID `json:"gameId" gorm:"type:uuid;index"`
	Slug       string    `json:"slug"`
	Name       string    `json:"name"`
	Standalone bool      `json:"standalone"`
	Active     bool      `json:"active"`
}

func (BgiExpansion) TableName() string {
	return "bgi_expansions"
}
