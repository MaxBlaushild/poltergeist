package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestDiscoveryHandle struct {
	db *gorm.DB
}

func (h *pointOfInterestDiscoveryHandle) GetDiscoveriesForTeam(teamID uuid.UUID) ([]models.PointOfInterestDiscovery, error) {
	var discoveries []models.PointOfInterestDiscovery
	if err := h.db.Where("team_id = ?", teamID).Find(&discoveries).Error; err != nil {
		return nil, err
	}
	return discoveries, nil
}

func (h *pointOfInterestDiscoveryHandle) GetDiscoveriesForUser(userID uuid.UUID) ([]models.PointOfInterestDiscovery, error) {
	var discoveries []models.PointOfInterestDiscovery
	if err := h.db.Where("user_id = ?", userID).Find(&discoveries).Error; err != nil {
		return nil, err
	}
	return discoveries, nil
}

func (h *pointOfInterestDiscoveryHandle) ExistsForTeamAndPointOfInterest(ctx context.Context, teamID uuid.UUID, pointOfInterestID uuid.UUID) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).
		Model(&models.PointOfInterestDiscovery{}).
		Where("team_id = ? AND point_of_interest_id = ?", teamID, pointOfInterestID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (h *pointOfInterestDiscoveryHandle) ExistsForUserAndPointOfInterest(ctx context.Context, userID uuid.UUID, pointOfInterestID uuid.UUID) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).
		Model(&models.PointOfInterestDiscovery{}).
		Where("user_id = ? AND point_of_interest_id = ?", userID, pointOfInterestID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (h *pointOfInterestDiscoveryHandle) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.PointOfInterestDiscovery{}).Error
}

func (h *pointOfInterestDiscoveryHandle) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Where("id = ?", id).Delete(&models.PointOfInterestDiscovery{}).Error
}

func (h *pointOfInterestDiscoveryHandle) Create(ctx context.Context, discovery *models.PointOfInterestDiscovery) error {
	return h.db.WithContext(ctx).Create(discovery).Error
}
