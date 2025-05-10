package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userZoneReputationHandler struct {
	db *gorm.DB
}

func (h *userZoneReputationHandler) Create(ctx context.Context, userZoneReputation *models.UserZoneReputation) error {
	return h.db.Create(userZoneReputation).Error
}

func (h *userZoneReputationHandler) FindOrCreateForUserAndZone(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID) (*models.UserZoneReputation, error) {
	userZoneReputation, err := h.FindByUserIDAndZoneID(ctx, userID, zoneID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if userZoneReputation != nil {
		return userZoneReputation, nil
	}

	userZoneReputation = &models.UserZoneReputation{
		ID:                uuid.New(),
		UserID:            userID,
		ZoneID:            zoneID,
		Level:             1,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		TotalReputation:   0,
		ReputationOnLevel: 0,
	}

	return userZoneReputation, h.Create(ctx, userZoneReputation)
}

func (h *userZoneReputationHandler) FindByUserIDAndZoneID(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID) (*models.UserZoneReputation, error) {
	var userZoneReputation models.UserZoneReputation
	if err := h.db.Where("user_id = ? AND zone_id = ?", userID, zoneID).First(&userZoneReputation).Error; err != nil {
		return nil, err
	}
	return &userZoneReputation, nil
}

func (h *userZoneReputationHandler) ProcessReputationPointAdditions(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID, reputationPoints int) (*models.UserZoneReputation, error) {
	userZoneReputation, err := h.FindOrCreateForUserAndZone(ctx, userID, zoneID)
	if err != nil {
		return nil, err
	}

	userZoneReputation.AddReputationPoints(reputationPoints)

	return userZoneReputation, h.db.Save(userZoneReputation).Error
}
