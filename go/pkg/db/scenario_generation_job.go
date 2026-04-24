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
	resolvedGenreID, err := resolveScenarioGenerationJobGenreID(ctx, h.db, job)
	if err != nil {
		return err
	}
	job.GenreID = resolvedGenreID
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *scenarioGenerationJobHandle) Update(ctx context.Context, job *models.ScenarioGenerationJob) error {
	resolvedGenreID, err := resolveScenarioGenerationJobGenreID(ctx, h.db, job)
	if err != nil {
		return err
	}
	job.GenreID = resolvedGenreID
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *scenarioGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ScenarioGenerationJob, error) {
	var job models.ScenarioGenerationJob
	if err := h.db.WithContext(ctx).Preload("Genre").First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *scenarioGenerationJobHandle) FindRecent(ctx context.Context, limit int) ([]models.ScenarioGenerationJob, error) {
	var jobs []models.ScenarioGenerationJob
	q := h.db.WithContext(ctx).Preload("Genre").Order("created_at DESC")
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
	q := h.db.WithContext(ctx).Preload("Genre").Where("zone_id = ?", zoneID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (h *scenarioGenerationJobHandle) ListAdmin(
	ctx context.Context,
	params ScenarioGenerationJobAdminListParams,
) (*ScenarioGenerationJobAdminListResult, error) {
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	countQuery := h.db.WithContext(ctx).Model(&models.ScenarioGenerationJob{})
	if params.ZoneID != nil {
		countQuery = countQuery.Where("zone_id = ?", *params.ZoneID)
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	listQuery := h.db.WithContext(ctx).Preload("Genre").Order("created_at DESC")
	if params.ZoneID != nil {
		listQuery = listQuery.Where("zone_id = ?", *params.ZoneID)
	}

	var jobs []models.ScenarioGenerationJob
	if err := listQuery.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&jobs).Error; err != nil {
		return nil, err
	}

	return &ScenarioGenerationJobAdminListResult{
		Jobs:  jobs,
		Total: total,
	}, nil
}
