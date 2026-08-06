package models

import (
	"time"

	"github.com/google/uuid"
)

// BgiSleeveProfile is R-2.2's named sleeve class. TotalCardThicknessMm is
// the one tested place the "a sleeve wraps both faces" arithmetic lives —
// callers must use it rather than reading BaseCardThicknessMm directly.
type BgiSleeveProfile struct {
	ID                        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt                 time.Time `json:"createdAt"`
	UpdatedAt                 time.Time `json:"updatedAt"`
	ClassKey                  string    `json:"classKey" gorm:"column:class_key;uniqueIndex"`
	Label                     string    `json:"label"`
	BaseCardThicknessMm       float64   `json:"baseCardThicknessMm" gorm:"column:base_card_thickness_mm"`
	SleeveMaterialThicknessMm float64   `json:"sleeveMaterialThicknessMm" gorm:"column:sleeve_material_thickness_mm"`
	Verified                  bool      `json:"verified"`
	SourceURL                 string    `json:"sourceUrl" gorm:"column:source_url"`
}

func (BgiSleeveProfile) TableName() string {
	return "bgi_sleeve_profiles"
}

// TotalCardThicknessMm is the real per-card stack thickness once sleeved: the
// card itself plus a sleeve-film layer on both faces. R-2.2: "sleeved cards
// are dramatically thicker than unsleeved" — this is the exact arithmetic
// that must be right, or card wells come out too shallow.
func (p BgiSleeveProfile) TotalCardThicknessMm() float64 {
	return p.BaseCardThicknessMm + 2*p.SleeveMaterialThicknessMm
}
