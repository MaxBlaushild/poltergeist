package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func defaultScenarioGenreID(
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

func resolveScenarioTemplateGenreID(
	ctx context.Context,
	database *gorm.DB,
	template *models.ScenarioTemplate,
) (uuid.UUID, error) {
	if template != nil && template.GenreID != uuid.Nil {
		return template.GenreID, nil
	}
	return defaultScenarioGenreID(ctx, database)
}

func resolveScenarioTemplateGenreIDForUpdate(
	ctx context.Context,
	database *gorm.DB,
	templateID uuid.UUID,
	template *models.ScenarioTemplate,
) (uuid.UUID, error) {
	if template != nil && template.GenreID != uuid.Nil {
		return template.GenreID, nil
	}
	if templateID != uuid.Nil {
		var existing models.ScenarioTemplate
		if err := database.WithContext(ctx).
			Select("genre_id").
			Where("id = ?", templateID).
			First(&existing).Error; err == nil {
			if existing.GenreID != uuid.Nil {
				return existing.GenreID, nil
			}
		} else if err != gorm.ErrRecordNotFound {
			return uuid.Nil, err
		}
	}
	return defaultScenarioGenreID(ctx, database)
}

func resolveScenarioGenreID(
	ctx context.Context,
	database *gorm.DB,
	scenario *models.Scenario,
) (uuid.UUID, error) {
	if scenario != nil && scenario.GenreID != uuid.Nil {
		return scenario.GenreID, nil
	}
	return defaultScenarioGenreID(ctx, database)
}

func resolveScenarioGenreIDForUpdate(
	ctx context.Context,
	database *gorm.DB,
	scenarioID uuid.UUID,
	scenario *models.Scenario,
) (uuid.UUID, error) {
	if scenario != nil && scenario.GenreID != uuid.Nil {
		return scenario.GenreID, nil
	}
	if scenarioID != uuid.Nil {
		var existing models.Scenario
		if err := database.WithContext(ctx).
			Select("genre_id").
			Where("id = ?", scenarioID).
			First(&existing).Error; err == nil {
			if existing.GenreID != uuid.Nil {
				return existing.GenreID, nil
			}
		} else if err != gorm.ErrRecordNotFound {
			return uuid.Nil, err
		}
	}
	return defaultScenarioGenreID(ctx, database)
}

func resolveScenarioGenerationJobGenreID(
	ctx context.Context,
	database *gorm.DB,
	job *models.ScenarioGenerationJob,
) (uuid.UUID, error) {
	if job != nil && job.GenreID != uuid.Nil {
		return job.GenreID, nil
	}
	return defaultScenarioGenreID(ctx, database)
}

func resolveScenarioTemplateGenerationJobGenreID(
	ctx context.Context,
	database *gorm.DB,
	job *models.ScenarioTemplateGenerationJob,
) (uuid.UUID, error) {
	if job != nil && job.GenreID != uuid.Nil {
		return job.GenreID, nil
	}
	return defaultScenarioGenreID(ctx, database)
}
