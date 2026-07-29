package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type reefConfigurationHandle struct {
	db *gorm.DB
}

func (h *reefConfigurationHandle) Create(ctx context.Context, cfg *models.ReefConfiguration) (*models.ReefConfiguration, error) {
	if cfg.ID == uuid.Nil {
		cfg.ID = uuid.New()
	}
	if err := h.db.WithContext(ctx).Create(cfg).Error; err != nil {
		return nil, err
	}
	return cfg, nil
}

func (h *reefConfigurationHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ReefConfiguration, error) {
	var cfg models.ReefConfiguration
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (h *reefConfigurationHandle) Update(ctx context.Context, cfg *models.ReefConfiguration) error {
	return h.db.WithContext(ctx).Save(cfg).Error
}

// CountByStatusSince backs R-9.2's validation rejection rate: rejected /
// (valid + rejected) over configurations that actually reached a server-side
// slice, per product since a given time.
func (h *reefConfigurationHandle) CountByStatusSince(ctx context.Context, status string, since time.Time) (int64, error) {
	var count int64
	err := h.db.WithContext(ctx).Model(&models.ReefConfiguration{}).
		Where("status = ? AND created_at >= ?", status, since).
		Count(&count).Error
	return count, err
}

type reefSliceResultHandle struct {
	db *gorm.DB
}

func (h *reefSliceResultHandle) FindByGeometryHash(ctx context.Context, geometryHash string) (*models.ReefSliceResult, error) {
	var result models.ReefSliceResult
	err := h.db.WithContext(ctx).Where("geometry_hash = ?", geometryHash).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create inserts a new cache row, doing nothing if the geometry_hash already
// exists (R-3.3: identical inputs must never regenerate or re-slice — a
// concurrent request that lost the race to populate the cache should not
// error, it should just read what's there). Used by the preview path, where
// a losing race just means "someone else's preview is already cached" — no
// data is lost either way, since a pending placeholder is all preview ever
// writes.
func (h *reefSliceResultHandle) Create(ctx context.Context, result *models.ReefSliceResult) error {
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "geometry_hash"}}, DoNothing: true}).
		Create(result).Error
}

func (h *reefSliceResultHandle) Update(ctx context.Context, result *models.ReefSliceResult) error {
	return h.db.WithContext(ctx).Save(result).Error
}

// Upsert writes a fully-computed result (from the full-generation path),
// overwriting whatever's already at this geometry_hash on conflict rather
// than discarding the write. DoNothing (what Create uses) is wrong here: a
// pending preview placeholder for this exact hash can be inserted at any
// point up until the moment this call commits — including after an initial
// existence check earlier in the caller, since generation+slicing takes
// long enough for a fresh preview request to land in that window — so only
// an atomic upsert, not a check-then-act pattern, actually closes the race.
// Every caller of this always has a complete, non-blank result to write, so
// last-write-wins on a genuine double-computation is harmless.
func (h *reefSliceResultHandle) Upsert(ctx context.Context, result *models.ReefSliceResult) error {
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "geometry_hash"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"updated_at", "product_id", "status", "rejection_rule", "rejection_reason",
				"weight_g", "print_time_s", "bbox_mm", "plate_fits", "support_required",
				"support_material_percent", "min_wall_mm", "sealed_void", "warnings",
				"slicer_version", "openscad_version", "stl_key", "preview_key", "price_cents",
			}),
		}).
		Create(result).Error
}

type reefGenerationJobHandle struct {
	db *gorm.DB
}

func (h *reefGenerationJobHandle) Create(ctx context.Context, job *models.ReefGenerationJob) (*models.ReefGenerationJob, error) {
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	if err := h.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}

func (h *reefGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ReefGenerationJob, error) {
	var job models.ReefGenerationJob
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (h *reefGenerationJobHandle) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"error":      errMsg,
		"updated_at": time.Now(),
	}
	return h.db.WithContext(ctx).Model(&models.ReefGenerationJob{}).Where("id = ?", id).Updates(updates).Error
}

func (h *reefGenerationJobHandle) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.ReefGenerationJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"attempts":   gorm.Expr("attempts + 1"),
			"status":     models.ReefGenerationJobStatusRunning,
			"locked_at":  time.Now(),
			"updated_at": time.Now(),
		}).Error
}
