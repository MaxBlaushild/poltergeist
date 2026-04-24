package db

import (
	"context"
	"errors"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type zoneShroudConfigHandle struct {
	db *gorm.DB
}

func (h *zoneShroudConfigHandle) Get(ctx context.Context) (*models.ZoneShroudConfig, error) {
	return h.getOrCreate(ctx, h.db.WithContext(ctx))
}

func (h *zoneShroudConfigHandle) Upsert(ctx context.Context, config *models.ZoneShroudConfig) (*models.ZoneShroudConfig, error) {
	if config == nil {
		return nil, gorm.ErrInvalidData
	}
	config.ID = 1
	if err := h.db.WithContext(ctx).Save(config).Error; err != nil {
		return nil, err
	}
	return config, nil
}

func (h *zoneShroudConfigHandle) getOrCreate(ctx context.Context, db *gorm.DB) (*models.ZoneShroudConfig, error) {
	var config models.ZoneShroudConfig
	err := db.WithContext(ctx).First(&config, 1).Error
	if err == nil {
		return &config, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	config = models.ZoneShroudConfig{
		ID:                          1,
		PatternTileURL:              "",
		PatternTilePrompt:           "",
		PatternTileGenerationStatus: models.ZoneKindPatternTileGenerationStatusNone,
		PatternTileGenerationError:  "",
	}
	if err := db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}
