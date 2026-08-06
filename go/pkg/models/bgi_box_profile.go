package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	BgiBoxProfileSourceOriginal    = "original"
	BgiBoxProfileSourceAftermarket = "aftermarket"
)

// BgiBoxProfile carries the master constraint of R-2.1: every downstream fit
// check compares an assembled tray-set height against InteriorDepthMm.
// Structurally mirrors ReefTankProfile's verified/source_url pattern.
// DepthIsPlaceholder flags rows (like Terraforming Mars's seed) where the
// depth figure specifically has no independent source, even when
// length/width do — depth is usually the harder dimension to find published.
type BgiBoxProfile struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	GameID             uuid.UUID `json:"gameId" gorm:"type:uuid;index"`
	Slug               string    `json:"slug"`
	Label              string    `json:"label"`
	Source             string    `json:"source"`
	InteriorLengthMm   float64   `json:"interiorLengthMm" gorm:"column:interior_length_mm"`
	InteriorWidthMm    float64   `json:"interiorWidthMm" gorm:"column:interior_width_mm"`
	InteriorDepthMm    float64   `json:"interiorDepthMm" gorm:"column:interior_depth_mm"`
	DepthIsPlaceholder bool      `json:"depthIsPlaceholder" gorm:"column:depth_is_placeholder"`
	Verified           bool      `json:"verified"`
	SourceURL          string    `json:"sourceUrl" gorm:"column:source_url"`
	MeasurementNotes   string    `json:"measurementNotes" gorm:"column:measurement_notes"`
}

func (BgiBoxProfile) TableName() string {
	return "bgi_box_profiles"
}
