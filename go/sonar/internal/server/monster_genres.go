package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *server) resolveZoneGenre(
	ctx context.Context,
	rawGenreID string,
) (*models.ZoneGenre, error) {
	trimmedGenreID := strings.TrimSpace(rawGenreID)
	if s == nil || s.dbClient == nil {
		if trimmedGenreID == "" {
			return &models.ZoneGenre{
				Name:       models.DefaultZoneGenreNameFantasy,
				PromptSeed: models.DefaultFantasyZoneGenrePromptSeed(),
				Active:     true,
			}, nil
		}
		genreID, err := uuid.Parse(trimmedGenreID)
		if err != nil || genreID == uuid.Nil {
			return nil, fmt.Errorf("genreId must be a valid UUID")
		}
		return &models.ZoneGenre{
			ID:     genreID,
			Active: true,
		}, nil
	}
	if trimmedGenreID == "" {
		genre, err := s.dbClient.ZoneGenre().FindByName(
			ctx,
			models.DefaultZoneGenreNameFantasy,
		)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("default fantasy genre not found")
			}
			return nil, err
		}
		return genre, nil
	}

	genreID, err := uuid.Parse(trimmedGenreID)
	if err != nil || genreID == uuid.Nil {
		return nil, fmt.Errorf("genreId must be a valid UUID")
	}

	genre, err := s.dbClient.ZoneGenre().FindByID(ctx, genreID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("genre not found")
		}
		return nil, err
	}
	return genre, nil
}

func (s *server) resolveMonsterGenre(
	ctx context.Context,
	rawGenreID string,
) (*models.ZoneGenre, error) {
	return s.resolveZoneGenre(ctx, rawGenreID)
}

func isBaselineFantasyMonsterGenre(genre *models.ZoneGenre) bool {
	if genre == nil {
		return true
	}
	if !models.IsFantasyZoneGenreName(genre.Name) {
		return false
	}
	trimmedPromptSeed := strings.TrimSpace(genre.PromptSeed)
	return trimmedPromptSeed == "" ||
		trimmedPromptSeed == models.DefaultFantasyZoneGenrePromptSeed()
}

func monsterGenrePromptSeed(genre *models.ZoneGenre) string {
	if genre == nil {
		return models.DefaultFantasyZoneGenrePromptSeed()
	}
	trimmedPromptSeed := strings.TrimSpace(genre.PromptSeed)
	if trimmedPromptSeed != "" {
		return trimmedPromptSeed
	}
	if models.IsFantasyZoneGenreName(genre.Name) {
		return models.DefaultFantasyZoneGenrePromptSeed()
	}
	return ""
}

func monsterGenrePromptLabel(genre *models.ZoneGenre) string {
	if genre == nil {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	trimmedName := strings.TrimSpace(genre.Name)
	if trimmedName == "" {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	return strings.ToLower(trimmedName)
}
