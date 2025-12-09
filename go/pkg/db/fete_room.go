package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type feteRoomHandler struct {
	db *gorm.DB
}

func (h *feteRoomHandler) Create(ctx context.Context, room *models.FeteRoom) error {
	room.ID = uuid.New()
	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Create(room).Error
}

func (h *feteRoomHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.FeteRoom, error) {
	var room models.FeteRoom
	if err := h.db.WithContext(ctx).
		Preload("CurrentTeam").
		First(&room, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

func (h *feteRoomHandler) FindAll(ctx context.Context) ([]models.FeteRoom, error) {
	var rooms []models.FeteRoom
	if err := h.db.WithContext(ctx).
		Preload("CurrentTeam").
		Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (h *feteRoomHandler) Update(ctx context.Context, id uuid.UUID, updates *models.FeteRoom) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Model(&models.FeteRoom{}).
		Where("id = ?", id).
		Select("Name", "Open", "CurrentTeamID", "HueLightID", "UpdatedAt").
		Updates(updates).Error
}

func (h *feteRoomHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.FeteRoom{}, id).Error
}

