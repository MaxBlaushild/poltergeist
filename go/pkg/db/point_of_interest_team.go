package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestTeamHandle struct {
	db *gorm.DB
}

func (h *pointOfInterestTeamHandle) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.PointOfInterestTeam, error) {
	var pointOfInterestTeams []models.PointOfInterestTeam

	if err := h.db.WithContext(ctx).Where("team_id = ?", teamID).Find(&pointOfInterestTeams).Error; err != nil {
		return nil, err
	}

	return pointOfInterestTeams, nil
}
