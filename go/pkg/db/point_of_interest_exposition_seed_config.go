package db

import (
	"context"
	"errors"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type pointOfInterestExpositionSeedConfigHandle struct {
	db *gorm.DB
}

func (h *pointOfInterestExpositionSeedConfigHandle) Get(ctx context.Context) (*models.PointOfInterestExpositionSeedConfig, error) {
	return h.getOrCreate(ctx, h.db.WithContext(ctx))
}

func (h *pointOfInterestExpositionSeedConfigHandle) Upsert(
	ctx context.Context,
	config *models.PointOfInterestExpositionSeedConfig,
) (*models.PointOfInterestExpositionSeedConfig, error) {
	if config == nil {
		return nil, gorm.ErrInvalidData
	}
	config.ID = 1
	if err := h.db.WithContext(ctx).Save(config).Error; err != nil {
		return nil, err
	}
	return config, nil
}

func (h *pointOfInterestExpositionSeedConfigHandle) getOrCreate(
	ctx context.Context,
	db *gorm.DB,
) (*models.PointOfInterestExpositionSeedConfig, error) {
	var config models.PointOfInterestExpositionSeedConfig
	err := db.WithContext(ctx).First(&config, 1).Error
	if err == nil {
		return &config, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	config = models.PointOfInterestExpositionSeedConfig{
		ID:       1,
		Profiles: models.DefaultPointOfInterestExpositionSeedProfiles(),
	}
	if err := db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}
