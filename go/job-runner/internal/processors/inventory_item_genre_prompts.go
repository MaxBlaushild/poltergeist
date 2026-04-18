package processors

import (
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func isBaselineFantasyInventoryItemGenre(genre *models.ZoneGenre) bool {
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

func inventoryItemGenrePromptSeed(genre *models.ZoneGenre) string {
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

func inventoryItemGenrePromptLabel(genre *models.ZoneGenre) string {
	if genre == nil {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	trimmedName := strings.TrimSpace(genre.Name)
	if trimmedName == "" {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	return strings.ToLower(trimmedName)
}

func inventoryItemImagePrompt(
	name string,
	description string,
	rarityTier string,
	genre *models.ZoneGenre,
) string {
	trimmedDescription := strings.TrimSpace(description)
	if trimmedDescription == "" {
		trimmedDescription = "A unique item"
	}
	if isBaselineFantasyInventoryItemGenre(genre) {
		if strings.TrimSpace(description) == "" {
			trimmedDescription = "A unique fantasy item"
		}
		return fmt.Sprintf(inventoryItemPromptTemplate, name, trimmedDescription, rarityTier)
	}
	promptSeed := inventoryItemGenrePromptSeed(genre)
	if promptSeed == "" {
		promptSeed = fmt.Sprintf(
			"Keep the icon unmistakably rooted in %s conventions rather than default fantasy.",
			inventoryItemGenrePromptLabel(genre),
		)
	}
	genreDescription := fmt.Sprintf(
		"%s Genre direction: %s. %s. Make the icon unmistakably %s rather than default fantasy.",
		trimmedDescription,
		inventoryItemGenrePromptLabel(genre),
		promptSeed,
		inventoryItemGenrePromptLabel(genre),
	)
	return fmt.Sprintf(inventoryItemPromptTemplate, name, genreDescription, rarityTier)
}

func inventoryItemSuggestionGenreInstructionBlock(genre *models.ZoneGenre) string {
	if isBaselineFantasyInventoryItemGenre(genre) {
		return ""
	}
	promptSeed := inventoryItemGenrePromptSeed(genre)
	if promptSeed == "" {
		promptSeed = fmt.Sprintf(
			"Keep the inventory concepts unmistakably rooted in %s conventions rather than generic fantasy.",
			inventoryItemGenrePromptLabel(genre),
		)
	}
	return fmt.Sprintf(
		`Genre override:
- genre: %s
- creative seed: %s

Additional rules:
- Override the default urban fantasy baseline when needed; the batch should feel authentically %s.
- Favor names, materials, motifs, silhouettes, and gameplay affordances that fit %s conventions.
`,
		inventoryItemGenrePromptLabel(genre),
		promptSeed,
		inventoryItemGenrePromptLabel(genre),
		inventoryItemGenrePromptLabel(genre),
	)
}
