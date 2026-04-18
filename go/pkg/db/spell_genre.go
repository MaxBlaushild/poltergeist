package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func defaultSpellGenreID(
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

func resolveSpellGenreID(
	ctx context.Context,
	database *gorm.DB,
	spell *models.Spell,
) (uuid.UUID, error) {
	if spell != nil && spell.GenreID != uuid.Nil {
		return spell.GenreID, nil
	}
	return defaultSpellGenreID(ctx, database)
}
