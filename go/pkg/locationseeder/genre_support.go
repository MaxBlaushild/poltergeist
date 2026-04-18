package locationseeder

import (
	"context"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func (c *client) resolvePointOfInterestGenre(
	ctx context.Context,
	preloaded *models.ZoneGenre,
	genreID uuid.UUID,
) (*models.ZoneGenre, error) {
	if preloaded != nil {
		if strings.TrimSpace(preloaded.PromptSeed) == "" &&
			models.IsFantasyZoneGenreName(preloaded.Name) {
			preloaded.PromptSeed = models.DefaultFantasyZoneGenrePromptSeed()
		}
		return preloaded, nil
	}
	if c != nil && c.dbClient != nil {
		if genreID != uuid.Nil {
			genre, err := c.dbClient.ZoneGenre().FindByID(ctx, genreID)
			if err != nil {
				return nil, err
			}
			if genre != nil {
				return genre, nil
			}
		}
		genre, err := c.dbClient.ZoneGenre().FindByName(
			ctx,
			models.DefaultZoneGenreNameFantasy,
		)
		if err != nil {
			return nil, err
		}
		if genre != nil {
			return genre, nil
		}
	}

	resolvedID := genreID
	if resolvedID == uuid.Nil {
		resolvedID = uuid.New()
	}
	return &models.ZoneGenre{
		ID:         resolvedID,
		Name:       models.DefaultZoneGenreNameFantasy,
		PromptSeed: models.DefaultFantasyZoneGenrePromptSeed(),
		Active:     true,
	}, nil
}
