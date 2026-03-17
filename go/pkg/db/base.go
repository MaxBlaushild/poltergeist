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

	return h.FindByUserID(ctx, userID)
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
