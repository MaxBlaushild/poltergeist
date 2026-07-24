package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	ReefConfigurationStatusPending  = "pending"
	ReefConfigurationStatusValid    = "valid"
	ReefConfigurationStatusRejected = "rejected"
)

// ReefConfiguration is one visitor's parameter selection for a configurable
// product. GeometryHash is nil until a slice has resolved it (R-5.1: nothing
// enters a cart without a passing server-side slice).
type ReefConfiguration struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	ProductID       uuid.UUID      `json:"productId" gorm:"type:uuid;index"`
	Params          datatypes.JSON `json:"params"`
	GeometryHash    *string        `json:"geometryHash" gorm:"column:geometry_hash;index"`
	Status          string         `json:"status"`
	RejectionReason string         `json:"rejectionReason" gorm:"column:rejection_reason"`
	PriceCents      *int64         `json:"priceCents" gorm:"column:price_cents"`
	SessionID       string         `json:"sessionId" gorm:"column:session_id"`
}

func (ReefConfiguration) TableName() string {
	return "reef_configurations"
}

const (
	ReefSliceStatusPending  = "pending"
	ReefSliceStatusValid    = "valid"
	ReefSliceStatusRejected = "rejected"
)

// ReefSliceResult is the cache keyed by geometry_hash (R-3.3): identical
// params for a product must never regenerate or re-slice. It is also the
// source of truth for weight/print-time/printability (R-2.7) and for price
// (R-6.2), since price is a pure function of a slice result.
type ReefSliceResult struct {
	GeometryHash    string         `json:"geometryHash" gorm:"column:geometry_hash;primaryKey"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	ProductID       uuid.UUID      `json:"productId" gorm:"type:uuid;index"`
	Status          string         `json:"status"`
	RejectionRule   string         `json:"rejectionRule" gorm:"column:rejection_rule"`
	RejectionReason string         `json:"rejectionReason" gorm:"column:rejection_reason"`
	WeightG         *float64       `json:"weightG" gorm:"column:weight_g"`
	PrintTimeS      *int64         `json:"printTimeS" gorm:"column:print_time_s"`
	BboxMm          datatypes.JSON `json:"bboxMm" gorm:"column:bbox_mm"`
	PlateFits       *bool          `json:"plateFits" gorm:"column:plate_fits"`
	SupportRequired *bool          `json:"supportRequired" gorm:"column:support_required"`
	MinWallMm       *float64       `json:"minWallMm" gorm:"column:min_wall_mm"`
	SealedVoid      *bool          `json:"sealedVoid" gorm:"column:sealed_void"`
	Warnings        datatypes.JSON `json:"warnings"`
	SlicerVersion   string         `json:"slicerVersion" gorm:"column:slicer_version"`
	OpenSCADVersion string         `json:"openscadVersion" gorm:"column:openscad_version"`
	STLKey          string         `json:"stlKey" gorm:"column:stl_key"`
	PreviewKey      string         `json:"previewKey" gorm:"column:preview_key"`
	PriceCents      *int64         `json:"priceCents" gorm:"column:price_cents"`
}

func (ReefSliceResult) TableName() string {
	return "reef_slice_results"
}

const (
	ReefGenerationJobKindPreview = "preview"
	ReefGenerationJobKindFull    = "full"

	ReefGenerationJobStatusQueued    = "queued"
	ReefGenerationJobStatusRunning   = "running"
	ReefGenerationJobStatusCompleted = "completed"
	ReefGenerationJobStatusFailed    = "failed"
)

// ReefGenerationJob is a status/audit record the API polls for job progress.
// Execution itself is dispatched via the repo's existing asynq queue
// (go/pkg/jobs, go/job-runner) per R-2.10 — this table tracks state, it is
// not the dispatch mechanism.
type ReefGenerationJob struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	ConfigurationID uuid.UUID  `json:"configurationId" gorm:"type:uuid;index"`
	Kind            string     `json:"kind"`
	Status          string     `json:"status"`
	Attempts        int        `json:"attempts"`
	LockedAt        *time.Time `json:"lockedAt" gorm:"column:locked_at"`
	Error           string     `json:"error"`
}

func (ReefGenerationJob) TableName() string {
	return "reef_generation_jobs"
}
