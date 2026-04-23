package models

import (
	"fmt"
	"strings"
)

func ZoneKindPromptSlug(zoneKind *ZoneKind) string {
	if zoneKind == nil {
		return ""
	}
	return NormalizeZoneKind(zoneKind.Slug)
}

func ZoneKindPromptLabel(zoneKind *ZoneKind) string {
	if zoneKind == nil {
		return ""
	}
	if name := strings.TrimSpace(zoneKind.Name); name != "" {
		return name
	}
	slug := ZoneKindPromptSlug(zoneKind)
	if slug == "" {
		return ""
	}
	parts := strings.FieldsFunc(slug, func(char rune) bool {
		return char == '-' || char == '_'
	})
	for index, part := range parts {
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		parts[index] = strings.ToUpper(lower[:1]) + lower[1:]
	}
	return strings.Join(parts, " ")
}

func ZoneKindPromptSeed(zoneKind *ZoneKind) string {
	if zoneKind == nil {
		return ""
	}
	if description := strings.TrimSpace(zoneKind.Description); description != "" {
		return description
	}
	label := strings.TrimSpace(ZoneKindPromptLabel(zoneKind))
	if label == "" {
		return ""
	}
	return fmt.Sprintf(
		"Keep the content naturally suited to %s zones.",
		strings.ToLower(label),
	)
}

func ZoneKindsPromptOptions(zoneKinds []ZoneKind) string {
	if len(zoneKinds) == 0 {
		return "(none provided)"
	}

	lines := make([]string, 0, len(zoneKinds))
	seen := map[string]struct{}{}
	for i := range zoneKinds {
		slug := ZoneKindPromptSlug(&zoneKinds[i])
		if slug == "" {
			continue
		}
		if _, exists := seen[slug]; exists {
			continue
		}
		seen[slug] = struct{}{}

		label := strings.TrimSpace(ZoneKindPromptLabel(&zoneKinds[i]))
		if label == "" {
			label = slug
		}
		seed := strings.TrimSpace(ZoneKindPromptSeed(&zoneKinds[i]))
		if seed == "" {
			lines = append(lines, fmt.Sprintf("- %s: %s", slug, label))
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s: %s (%s)", slug, label, seed))
	}

	if len(lines) == 0 {
		return "(none provided)"
	}
	return strings.Join(lines, "\n")
}
