package processors

import (
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

type inventoryItemZoneKindCueRule struct {
	itemKeywords []string
	zoneCues     []string
}

var inventoryItemZoneKindCueRules = []inventoryItemZoneKindCueRule{
	{itemKeywords: []string{"forest", "bark", "moss", "thorn", "grove", "herb", "sap"}, zoneCues: []string{"forest", "wood", "grove", "wild", "jungle"}},
	{itemKeywords: []string{"swamp", "bog", "mire", "marsh", "rot", "ooze", "toxin", "reed"}, zoneCues: []string{"swamp", "bog", "marsh", "wetland", "mire"}},
	{itemKeywords: []string{"desert", "dune", "sand", "sun", "glass", "scorch", "arid"}, zoneCues: []string{"desert", "dune", "sand", "waste", "arid"}},
	{itemKeywords: []string{"mountain", "peak", "ore", "granite", "cliff", "ridge", "stone"}, zoneCues: []string{"mountain", "peak", "highland", "cliff", "rock"}},
	{itemKeywords: []string{"cave", "cavern", "mine", "fungus", "crystal", "burrow", "tunnel"}, zoneCues: []string{"cave", "cavern", "underground", "mine", "tunnel"}},
	{itemKeywords: []string{"crypt", "grave", "bone", "funeral", "tomb", "catacomb", "mourning"}, zoneCues: []string{"crypt", "grave", "tomb", "cemetery", "catacomb", "ruin"}},
	{itemKeywords: []string{"coast", "reef", "shore", "harbor", "salt", "tide", "pearl"}, zoneCues: []string{"coast", "reef", "shore", "sea", "ocean", "water"}},
	{itemKeywords: []string{"river", "lake", "torrent", "rain", "flood", "waterfall"}, zoneCues: []string{"river", "lake", "water", "wetland"}},
	{itemKeywords: []string{"ice", "frost", "snow", "glacier", "winter", "tundra"}, zoneCues: []string{"ice", "snow", "tundra", "glacier", "winter"}},
	{itemKeywords: []string{"fire", "ember", "cinder", "ash", "magma", "lava", "volcan"}, zoneCues: []string{"volcan", "lava", "magma", "ash", "fire"}},
	{itemKeywords: []string{"street", "market", "neon", "alley", "district", "metro", "urban"}, zoneCues: []string{"city", "urban", "street", "ruin", "fort"}},
}

func buildInventoryItemZoneKindInstructionBlock(
	zoneKinds []models.ZoneKind,
	preferred *models.ZoneKind,
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
			lines = append(lines, fmt.Sprintf("- requested zone kind: %s", label))
		}
		if slug != "" {
			lines = append(lines, fmt.Sprintf("- requested zone kind slug: %s", slug))
		}
		if seed != "" {
			lines = append(lines, fmt.Sprintf("- requested zone kind creative seed: %s", seed))
		}
	}
	lines = append(
		lines,
		"",
		"Allowed zone kinds:",
		models.ZoneKindsPromptOptions(zoneKinds),
		"",
		"Additional rules:",
		"- Every draft item must include zoneKind as one allowed slug exactly as written.",
		"- Choose the single best-fit zone kind for where the item would most naturally be found, crafted, traded, looted, or culturally associated.",
		"- Base the decision on materials, provenance, enemies, factions, environment, hazards, traversal, and the overall vibe implied by the item.",
	)
	if preferred != nil {
		lines = append(
			lines,
			"- If the requested zone kind is still a strong fit, keep it.",
		)
	}
	return strings.Join(lines, "\n")
}

func normalizeInventoryItemSuggestionZoneKind(
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

func deriveInventoryItemSuggestionZoneKindHeuristically(
	zoneKinds []models.ZoneKind,
	fallback string,
	texts ...string,
) string {
	fallback = normalizeInventoryItemSuggestionZoneKind("", zoneKinds, fallback)
	if len(zoneKinds) == 0 {
		return fallback
	}
	if len(zoneKinds) == 1 {
		return models.NormalizeZoneKind(zoneKinds[0].Slug)
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

		for _, rule := range inventoryItemZoneKindCueRules {
			if !containsAnyInventoryItemZoneKeyword(combined, rule.itemKeywords) {
				continue
			}
			if containsAnyInventoryItemZoneKeyword(haystack, rule.zoneCues) {
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

func containsAnyInventoryItemZoneKeyword(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
