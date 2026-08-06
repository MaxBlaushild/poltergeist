package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	BgiTraySliceStatusPending  = "pending"
	BgiTraySliceStatusValid    = "valid"
	BgiTraySliceStatusRejected = "rejected"
)

// BgiTraySliceResult is a structural clone of ReefSliceResult, keyed by
// geometry_hash, FK retargeted to BgiTrayTemplate. Computed per resolved
// tray (not per set) — see go/pkg/reef/set's two-level cache design.
type BgiTraySliceResult struct {
	GeometryHash    string         `json:"geometryHash" gorm:"column:geometry_hash;primaryKey"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	TrayTemplateID  uuid.UUID      `json:"trayTemplateId" gorm:"type:uuid;column:tray_template_id;index"`
	Status          string         `json:"status"`
	RejectionRule   string         `json:"rejectionRule" gorm:"column:rejection_rule"`
	RejectionReason string         `json:"rejectionReason" gorm:"column:rejection_reason"`
	WeightG         *float64       `json:"weightG" gorm:"column:weight_g"`
	PrintTimeS      *int64         `json:"printTimeS" gorm:"column:print_time_s"`
	BboxMm          datatypes.JSON `json:"bboxMm" gorm:"column:bbox_mm"`
	PlateFits       *bool          `json:"plateFits" gorm:"column:plate_fits"`

	SupportRequired        *bool    `json:"supportRequired" gorm:"column:support_required"`
	SupportMaterialPercent *float64 `json:"supportMaterialPercent" gorm:"column:support_material_percent"`
	MinWallMm              *float64 `json:"minWallMm" gorm:"column:min_wall_mm"`
	// SealedVoid is kept for schema symmetry with the shared validate.Metadata
	// shape; bgi trays run with SealedVoidRuleEnabled=false (open-top wells
	// have no cavity story at all — see go/pkg/reef/validate).
	SealedVoid      *bool          `json:"sealedVoid" gorm:"column:sealed_void"`
	Warnings        datatypes.JSON `json:"warnings"`
	SlicerVersion   string         `json:"slicerVersion" gorm:"column:slicer_version"`
	OpenSCADVersion string         `json:"openscadVersion" gorm:"column:openscad_version"`
	STLKey          string         `json:"stlKey" gorm:"column:stl_key"`
	PreviewKey      string         `json:"previewKey" gorm:"column:preview_key"`
	PriceCents      *int64         `json:"priceCents" gorm:"column:price_cents"`
}

func (BgiTraySliceResult) TableName() string {
	return "bgi_tray_slice_results"
}
