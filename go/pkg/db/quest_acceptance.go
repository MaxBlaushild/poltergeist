package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questAcceptanceHandle struct {
	db *gorm.DB
}

func (h *questAcceptanceHandle) Create(ctx context.Context, questAcceptance *models.QuestAcceptance) error {
	return h.db.WithContext(ctx).Create(questAcceptance).Error
}

func (h *questAcceptanceHandle) FindByUserAndQuest(ctx context.Context, userID uuid.UUID, pointOfInterestGroupID uuid.UUID) (*models.QuestAcceptance, error) {
	var questAcceptance models.QuestAcceptance
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND point_of_interest_group_id = ?", userID, pointOfInterestGroupID).
		First(&questAcceptance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &questAcceptance, nil
}

func (h *questAcceptanceHandle) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.QuestAcceptance, error) {
	var questAcceptances []models.QuestAcceptance
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&questAcceptances).Error; err != nil {
		return nil, err
	}
	return questAcceptances, nil
}

func (h *questAcceptanceHandle) MarkTurnedIn(ctx context.Context, userID uuid.UUID, pointOfInterestGroupID uuid.UUID) error {
	now := time.Now()
	return h.db.WithContext(ctx).
		Model(&models.QuestAcceptance{}).
		Where("user_id = ? AND point_of_interest_group_id = ?", userID, pointOfInterestGroupID).
		Update("turned_in_at", now).Error
}
