package server

import (
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

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
