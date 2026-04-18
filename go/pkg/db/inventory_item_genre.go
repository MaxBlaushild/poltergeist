package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func defaultInventoryItemGenreID(
	ctx context.Context,
	database *gorm.DB,
) (uuid.UUID, error) {
	var genre models.ZoneGenre
	if err := database.WithContext(ctx).
		Select("id").
		Where("LOWER(name) = LOWER(?)", models.DefaultZoneGenreNameFantasy).
		Order("sort_order ASC").
		Order("created_at ASC").
		First(&genre).Error; err != nil {
		return uuid.Nil, err
	}
	return genre.ID, nil
}

func resolveInventoryItemGenreID(
	ctx context.Context,
	database *gorm.DB,
	item *models.InventoryItem,
) (uuid.UUID, error) {
	if item != nil && item.GenreID != uuid.Nil {
		return item.GenreID, nil
	}
	return defaultInventoryItemGenreID(ctx, database)
}

func resolveInventoryItemSuggestionJobGenreID(
	ctx context.Context,
	database *gorm.DB,
	job *models.InventoryItemSuggestionJob,
) (uuid.UUID, error) {
	if job != nil && job.GenreID != uuid.Nil {
		return job.GenreID, nil
	}
	return defaultInventoryItemGenreID(ctx, database)
}
