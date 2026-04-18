package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func defaultMonsterGenreID(ctx context.Context, database *gorm.DB) (uuid.UUID, error) {
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

func resolveMonsterTemplateGenreID(
	ctx context.Context,
	database *gorm.DB,
	template *models.MonsterTemplate,
) (uuid.UUID, error) {
	if template != nil && template.GenreID != uuid.Nil {
		return template.GenreID, nil
	}
	return defaultMonsterGenreID(ctx, database)
}

func resolveMonsterGenreID(
	ctx context.Context,
	database *gorm.DB,
	monster *models.Monster,
) (uuid.UUID, error) {
	if monster != nil && monster.GenreID != uuid.Nil {
		return monster.GenreID, nil
	}
	if monster != nil && monster.Template != nil && monster.Template.GenreID != uuid.Nil {
		return monster.Template.GenreID, nil
	}
	if monster != nil && monster.TemplateID != nil && *monster.TemplateID != uuid.Nil {
		var template models.MonsterTemplate
		if err := database.WithContext(ctx).
			Select("genre_id").
			Where("id = ?", *monster.TemplateID).
			First(&template).Error; err != nil {
			return uuid.Nil, err
		}
		if template.GenreID != uuid.Nil {
			return template.GenreID, nil
		}
	}
	return defaultMonsterGenreID(ctx, database)
}
