package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	BgiConfigurationStatusPending  = "pending"
	BgiConfigurationStatusValid    = "valid"
	BgiConfigurationStatusRejected = "rejected"
)

// BgiConfiguration is one visitor's parameter selection for a tray-set
// product — sleeve/box/color choices, not raw tray geometry. ConfigHash is
// nil until set-assembly has resolved it into a BgiSetResolution.
type BgiConfiguration struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	ProductID       uuid.UUID      `json:"productId" gorm:"type:uuid;index"`
	Params          datatypes.JSON `json:"params"`
	ConfigHash      *string        `json:"configHash" gorm:"column:config_hash;index"`
	Status          string         `json:"status"`
	RejectionReason string         `json:"rejectionReason" gorm:"column:rejection_reason"`
	PriceCents      *int64         `json:"priceCents" gorm:"column:price_cents"`
	SessionID       string         `json:"sessionId" gorm:"column:session_id"`
}

func (BgiConfiguration) TableName() string {
	return "bgi_configurations"
}

// BgiSetResolution is R-4.3's config_hash cache: identical
// (game, expansions, sleeve, box, color) combinations resolve to one
// assembled recipe (ResolvedTrays) without re-running set-assembly. Individual
// trays still cache independently by their own geometry_hash in
// bgi_tray_slice_results, so two different config_hashes can share the same
// underlying tray geometry — see go/pkg/reef/set.
type BgiSetResolution struct {
	ConfigHash            string         `json:"configHash" gorm:"column:config_hash;primaryKey"`
	CreatedAt             time.Time      `json:"createdAt"`
	ProductID             uuid.UUID      `json:"productId" gorm:"type:uuid;index"`
	GameID                uuid.UUID      `json:"gameId" gorm:"type:uuid"`
	ExpansionIDs          datatypes.JSON `json:"expansionIds" gorm:"column:expansion_ids"`
	SleeveProfileID       uuid.UUID      `json:"sleeveProfileId" gorm:"type:uuid;column:sleeve_profile_id"`
	BoxProfileID          uuid.UUID      `json:"boxProfileId" gorm:"type:uuid;column:box_profile_id"`
	ResolvedTrays         datatypes.JSON `json:"resolvedTrays" gorm:"column:resolved_trays"`
	UnassembledComponents datatypes.JSON `json:"unassembledComponents" gorm:"column:unassembled_components"`
	AssembledHeightMm     *float64       `json:"assembledHeightMm" gorm:"column:assembled_height_mm"`
	FitsBox               *bool          `json:"fitsBox" gorm:"column:fits_box"`
}

func (BgiSetResolution) TableName() string {
	return "bgi_set_resolutions"
}

const (
	BgiGenerationJobKindFullSet = "full_set"

	BgiGenerationJobStatusQueued    = "queued"
	BgiGenerationJobStatusRunning   = "running"
	BgiGenerationJobStatusCompleted = "completed"
	BgiGenerationJobStatusFailed    = "failed"
)

// BgiGenerationJob is a status/audit record the API polls for job progress —
// mirrors ReefGenerationJob exactly. Execution is dispatched via the shared
// asynq queue (go/pkg/jobs, go/job-runner).
type BgiGenerationJob struct {
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

func (BgiGenerationJob) TableName() string {
	return "bgi_generation_jobs"
}
