package db

import (
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
