package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type albumInviteHandle struct {
	db *gorm.DB
}

func (h *albumInviteHandle) Create(ctx context.Context, albumID, invitedUserID, invitedByID uuid.UUID, role string) (*models.AlbumInvite, error) {
	if role != "admin" && role != "poster" {
		role = "poster"
	}
	inv := &models.AlbumInvite{
		AlbumID:       albumID,
		InvitedUserID: invitedUserID,
		InvitedByID:   invitedByID,
		Role:          role,
		Status:        "pending",
	}
	err := h.db.WithContext(ctx).Create(inv).Error
	return inv, err
}

func (h *albumInviteHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.AlbumInvite, error) {
	var inv models.AlbumInvite
	err := h.db.WithContext(ctx).Where("id = ?", id).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (h *albumInviteHandle) FindPendingByAlbumID(ctx context.Context, albumID uuid.UUID) ([]models.AlbumInvite, error) {
	var invs []models.AlbumInvite
	err := h.db.WithContext(ctx).Where("album_id = ? AND status = ?", albumID, "pending").
		Order("created_at DESC").Find(&invs).Error
	return invs, err
}

func (h *albumInviteHandle) FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]models.AlbumInvite, error) {
	var invs []models.AlbumInvite
	err := h.db.WithContext(ctx).Where("invited_user_id = ? AND status = ?", userID, "pending").
		Order("created_at DESC").Find(&invs).Error
	return invs, err
}

func (h *albumInviteHandle) FindByUserIDAndStatus(ctx context.Context, userID uuid.UUID, status string) ([]models.AlbumInvite, error) {
	var invs []models.AlbumInvite
	err := h.db.WithContext(ctx).Where("invited_user_id = ? AND status = ?", userID, status).
		Order("created_at DESC").Find(&invs).Error
	return invs, err
}

func (h *albumInviteHandle) FindByAlbumIDAndStatus(ctx context.Context, albumID uuid.UUID, status string) ([]models.AlbumInvite, error) {
	var invs []models.AlbumInvite
	err := h.db.WithContext(ctx).Where("album_id = ? AND status = ?", albumID, status).
		Order("created_at DESC").Find(&invs).Error
	return invs, err
}

func (h *albumInviteHandle) FindByAlbumAndUser(ctx context.Context, albumID, userID uuid.UUID) (*models.AlbumInvite, error) {
	var inv models.AlbumInvite
	err := h.db.WithContext(ctx).
		Where("album_id = ? AND invited_user_id = ?", albumID, userID).
		First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (h *albumInviteHandle) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return h.db.WithContext(ctx).Model(&models.AlbumInvite{}).
		Where("id = ?", id).Update("status", status).Error
}

func (h *albumInviteHandle) Reinvite(ctx context.Context, albumID, userID uuid.UUID, role string) (*models.AlbumInvite, error) {
	if role != "admin" && role != "poster" {
		role = "poster"
	}
	err := h.db.WithContext(ctx).Model(&models.AlbumInvite{}).
		Where("album_id = ? AND invited_user_id = ?", albumID, userID).
		Updates(map[string]interface{}{"status": "pending", "role": role}).Error
	if err != nil {
		return nil, err
	}
	return h.FindByAlbumAndUser(ctx, albumID, userID)
}
