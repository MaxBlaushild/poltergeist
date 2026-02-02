package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type albumMemberHandle struct {
	db *gorm.DB
}

func (h *albumMemberHandle) Add(ctx context.Context, albumID, userID uuid.UUID, role string) error {
	return h.db.WithContext(ctx).Create(&models.AlbumMember{
		AlbumID: albumID,
		UserID:  userID,
		Role:    role,
	}).Error
}

func (h *albumMemberHandle) Remove(ctx context.Context, albumID, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("album_id = ? AND user_id = ?", albumID, userID).
		Delete(&models.AlbumMember{}).Error
}

func (h *albumMemberHandle) GetRole(ctx context.Context, albumID, userID uuid.UUID) (string, error) {
	var m models.AlbumMember
	err := h.db.WithContext(ctx).Where("album_id = ? AND user_id = ?", albumID, userID).First(&m).Error
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

func (h *albumMemberHandle) FindByAlbumID(ctx context.Context, albumID uuid.UUID) ([]models.AlbumMember, error) {
	var members []models.AlbumMember
	err := h.db.WithContext(ctx).Where("album_id = ?", albumID).Find(&members).Error
	return members, err
}

func (h *albumMemberHandle) FindAlbumIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := h.db.WithContext(ctx).Model(&models.AlbumMember{}).
		Where("user_id = ?", userID).Pluck("album_id", &ids).Error
	return ids, err
}

func (h *albumMemberHandle) UpdateRole(ctx context.Context, albumID, userID uuid.UUID, role string) error {
	return h.db.WithContext(ctx).Model(&models.AlbumMember{}).
		Where("album_id = ? AND user_id = ?", albumID, userID).
		Update("role", role).Error
}
