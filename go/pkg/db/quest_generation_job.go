package db

import (
	"context"
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questGenerationJobHandle struct {
	db *gorm.DB
}

func (h *questGenerationJobHandle) Create(ctx context.Context, job *models.QuestGenerationJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *questGenerationJobHandle) Update(ctx context.Context, job *models.QuestGenerationJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *questGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestGenerationJob, error) {
	var job models.QuestGenerationJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *questGenerationJobHandle) FindByZoneQuestArchetypeID(ctx context.Context, zoneQuestArchetypeID uuid.UUID, limit int) ([]*models.QuestGenerationJob, error) {
	var jobs []*models.QuestGenerationJob
	query := h.db.WithContext(ctx).
		Where("zone_quest_archetype_id = ?", zoneQuestArchetypeID).
		Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (h *questGenerationJobHandle) FindByQuestArchetypeIDAndZoneID(ctx context.Context, questArchetypeID uuid.UUID, zoneID uuid.UUID, limit int) ([]*models.QuestGenerationJob, error) {
	var jobs []*models.QuestGenerationJob
	query := h.db.WithContext(ctx).
		Where("quest_archetype_id = ? AND zone_id = ?", questArchetypeID, zoneID).
		Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (h *questGenerationJobHandle) MarkInProgress(ctx context.Context, id uuid.UUID) error {
	updates := map[string]interface{}{
		"status":     models.QuestGenerationStatusInProgress,
		"updated_at": time.Now(),
	}
	return h.db.WithContext(ctx).
		Model(&models.QuestGenerationJob{}).
		Where("id = ? AND status = ?", id, models.QuestGenerationStatusQueued).
		Updates(updates).Error
}

func (h *questGenerationJobHandle) RecordSuccess(ctx context.Context, id uuid.UUID, questID uuid.UUID) error {
	payload := fmt.Sprintf("[\"%s\"]", questID.String())
	return h.db.WithContext(ctx).Exec(`
		UPDATE quest_generation_jobs
		SET completed_count = completed_count + 1,
			quest_ids = COALESCE(quest_ids, '[]'::jsonb) || ?::jsonb,
			updated_at = NOW(),
			status = CASE
				WHEN completed_count + 1 + failed_count >= total_count AND failed_count = 0 THEN ?
				WHEN completed_count + 1 + failed_count >= total_count AND failed_count > 0 THEN ?
				ELSE ?
			END
		WHERE id = ?`, payload, models.QuestGenerationStatusCompleted, models.QuestGenerationStatusFailed, models.QuestGenerationStatusInProgress, id).Error
}

func (h *questGenerationJobHandle) RecordFailure(ctx context.Context, id uuid.UUID, errMsg string) error {
	return h.db.WithContext(ctx).Exec(`
		UPDATE quest_generation_jobs
		SET failed_count = failed_count + 1,
			error_message = ?,
			updated_at = NOW(),
			status = CASE
				WHEN completed_count + failed_count + 1 >= total_count THEN ?
				ELSE ?
			END
		WHERE id = ?`, errMsg, models.QuestGenerationStatusFailed, models.QuestGenerationStatusInProgress, id).Error
}
