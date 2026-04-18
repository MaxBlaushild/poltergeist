package processors

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const nonFantasySpellIconPromptTemplate = "A retro 16-bit RPG %s icon for %s. School: %s. %s. %s. Genre direction: %s. Make the icon unmistakably %s rather than default fantasy. No characters, no text, no logos, transparent background, centered composition, crisp outlines, limited palette."

func loadSpellGenre(
	ctx context.Context,
	dbClient db.DbClient,
	genreID uuid.UUID,
) (*models.ZoneGenre, error) {
	if dbClient == nil {
		if genreID == uuid.Nil {
			return &models.ZoneGenre{
				Name:       models.DefaultZoneGenreNameFantasy,
				PromptSeed: models.DefaultFantasyZoneGenrePromptSeed(),
				Active:     true,
			}, nil
		}
		return &models.ZoneGenre{
			ID:     genreID,
			Active: true,
		}, nil
	}
	if genreID == uuid.Nil {
		return dbClient.ZoneGenre().FindByName(ctx, models.DefaultZoneGenreNameFantasy)
	}
	return dbClient.ZoneGenre().FindByID(ctx, genreID)
}

func isBaselineFantasySpellGenre(genre *models.ZoneGenre) bool {
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

func spellGenrePromptSeed(genre *models.ZoneGenre) string {
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

func spellGenrePromptLabel(genre *models.ZoneGenre) string {
	if genre == nil {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	trimmedName := strings.TrimSpace(genre.Name)
	if trimmedName == "" {
		return strings.ToLower(models.DefaultZoneGenreNameFantasy)
	}
	return strings.ToLower(trimmedName)
}

func spellAbilityTypePromptLabel(
	abilityType models.SpellAbilityType,
	plural bool,
) string {
	if abilityType == models.SpellAbilityTypeTechnique {
		if plural {
			return "techniques"
		}
		return "technique"
	}
	if plural {
		return "spells"
	}
	return "spell"
}

func spellAbilityGenreInstructionBlock(
	genre *models.ZoneGenre,
	abilityType models.SpellAbilityType,
) string {
	if isBaselineFantasySpellGenre(genre) {
		return ""
	}
	promptSeed := spellGenrePromptSeed(genre)
	if promptSeed == "" {
		promptSeed = fmt.Sprintf(
			"Keep the generated %s unmistakably rooted in %s conventions rather than default fantasy.",
			spellAbilityTypePromptLabel(abilityType, true),
			spellGenrePromptLabel(genre),
		)
	}
	return fmt.Sprintf(
		`Genre direction:
- genre: %s
- creative seed: %s

Additional rules:
- Override the default fantasy RPG baseline when needed; the generated %s should feel authentically %s.
- Favor names, schools, motifs, effects, silhouettes, and combat framing that fit %s conventions.
- Keep the results playable inside the current RPG mechanics even when the fiction shifts genres.
`,
		spellGenrePromptLabel(genre),
		promptSeed,
		spellAbilityTypePromptLabel(abilityType, true),
		spellGenrePromptLabel(genre),
		spellGenrePromptLabel(genre),
	)
}

func spellIconPrompt(
	name string,
	description string,
	school string,
	effectText string,
	abilityType models.SpellAbilityType,
	genre *models.ZoneGenre,
) string {
	if isBaselineFantasySpellGenre(genre) {
		return fmt.Sprintf(spellIconPromptTemplate, name, school, description, effectText)
	}
	promptSeed := spellGenrePromptSeed(genre)
	if promptSeed == "" {
		promptSeed = fmt.Sprintf(
			"Keep the icon unmistakably rooted in %s conventions rather than default fantasy.",
			spellGenrePromptLabel(genre),
		)
	}
	return fmt.Sprintf(
		nonFantasySpellIconPromptTemplate,
		spellAbilityTypePromptLabel(abilityType, false),
		name,
		school,
		description,
		effectText,
		promptSeed,
		spellGenrePromptLabel(genre),
	)
}
