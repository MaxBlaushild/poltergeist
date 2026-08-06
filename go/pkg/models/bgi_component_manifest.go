package models

import (
	"time"

	"github.com/google/uuid"
)

// BgiComponentManifest is R-2.4's "facts, not designs" — one row per
// component type a game (or expansion) ships with. ExpansionID nil means
// the row belongs to the base game.
type BgiComponentManifest struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	GameID        uuid.UUID  `json:"gameId" gorm:"type:uuid;index"`
	ExpansionID   *uuid.UUID `json:"expansionId" gorm:"type:uuid;column:expansion_id"`
	ComponentType string     `json:"componentType" gorm:"column:component_type"`
	CardWidthMm   *float64   `json:"cardWidthMm" gorm:"column:card_width_mm"`
	CardHeightMm  *float64   `json:"cardHeightMm" gorm:"column:card_height_mm"`
	Count         int        `json:"count"`
	Verified      bool       `json:"verified"`
	SourceURL     string     `json:"sourceUrl" gorm:"column:source_url"`
	Notes         string     `json:"notes"`
}

func (BgiComponentManifest) TableName() string {
	return "bgi_component_manifests"
}
