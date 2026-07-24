package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ReefTankProfile seeds verified rim/glass dimensions (R-3.4). Rows without a
// real source_url must be excluded from configurator dropdowns by callers
// (WHERE verified = true) — the DB enforces verified rows carry a source_url
// via a CHECK constraint, but does not enforce the reverse.
type ReefTankProfile struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	Manufacturer     string         `json:"manufacturer"`
	Model            string         `json:"model"`
	RimThicknessMm   float64        `json:"rimThicknessMm" gorm:"column:rim_thickness_mm"`
	RimWidthMm       float64        `json:"rimWidthMm" gorm:"column:rim_width_mm"`
	GlassThicknessMm float64        `json:"glassThicknessMm" gorm:"column:glass_thickness_mm"`
	EuroBrace        bool           `json:"euroBrace" gorm:"column:euro_brace"`
	InternalDims     datatypes.JSON `json:"internalDims" gorm:"column:internal_dims"`
	Verified         bool           `json:"verified"`
	SourceURL        string         `json:"sourceUrl" gorm:"column:source_url"`
}

func (ReefTankProfile) TableName() string {
	return "reef_tank_profiles"
}
