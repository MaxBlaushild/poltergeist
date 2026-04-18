package processors

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func loadScenarioGenre(
	ctx context.Context,
	dbClient db.DbClient,
	genreID uuid.UUID,
	preloaded *models.ZoneGenre,
) (*models.ZoneGenre, error) {
	if preloaded != nil {
		return preloaded, nil
	}
	if dbClient == nil {
		return &models.ZoneGenre{
			ID:         genreID,
			Name:       models.DefaultZoneGenreNameFantasy,
			PromptSeed: models.DefaultFantasyZoneGenrePromptSeed(),
			Active:     true,
		}, nil
	}
	if genreID != uuid.Nil {
		genre, err := dbClient.ZoneGenre().FindByID(ctx, genreID)
		if err != nil {
			return nil, err
		}
		if genre != nil {
			return genre, nil
		}
	}
	return dbClient.ZoneGenre().FindByName(ctx, models.DefaultZoneGenreNameFantasy)
}

func isBaselineFantasyScenarioGenre(genre *models.ZoneGenre) bool {
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

func scenarioGenrePromptSeed(genre *models.ZoneGenre) string {
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

func scenarioGenrePromptLabel(genre *models.ZoneGenre) string {
	if genre == nil {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	trimmedName := strings.TrimSpace(genre.Name)
	if trimmedName == "" {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	return strings.ToLower(trimmedName)
}

func scenarioGenreInstructionBlock(genre *models.ZoneGenre) string {
	if isBaselineFantasyScenarioGenre(genre) {
		return ""
	}
	promptSeed := scenarioGenrePromptSeed(genre)
	if promptSeed == "" {
		promptSeed = fmt.Sprintf(
			"Keep the scenario unmistakably rooted in %s conventions rather than default fantasy.",
			scenarioGenrePromptLabel(genre),
		)
	}
	return fmt.Sprintf(
		`Genre direction:
- genre: %s
- creative seed: %s

Additional rules:
- The scenario must feel authentically %s rather than generic fantasy.
- Use genre-appropriate stakes, props, factions, and environmental logic.
`,
		scenarioGenrePromptLabel(genre),
		promptSeed,
		scenarioGenrePromptLabel(genre),
	)
}

func scenarioGenreImageDirection(genre *models.ZoneGenre) string {
	if isBaselineFantasyScenarioGenre(genre) {
		return ""
	}
	promptSeed := scenarioGenrePromptSeed(genre)
	if promptSeed == "" {
		promptSeed = fmt.Sprintf(
			"Use unmistakable %s visual language with setting-authentic silhouettes, props, and atmosphere.",
			scenarioGenrePromptLabel(genre),
		)
	}
	return fmt.Sprintf(
		"Render the scene using unmistakable %s visual language. %s",
		scenarioGenrePromptLabel(genre),
		promptSeed,
	)
}
