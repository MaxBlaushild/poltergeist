package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type feteRoomTeamHandler struct {
	db *gorm.DB
}

func (h *feteRoomTeamHandler) Create(ctx context.Context, roomTeam *models.FeteRoomTeam) error {
	roomTeam.ID = uuid.New()
	roomTeam.CreatedAt = time.Now()
	roomTeam.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Create(roomTeam).Error
}

func (h *feteRoomTeamHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.FeteRoomTeam, error) {
	var roomTeam models.FeteRoomTeam
	if err := h.db.WithContext(ctx).
		Preload("FeteRoom").
		Preload("Team").
		First(&roomTeam, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &roomTeam, nil
}

func (h *feteRoomTeamHandler) FindAll(ctx context.Context) ([]models.FeteRoomTeam, error) {
	var roomTeams []models.FeteRoomTeam
	if err := h.db.WithContext(ctx).
		Preload("FeteRoom").
		Preload("Team").
		Find(&roomTeams).Error; err != nil {
		return nil, err
	}
	return roomTeams, nil
}

func (h *feteRoomTeamHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.FeteRoomTeam{}, id).Error
}

func (h *feteRoomTeamHandler) FindByRoomIDAndTeamID(ctx context.Context, roomID, teamID uuid.UUID) (*models.FeteRoomTeam, error) {
	var roomTeam models.FeteRoomTeam
	if err := h.db.WithContext(ctx).
		Where("fete_room_id = ? AND team_id = ?", roomID, teamID).
		First(&roomTeam).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &roomTeam, nil
}

