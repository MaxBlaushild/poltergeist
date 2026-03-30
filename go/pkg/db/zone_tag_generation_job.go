package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type zoneTagGenerationJobHandle struct {
	db *gorm.DB
}

func (h *zoneTagGenerationJobHandle) Create(ctx context.Context, job *models.ZoneTagGenerationJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *zoneTagGenerationJobHandle) Update(ctx context.Context, job *models.ZoneTagGenerationJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *zoneTagGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ZoneTagGenerationJob, error) {
	var job models.ZoneTagGenerationJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	if job.SelectedTags == nil {
		job.SelectedTags = models.StringArray{}
	}
	return &job, nil
}

func (h *zoneTagGenerationJobHandle) FindRecent(ctx context.Context, limit int) ([]models.ZoneTagGenerationJob, error) {
	var jobs []models.ZoneTagGenerationJob
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	for i := range jobs {
		if jobs[i].SelectedTags == nil {
			jobs[i].SelectedTags = models.StringArray{}
		}
	}
	return jobs, nil
}

func (h *zoneTagGenerationJobHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID, limit int) ([]models.ZoneTagGenerationJob, error) {
	var jobs []models.ZoneTagGenerationJob
	q := h.db.WithContext(ctx).Where("zone_id = ?", zoneID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	for i := range jobs {
		if jobs[i].SelectedTags == nil {
			jobs[i].SelectedTags = models.StringArray{}
		}
	}
	return jobs, nil
}
