package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type feteRoomLinkedListTeamHandler struct {
	db *gorm.DB
}

func (h *feteRoomLinkedListTeamHandler) Create(ctx context.Context, item *models.FeteRoomLinkedListTeam) error {
	item.ID = uuid.New()
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Create(item).Error
}

func (h *feteRoomLinkedListTeamHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.FeteRoomLinkedListTeam, error) {
	var item models.FeteRoomLinkedListTeam
	if err := h.db.WithContext(ctx).
		Preload("FeteRoom").
		Preload("FirstTeam").
		Preload("SecondTeam").
		First(&item, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (h *feteRoomLinkedListTeamHandler) FindAll(ctx context.Context) ([]models.FeteRoomLinkedListTeam, error) {
	var items []models.FeteRoomLinkedListTeam
	if err := h.db.WithContext(ctx).
		Preload("FeteRoom").
		Preload("FirstTeam").
		Preload("SecondTeam").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (h *feteRoomLinkedListTeamHandler) Update(ctx context.Context, id uuid.UUID, updates *models.FeteRoomLinkedListTeam) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Model(&models.FeteRoomLinkedListTeam{}).Where("id = ?", id).Updates(updates).Error
}

func (h *feteRoomLinkedListTeamHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.FeteRoomLinkedListTeam{}, id).Error
}

func (h *feteRoomLinkedListTeamHandler) FindByRoomIDAndFirstTeamID(ctx context.Context, roomID, firstTeamID uuid.UUID) (*models.FeteRoomLinkedListTeam, error) {
	var item models.FeteRoomLinkedListTeam
	if err := h.db.WithContext(ctx).
		Preload("FeteRoom").
		Preload("FirstTeam").
		Preload("SecondTeam").
		Where("fete_room_id = ? AND first_team_id = ?", roomID, firstTeamID).
		First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

