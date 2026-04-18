package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func defaultPointOfInterestGenreID(
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

func resolvePointOfInterestGenreID(
	ctx context.Context,
	database *gorm.DB,
	pointOfInterest *models.PointOfInterest,
) (uuid.UUID, error) {
	if pointOfInterest != nil && pointOfInterest.GenreID != uuid.Nil {
		return pointOfInterest.GenreID, nil
	}
	return defaultPointOfInterestGenreID(ctx, database)
}

func resolvePointOfInterestGenreIDForUpdate(
	ctx context.Context,
	database *gorm.DB,
	pointOfInterestID uuid.UUID,
	pointOfInterest *models.PointOfInterest,
) (uuid.UUID, error) {
	if pointOfInterest != nil && pointOfInterest.GenreID != uuid.Nil {
		return pointOfInterest.GenreID, nil
	}
	if pointOfInterestID != uuid.Nil {
		var existing models.PointOfInterest
		if err := database.WithContext(ctx).
			Select("genre_id").
			Where("id = ?", pointOfInterestID).
			First(&existing).Error; err == nil {
			if existing.GenreID != uuid.Nil {
				return existing.GenreID, nil
			}
		} else if err != gorm.ErrRecordNotFound {
			return uuid.Nil, err
		}
	}
	return defaultPointOfInterestGenreID(ctx, database)
}

func resolvePointOfInterestImportGenreID(
	ctx context.Context,
	database *gorm.DB,
	item *models.PointOfInterestImport,
) (uuid.UUID, error) {
	if item != nil && item.GenreID != uuid.Nil {
		return item.GenreID, nil
	}
	return defaultPointOfInterestGenreID(ctx, database)
}
