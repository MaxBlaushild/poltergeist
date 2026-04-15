package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questAcceptanceV2Handle struct {
	db *gorm.DB
}

func (h *questAcceptanceV2Handle) Create(ctx context.Context, acceptance *models.QuestAcceptanceV2) error {
	return h.db.WithContext(ctx).Create(acceptance).Error
}

func (h *questAcceptanceV2Handle) FindByUserAndQuest(ctx context.Context, userID uuid.UUID, questID uuid.UUID) (*models.QuestAcceptanceV2, error) {
	var acceptance models.QuestAcceptanceV2
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND quest_id = ?", userID, questID).
		First(&acceptance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &acceptance, nil
}

func (h *questAcceptanceV2Handle) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.QuestAcceptanceV2, error) {
	var acceptances []models.QuestAcceptanceV2
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&acceptances).Error; err != nil {
		return nil, err
	}
	return acceptances, nil
}

func (h *questAcceptanceV2Handle) UpdateCurrentNode(ctx context.Context, id uuid.UUID, currentNodeID *uuid.UUID) error {
	updates := map[string]interface{}{
		"updated_at":            time.Now(),
		"current_quest_node_id": currentNodeID,
	}
	return h.db.WithContext(ctx).
		Model(&models.QuestAcceptanceV2{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (h *questAcceptanceV2Handle) MarkObjectivesCompleted(ctx context.Context, id uuid.UUID, completedAt time.Time) error {
	return h.db.WithContext(ctx).
		Model(&models.QuestAcceptanceV2{}).
		Where("id = ? AND objectives_completed_at IS NULL", id).
		Updates(map[string]interface{}{
			"objectives_completed_at": completedAt,
			"updated_at":              completedAt,
		}).Error
}

func (h *questAcceptanceV2Handle) MarkClosed(
	ctx context.Context,
	id uuid.UUID,
	closedAt time.Time,
	closureMethod models.QuestClosureMethod,
	debriefPending bool,
	debriefedAt *time.Time,
) error {
	updates := map[string]interface{}{
		"objectives_completed_at": gorm.Expr("COALESCE(objectives_completed_at, ?)", closedAt),
		"closed_at":               closedAt,
		"closure_method":          models.NormalizeQuestClosureMethod(string(closureMethod)),
		"debrief_pending":         debriefPending,
		"debriefed_at":            debriefedAt,
		"updated_at":              closedAt,
	}
	if debriefedAt != nil {
		updates["turned_in_at"] = *debriefedAt
	} else {
		updates["turned_in_at"] = nil
	}
	return h.db.WithContext(ctx).
		Model(&models.QuestAcceptanceV2{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (h *questAcceptanceV2Handle) MarkDebriefed(ctx context.Context, id uuid.UUID, debriefedAt time.Time) error {
	return h.db.WithContext(ctx).
		Model(&models.QuestAcceptanceV2{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"debrief_pending": false,
			"debriefed_at":    debriefedAt,
			"turned_in_at":    debriefedAt,
			"updated_at":      debriefedAt,
		}).Error
}

func (h *questAcceptanceV2Handle) MarkTurnedIn(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return h.db.WithContext(ctx).
		Model(&models.QuestAcceptanceV2{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"objectives_completed_at": gorm.Expr("COALESCE(objectives_completed_at, ?)", now),
			"closed_at":               gorm.Expr("COALESCE(closed_at, ?)", now),
			"closure_method":          models.QuestClosureMethodInPerson,
			"debrief_pending":         false,
			"debriefed_at":            now,
			"turned_in_at":            now,
			"updated_at":              now,
		}).Error
}
