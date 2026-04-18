package processors

import (
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

const nonFantasyCharacterPromptTemplate = "A retro 16-bit RPG pixel art character portrait of %s. %s. Genre direction: %s. Make the attire, silhouette, props, and expression unmistakably %s rather than default fantasy. Centered, shoulders-up, crisp outlines, limited colors, isolated subject, transparent background with alpha channel, no backdrop scenery, no frame, no text, no logos."

func isBaselineFantasyCharacterGenre(genre *models.ZoneGenre) bool {
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

func characterGenrePromptSeed(genre *models.ZoneGenre) string {
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

func characterGenrePromptLabel(genre *models.ZoneGenre) string {
	if genre == nil {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	trimmedName := strings.TrimSpace(genre.Name)
	if trimmedName == "" {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	return strings.ToLower(trimmedName)
}

func characterImagePrompt(
	name string,
	description string,
	genre *models.ZoneGenre,
) string {
	trimmedDescription := strings.TrimSpace(description)
	if trimmedDescription == "" {
		trimmedDescription = "A memorable character"
	}
	if isBaselineFantasyCharacterGenre(genre) {
		if strings.TrimSpace(description) == "" {
			trimmedDescription = "A memorable fantasy hero"
		}
		return fmt.Sprintf(characterPromptTemplate, name, trimmedDescription)
	}
	promptSeed := characterGenrePromptSeed(genre)
	if promptSeed == "" {
		promptSeed = fmt.Sprintf(
			"Keep the portrait rooted in %s conventions instead of generic fantasy archetypes.",
			characterGenrePromptLabel(genre),
		)
	}
	return fmt.Sprintf(
		nonFantasyCharacterPromptTemplate,
		name,
		trimmedDescription,
		promptSeed,
		characterGenrePromptLabel(genre),
	)
}
