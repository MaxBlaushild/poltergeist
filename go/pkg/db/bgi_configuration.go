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

type bgiConfigurationHandle struct {
	db *gorm.DB
}

func (h *bgiConfigurationHandle) Create(ctx context.Context, cfg *models.BgiConfiguration) (*models.BgiConfiguration, error) {
	if cfg.ID == uuid.Nil {
		cfg.ID = uuid.New()
	}
	if err := h.db.WithContext(ctx).Create(cfg).Error; err != nil {
		return nil, err
	}
	return cfg, nil
}

func (h *bgiConfigurationHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BgiConfiguration, error) {
	var cfg models.BgiConfiguration
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (h *bgiConfigurationHandle) Update(ctx context.Context, cfg *models.BgiConfiguration) error {
	return h.db.WithContext(ctx).Save(cfg).Error
}

type bgiSetResolutionHandle struct {
	db *gorm.DB
}

// FindByConfigHash is the R-4.3 recipe cache lookup: a hit skips
// set.Assemble entirely (see go/pkg/reef/set).
func (h *bgiSetResolutionHandle) FindByConfigHash(ctx context.Context, configHash string) (*models.BgiSetResolution, error) {
	var res models.BgiSetResolution
	err := h.db.WithContext(ctx).Where("config_hash = ?", configHash).First(&res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create does nothing on conflict — identical configs must never
// re-resolve, and a losing race just means someone else's resolution is
// already cached (same reasoning as reefSliceResultHandle.Create).
func (h *bgiSetResolutionHandle) Create(ctx context.Context, res *models.BgiSetResolution) error {
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "config_hash"}}, DoNothing: true}).
		Create(res).Error
}

type bgiTraySliceResultHandle struct {
	db *gorm.DB
}

func (h *bgiTraySliceResultHandle) FindByGeometryHash(ctx context.Context, geometryHash string) (*models.BgiTraySliceResult, error) {
	var result models.BgiTraySliceResult
	err := h.db.WithContext(ctx).Where("geometry_hash = ?", geometryHash).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *bgiTraySliceResultHandle) Create(ctx context.Context, result *models.BgiTraySliceResult) error {
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "geometry_hash"}}, DoNothing: true}).
		Create(result).Error
}

// Upsert mirrors reefSliceResultHandle.Upsert's reasoning exactly: a
// complete result from the full-generation path must win over any pending
// preview placeholder written for the same hash in the meantime.
func (h *bgiTraySliceResultHandle) Upsert(ctx context.Context, result *models.BgiTraySliceResult) error {
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "geometry_hash"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"updated_at", "tray_template_id", "status", "rejection_rule", "rejection_reason",
				"weight_g", "print_time_s", "bbox_mm", "plate_fits", "support_required",
				"support_material_percent", "min_wall_mm", "sealed_void", "warnings",
				"slicer_version", "openscad_version", "stl_key", "preview_key", "price_cents",
			}),
		}).
		Create(result).Error
}

type bgiGenerationJobHandle struct {
	db *gorm.DB
}

func (h *bgiGenerationJobHandle) Create(ctx context.Context, job *models.BgiGenerationJob) (*models.BgiGenerationJob, error) {
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	if err := h.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}

func (h *bgiGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BgiGenerationJob, error) {
	var job models.BgiGenerationJob
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (h *bgiGenerationJobHandle) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"error":      errMsg,
		"updated_at": time.Now(),
	}
	return h.db.WithContext(ctx).Model(&models.BgiGenerationJob{}).Where("id = ?", id).Updates(updates).Error
}

func (h *bgiGenerationJobHandle) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.BgiGenerationJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"attempts":   gorm.Expr("attempts + 1"),
			"status":     models.BgiGenerationJobStatusRunning,
			"locked_at":  time.Now(),
			"updated_at": time.Now(),
		}).Error
}
