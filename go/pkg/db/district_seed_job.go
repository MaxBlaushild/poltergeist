package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type districtSeedJobHandle struct {
	db *gorm.DB
}

func (h *districtSeedJobHandle) Create(ctx context.Context, job *models.DistrictSeedJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *districtSeedJobHandle) Update(ctx context.Context, job *models.DistrictSeedJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *districtSeedJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.DistrictSeedJob, error) {
	var job models.DistrictSeedJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *districtSeedJobHandle) FindFiltered(
	ctx context.Context,
	districtID *uuid.UUID,
	statuses []string,
	limit int,
) ([]models.DistrictSeedJob, error) {
	var jobs []models.DistrictSeedJob
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if districtID != nil {
		q = q.Where("district_id = ?", *districtID)
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

func (h *districtSeedJobHandle) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.DistrictSeedJob{}, "id = ?", id).Error
}
