package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questNodeProgressHandle struct {
	db *gorm.DB
}

func (h *questNodeProgressHandle) Create(ctx context.Context, progress *models.QuestNodeProgress) error {
	return h.db.WithContext(ctx).Create(progress).Error
}

func (h *questNodeProgressHandle) FindByAcceptanceID(ctx context.Context, acceptanceID uuid.UUID) ([]models.QuestNodeProgress, error) {
	var progress []models.QuestNodeProgress
	if err := h.db.WithContext(ctx).
		Where("quest_acceptance_id = ?", acceptanceID).
		Find(&progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}

func (h *questNodeProgressHandle) MarkCompleted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return h.db.WithContext(ctx).
		Model(&models.QuestNodeProgress{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"completed_at": now, "updated_at": now}).Error
}

func (h *questNodeProgressHandle) FindByAcceptanceAndNode(ctx context.Context, acceptanceID uuid.UUID, nodeID uuid.UUID) (*models.QuestNodeProgress, error) {
	var progress models.QuestNodeProgress
	if err := h.db.WithContext(ctx).
		Where("quest_acceptance_id = ? AND quest_node_id = ?", acceptanceID, nodeID).
		First(&progress).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &progress, nil
}

func (h *questNodeProgressHandle) DeleteByNodeID(ctx context.Context, nodeID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("quest_node_id = ?", nodeID).Delete(&models.QuestNodeProgress{}).Error
}
