package db

import (
	"context"
	stdErrors "errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type baseHandle struct {
	db *gorm.DB
}

func (h *baseHandle) UpsertForUser(ctx context.Context, userID uuid.UUID, latitude float64, longitude float64) (*models.Base, error) {
	base := &models.Base{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    userID,
		Latitude:  latitude,
		Longitude: longitude,
	}
	if err := base.SetGeometry(latitude, longitude); err != nil {
		return nil, err
	}

	if err := h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"latitude":   latitude,
				"longitude":  longitude,
				"geometry":   base.Geometry,
				"updated_at": time.Now(),
			}),
		}).
		Create(base).Error; err != nil {
		return nil, err
	}

	base, err := h.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if base != nil {
		if err := (&userBaseStructureHandle{db: h.db}).EnsureBuilt(ctx, base.ID, userID, "hearth", 1, 2, 2); err != nil {
			return nil, err
		}
	}
	return base, nil
}

func (h *baseHandle) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Base, error) {
	var base models.Base
	err := h.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		First(&base).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &base, nil
}

func (h *baseHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Base, error) {
	var base models.Base
	err := h.db.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&base).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &base, nil
}

func (h *baseHandle) FindByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]models.Base, error) {
	if len(userIDs) == 0 {
		return []models.Base{}, nil
	}
	var bases []models.Base
	if err := h.db.WithContext(ctx).
		Preload("User").
		Where("user_id IN ?", userIDs).
		Order("updated_at DESC").
		Find(&bases).Error; err != nil {
		return nil, err
	}
	return bases, nil
}

func (h *baseHandle) FindAll(ctx context.Context) ([]models.Base, error) {
	var bases []models.Base
	if err := h.db.WithContext(ctx).
		Preload("User").
		Order("updated_at DESC").
		Find(&bases).Error; err != nil {
		return nil, err
	}
	return bases, nil
}

func (h *baseHandle) UpdateDetails(ctx context.Context, id uuid.UUID, name *string, description *string) error {
	return h.db.WithContext(ctx).
		Model(&models.Base{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":        name,
			"description": description,
			"updated_at":  time.Now(),
		}).Error
}

func (h *baseHandle) UpdateFlavor(ctx context.Context, id uuid.UUID, description string, imageURL string) error {
	return h.db.WithContext(ctx).
		Model(&models.Base{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"description": description,
			"image_url":   imageURL,
			"updated_at":  time.Now(),
		}).Error
}

func (h *baseHandle) UpdateThumbnailURL(ctx context.Context, id uuid.UUID, thumbnailURL string) error {
	return h.db.WithContext(ctx).
		Model(&models.Base{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"thumbnail_url": thumbnailURL,
			"updated_at":    time.Now(),
		}).Error
}

func (h *baseHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Base{}, "id = ?", id).Error
}
