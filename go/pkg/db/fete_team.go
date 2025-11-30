package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type feteTeamHandler struct {
	db *gorm.DB
}

func (h *feteTeamHandler) Create(ctx context.Context, team *models.FeteTeam) error {
	team.ID = uuid.New()
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Create(team).Error
}

func (h *feteTeamHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.FeteTeam, error) {
	var team models.FeteTeam
	if err := h.db.WithContext(ctx).First(&team, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &team, nil
}

func (h *feteTeamHandler) FindAll(ctx context.Context) ([]models.FeteTeam, error) {
	var teams []models.FeteTeam
	if err := h.db.WithContext(ctx).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (h *feteTeamHandler) Update(ctx context.Context, id uuid.UUID, updates *models.FeteTeam) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Model(&models.FeteTeam{}).Where("id = ?", id).Updates(updates).Error
}

func (h *feteTeamHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.FeteTeam{}, id).Error
}

