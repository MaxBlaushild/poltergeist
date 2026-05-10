package db

import (
	"context"
	"errors"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type pointOfInterestShopkeeperSeedConfigHandle struct {
	db *gorm.DB
}

func (h *pointOfInterestShopkeeperSeedConfigHandle) Get(ctx context.Context) (*models.PointOfInterestShopkeeperSeedConfig, error) {
	return h.getOrCreate(ctx, h.db.WithContext(ctx))
}

func (h *pointOfInterestShopkeeperSeedConfigHandle) Upsert(
	ctx context.Context,
	config *models.PointOfInterestShopkeeperSeedConfig,
) (*models.PointOfInterestShopkeeperSeedConfig, error) {
	if config == nil {
		return nil, gorm.ErrInvalidData
	}
	config.ID = 1
	if err := h.db.WithContext(ctx).Save(config).Error; err != nil {
		return nil, err
	}
	return config, nil
}

func (h *pointOfInterestShopkeeperSeedConfigHandle) getOrCreate(
	ctx context.Context,
	db *gorm.DB,
) (*models.PointOfInterestShopkeeperSeedConfig, error) {
	var config models.PointOfInterestShopkeeperSeedConfig
	err := db.WithContext(ctx).First(&config, 1).Error
	if err == nil {
		return &config, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	config = models.PointOfInterestShopkeeperSeedConfig{
		ID:       1,
		Profiles: models.DefaultPointOfInterestShopkeeperSeedProfiles(),
	}
	if err := db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}
