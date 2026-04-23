package processors

import (
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

type scenarioZoneKindCueRule struct {
	contentKeywords []string
	zoneCues        []string
}

var scenarioZoneKindCueRules = []scenarioZoneKindCueRule{
	{contentKeywords: []string{"forest", "wood", "grove", "canopy", "briar", "thorn", "moss"}, zoneCues: []string{"forest", "wood", "grove", "wild", "jungle"}},
	{contentKeywords: []string{"swamp", "bog", "mire", "marsh", "ooze", "wetland"}, zoneCues: []string{"swamp", "bog", "marsh", "wetland", "mire"}},
	{contentKeywords: []string{"desert", "dune", "sand", "scorch", "sunbaked", "arid"}, zoneCues: []string{"desert", "dune", "sand", "waste", "arid"}},
	{contentKeywords: []string{"mountain", "peak", "cliff", "highland", "ridge", "crag"}, zoneCues: []string{"mountain", "peak", "highland", "cliff", "rock"}},
	{contentKeywords: []string{"cave", "cavern", "mine", "burrow", "tunnel", "underground"}, zoneCues: []string{"cave", "cavern", "underground", "mine", "tunnel"}},
	{contentKeywords: []string{"crypt", "grave", "tomb", "catacomb", "cemetery", "ruin"}, zoneCues: []string{"crypt", "grave", "tomb", "cemetery", "catacomb", "ruin"}},
	{contentKeywords: []string{"coast", "shore", "reef", "harbor", "sea", "ocean", "tide"}, zoneCues: []string{"coast", "reef", "shore", "sea", "ocean", "water"}},
	{contentKeywords: []string{"river", "lake", "torrent", "flood", "waterfall"}, zoneCues: []string{"river", "lake", "water", "wetland"}},
	{contentKeywords: []string{"ice", "frost", "snow", "glacier", "winter", "tundra"}, zoneCues: []string{"ice", "snow", "tundra", "glacier", "winter"}},
	{contentKeywords: []string{"fire", "ember", "cinder", "magma", "lava", "ash", "volcan"}, zoneCues: []string{"volcan", "lava", "magma", "ash", "fire"}},
	{contentKeywords: []string{"city", "street", "market", "district", "alley", "urban", "fort"}, zoneCues: []string{"city", "urban", "street", "ruin", "fort"}},
}

func buildScenarioZoneKindInstructionBlock(
	zoneKinds []models.ZoneKind,
	preferred *models.ZoneKind,
	preferredLabel string,
) string {
	if len(zoneKinds) == 0 {
		return ""
	}

	lines := []string{
		"Zone kind classification context:",
	}
	if preferred != nil {
		label := strings.TrimSpace(models.ZoneKindPromptLabel(preferred))
		slug := strings.TrimSpace(models.ZoneKindPromptSlug(preferred))
		seed := strings.TrimSpace(models.ZoneKindPromptSeed(preferred))
		if label != "" {
			lines = append(lines, fmt.Sprintf("- %s: %s", preferredLabel, label))
		}
		if slug != "" {
			lines = append(lines, fmt.Sprintf("- %s slug: %s", preferredLabel, slug))
		}
		if seed != "" {
			lines = append(lines, fmt.Sprintf("- %s creative seed: %s", preferredLabel, seed))
		}
	}
	lines = append(
		lines,
		"",
		"Allowed zone kinds:",
		models.ZoneKindsPromptOptions(zoneKinds),
		"",
		"Additional rules:",
		"- Return zoneKind as one allowed slug exactly as written.",
		"- Choose the single best-fit zone kind for where this content would most naturally belong in the reusable content library.",
		"- Base the decision on terrain, props, hazards, factions, traversal, and overall environmental vibe.",
	)
	if preferred != nil {
		lines = append(
			lines,
			fmt.Sprintf("- If the %s is still a strong fit, keep it.", preferredLabel),
		)
	}
	return strings.Join(lines, "\n")
}

func findZoneKindBySlug(zoneKinds []models.ZoneKind, raw string) *models.ZoneKind {
	normalized := models.NormalizeZoneKind(raw)
	if normalized == "" {
		return nil
	}
	for index := range zoneKinds {
		if models.NormalizeZoneKind(zoneKinds[index].Slug) == normalized {
			return &zoneKinds[index]
		}
	}
	return &models.ZoneKind{Slug: normalized}
}

func normalizeScenarioGeneratedZoneKind(
	raw string,
	zoneKinds []models.ZoneKind,
	fallback string,
) string {
	normalized := models.NormalizeZoneKind(raw)
	if normalized != "" {
		if len(zoneKinds) == 0 || zoneKindsContainSlug(zoneKinds, normalized) {
			return normalized
		}
	}
	fallback = models.NormalizeZoneKind(fallback)
	if fallback != "" {
		if len(zoneKinds) == 0 || zoneKindsContainSlug(zoneKinds, fallback) {
			return fallback
		}
	}
	return ""
}

func deriveScenarioZoneKindHeuristically(
	zoneKinds []models.ZoneKind,
	fallback string,
	texts ...string,
) string {
	fallback = normalizeScenarioGeneratedZoneKind("", zoneKinds, fallback)
	if len(zoneKinds) == 0 {
		return fallback
	}

	combined := strings.ToLower(strings.TrimSpace(strings.Join(texts, " ")))
	if combined == "" {
		return fallback
	}

	bestSlug := ""
	bestScore := 0
	for _, zoneKind := range zoneKinds {
		slug := models.NormalizeZoneKind(zoneKind.Slug)
		if slug == "" {
			continue
		}
		haystack := strings.ToLower(strings.TrimSpace(
			zoneKind.Name + " " + zoneKind.Description + " " + slug,
		))
		score := 0

		for _, rule := range scenarioZoneKindCueRules {
			if !containsAnyScenarioKeyword(combined, rule.contentKeywords) {
				continue
			}
			if containsAnyScenarioKeyword(haystack, rule.zoneCues) {
				score += 3
			}
		}

		for _, token := range strings.FieldsFunc(combined, func(char rune) bool {
			return (char < 'a' || char > 'z') && (char < '0' || char > '9')
		}) {
			if len(token) < 4 {
				continue
			}
			if strings.Contains(haystack, token) {
				score++
			}
		}

		if score > bestScore || (score == bestScore && score > 0 && slug < bestSlug) {
			bestScore = score
			bestSlug = slug
		}
	}
	if bestSlug != "" {
		return bestSlug
	}
	return fallback
}

func containsAnyScenarioKeyword(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
