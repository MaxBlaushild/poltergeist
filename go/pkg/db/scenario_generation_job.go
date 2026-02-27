package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type scenarioGenerationJobHandle struct {
	db *gorm.DB
}

func (h *scenarioGenerationJobHandle) Create(ctx context.Context, job *models.ScenarioGenerationJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *scenarioGenerationJobHandle) Update(ctx context.Context, job *models.ScenarioGenerationJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *scenarioGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ScenarioGenerationJob, error) {
	var job models.ScenarioGenerationJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *scenarioGenerationJobHandle) FindRecent(ctx context.Context, limit int) ([]models.ScenarioGenerationJob, error) {
	var jobs []models.ScenarioGenerationJob
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (h *scenarioGenerationJobHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID, limit int) ([]models.ScenarioGenerationJob, error) {
	var jobs []models.ScenarioGenerationJob
	q := h.db.WithContext(ctx).Where("zone_id = ?", zoneID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
