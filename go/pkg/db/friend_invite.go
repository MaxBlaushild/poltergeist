package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type friendInviteHandle struct {
	db *gorm.DB
}

func (h *friendInviteHandle) Create(ctx context.Context, inviterID uuid.UUID, inviteeID uuid.UUID) (*models.FriendInvite, error) {
	friendInvite := &models.FriendInvite{
		InviterID: inviterID,
		InviteeID: inviteeID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ID:        uuid.New(),
	}

	if err := h.db.WithContext(ctx).Create(friendInvite).Error; err != nil {
		return nil, err
	}

	return friendInvite, nil
}

func (h *friendInviteHandle) FindAllInvites(ctx context.Context, userID uuid.UUID) ([]models.FriendInvite, error) {
	var friendInvites []models.FriendInvite
	if err := h.db.Preload("Invitee").Preload("Inviter").WithContext(ctx).Where("invitee_id = ? OR inviter_id = ?", userID, userID).Find(&friendInvites).Error; err != nil {
		return nil, err
	}

	return friendInvites, nil
}

func (h *friendInviteHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.FriendInvite, error) {
	var friendInvite models.FriendInvite
	if err := h.db.Preload("Invitee").Preload("Inviter").WithContext(ctx).Where("id = ?", id).First(&friendInvite).Error; err != nil {
		return nil, err
	}

	return &friendInvite, nil
}

func (h *friendInviteHandle) Delete(ctx context.Context, id uuid.UUID) error {
	if err := h.db.WithContext(ctx).Where("id = ?", id).Delete(&models.FriendInvite{}).Error; err != nil {
		return err
	}

	return nil
}
