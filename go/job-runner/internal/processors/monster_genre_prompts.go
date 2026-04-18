package processors

import (
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

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

func monsterGenreName(genre *models.ZoneGenre) string {
	if genre == nil {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	trimmedName := strings.TrimSpace(genre.Name)
	if trimmedName == "" {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	return strings.ToLower(trimmedName)
}

func monsterGenreVisualDirective(genre *models.ZoneGenre) string {
	if isBaselineFantasyMonsterGenre(genre) {
		return "Aggressive fantasy creature"
	}
	promptSeed := monsterGenrePromptSeed(genre)
	if promptSeed == "" {
		return fmt.Sprintf(
			"Aggressive %s creature with genre-authentic visual language",
			monsterGenreName(genre),
		)
	}
	return fmt.Sprintf(
		"Aggressive %s creature. Genre direction: %s",
		monsterGenreName(genre),
		promptSeed,
	)
}
