package db

import (
	"context"
	"errors"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type albumShareHandle struct {
	db *gorm.DB
}

func (h *albumShareHandle) Create(ctx context.Context, albumID, createdBy uuid.UUID, token string) (*models.AlbumShare, error) {
	share := &models.AlbumShare{
		AlbumID:   albumID,
		CreatedBy: createdBy,
		Token:     token,
	}
	if err := h.db.WithContext(ctx).Create(share).Error; err != nil {
		return nil, err
	}
	return share, nil
}

func (h *albumShareHandle) FindByAlbumID(ctx context.Context, albumID uuid.UUID) (*models.AlbumShare, error) {
	var share models.AlbumShare
	if err := h.db.WithContext(ctx).Where("album_id = ?", albumID).First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &share, nil
}

func (h *albumShareHandle) FindByToken(ctx context.Context, token string) (*models.AlbumShare, error) {
	var share models.AlbumShare
	if err := h.db.WithContext(ctx).Where("token = ?", token).First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &share, nil
}
