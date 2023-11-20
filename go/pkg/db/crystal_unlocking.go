package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type crystalUnlockingHandle struct {
	db *gorm.DB
}

func (h *crystalUnlockingHandle) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.CrystalUnlocking, error) {
	var crystalUnlockings []models.CrystalUnlocking

	if err := h.db.WithContext(ctx).Where("team_id = ?", teamID).Find(&crystalUnlockings).Error; err != nil {
		return nil, err
	}

	return crystalUnlockings, nil
}
