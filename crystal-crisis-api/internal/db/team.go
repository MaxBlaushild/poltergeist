package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/crystal-crisis-api/internal/models"
	"gorm.io/gorm"
)

type teamHandle struct {
	db *gorm.DB
}

func (h *teamHandle) GetAll(ctx context.Context) ([]models.Team, error) {
	var teams []models.Team

	if err := h.db.WithContext(ctx).Preload("UserTeams").Find(&teams).Error; err != nil {
		return nil, err
	}

	return teams, nil
}

func (h *teamHandle) Create(ctx context.Context, userIDs []uint, teamName string) error {
	team := models.Team{Name: teamName}

	if err := h.db.WithContext(ctx).Create(&team).Error; err != nil {
		return err
	}

	userTeams := []models.UserTeam{}
	for _, userID := range userIDs {
		userTeams = append(userTeams, models.UserTeam{
			UserID: userID,
			TeamID: team.ID,
		})
	}

	return h.db.WithContext(ctx).Create(&userTeams).Error
}
