package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type zoneSeedJobHandle struct {
	db *gorm.DB
}

func (h *zoneSeedJobHandle) Create(ctx context.Context, job *models.ZoneSeedJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *zoneSeedJobHandle) Update(ctx context.Context, job *models.ZoneSeedJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *zoneSeedJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ZoneSeedJob, error) {
	var job models.ZoneSeedJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *zoneSeedJobHandle) FindRecent(ctx context.Context, limit int) ([]models.ZoneSeedJob, error) {
	return h.FindFiltered(ctx, nil, nil, limit)
}

func (h *zoneSeedJobHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID, limit int) ([]models.ZoneSeedJob, error) {
	return h.FindFiltered(ctx, &zoneID, nil, limit)
}

func (h *zoneSeedJobHandle) FindFiltered(
	ctx context.Context,
	zoneID *uuid.UUID,
	statuses []string,
	limit int,
) ([]models.ZoneSeedJob, error) {
	var jobs []models.ZoneSeedJob
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if zoneID != nil {
		q = q.Where("zone_id = ?", *zoneID)
	}
	if len(statuses) > 0 {
		q = q.Where("status IN ?", statuses)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (h *zoneSeedJobHandle) ReplaceZoneKind(ctx context.Context, currentKind string, nextKind string) (int, error) {
	normalizedCurrentKind := models.NormalizeZoneKind(currentKind)
	if normalizedCurrentKind == "" {
		return 0, nil
	}

	result := h.db.WithContext(ctx).
		Model(&models.ZoneSeedJob{}).
		Where("zone_kind = ?", normalizedCurrentKind).
		Update("zone_kind", models.NormalizeZoneKind(nextKind))
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

func (h *zoneSeedJobHandle) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.ZoneSeedJob{}, "id = ?", id).Error
}
