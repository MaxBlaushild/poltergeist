package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type albumPostHandle struct {
	db *gorm.DB
}

func (h *albumPostHandle) Add(ctx context.Context, albumID, postID uuid.UUID) error {
	return h.db.WithContext(ctx).Create(&models.AlbumPost{
		AlbumID: albumID,
		PostID:  postID,
	}).Error
}

func (h *albumPostHandle) Remove(ctx context.Context, albumID, postID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("album_id = ? AND post_id = ?", albumID, postID).
		Delete(&models.AlbumPost{}).Error
}

func (h *albumPostHandle) FindPostIDsByAlbumID(ctx context.Context, albumID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := h.db.WithContext(ctx).Model(&models.AlbumPost{}).
		Where("album_id = ?", albumID).Pluck("post_id", &ids).Error
	return ids, err
}

func (h *albumPostHandle) HasAny(ctx context.Context, albumID uuid.UUID) (bool, error) {
	var count int64
	err := h.db.WithContext(ctx).Model(&models.AlbumPost{}).
		Where("album_id = ?", albumID).Count(&count).Error
	return count > 0, err
}
