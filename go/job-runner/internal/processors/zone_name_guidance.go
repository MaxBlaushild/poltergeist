package processors

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/google/uuid"
)

type zoneNameDiversityContext struct {
	Guidance              string
	ForbiddenLeadingRoots []string
}

func buildZoneNameDiversityContext(
	ctx context.Context,
	dbClient db.DbClient,
	currentZoneID uuid.UUID,
) zoneNameDiversityContext {
	zones, err := dbClient.Zone().FindAll(ctx)
	if err != nil {
		log.Printf("Failed to load zones for naming guidance: %v", err)
		return zoneNameDiversityContext{
			Guidance: "- Existing zone names unavailable.\n- Avoid repeating the same opening word or cadence across multiple names.",
		}
	}

	sort.Slice(zones, func(i, j int) bool {
		return zones[i].CreatedAt.After(zones[j].CreatedAt)
	})

	existingNames := make([]string, 0, 8)
	leadingWordCounts := map[string]int{}
	seenNames := map[string]struct{}{}

	for _, zone := range zones {
		if zone == nil || zone.ID == currentZoneID {
			continue
		}

		name := strings.Join(strings.Fields(strings.TrimSpace(zone.Name)), " ")
		if name == "" {
			continue
		}
		if _, exists := seenNames[strings.ToLower(name)]; exists {
			continue
		}
		seenNames[strings.ToLower(name)] = struct{}{}

		if len(existingNames) < 8 {
			existingNames = append(existingNames, name)
		}

		firstRoot := normalizedZoneLeadingRoot(name)
		if firstRoot != "" {
			leadingWordCounts[firstRoot]++
		}
	}

	overusedLeadingRoots := make([]string, 0)
	for word, count := range leadingWordCounts {
		if count >= 2 {
			overusedLeadingRoots = append(overusedLeadingRoots, word)
		}
	}
	sort.Strings(overusedLeadingRoots)

	lines := []string{}
	if len(existingNames) > 0 {
		lines = append(lines, fmt.Sprintf("- Existing zone names: %s", strings.Join(existingNames, ", ")))
	} else {
		lines = append(lines, "- Existing zone names: none yet")
	}

	if len(overusedLeadingRoots) > 0 {
		lines = append(
			lines,
			fmt.Sprintf("- Avoid starting the new name with these overused opening roots or variants: %s", strings.Join(overusedLeadingRoots, ", ")),
		)
	} else {
		lines = append(lines, "- No repeated opening word is currently dominant, but still avoid cliché repeated prefixes.")
	}

	return zoneNameDiversityContext{
		Guidance:              strings.Join(lines, "\n"),
		ForbiddenLeadingRoots: overusedLeadingRoots,
	}
}

func normalizedZoneLeadingRoot(name string) string {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return ""
	}

	word := strings.Trim(parts[0], ".,;:!?\"'`()[]{}")
	return canonicalizeZoneNameRoot(word)
}

func canonicalizeZoneNameRoot(word string) string {
	word = strings.ToLower(strings.TrimSpace(word))
	if word == "" {
		return ""
	}
	if strings.HasPrefix(word, "whisper") {
		return "whisper"
	}
	if len(word) > 6 && strings.HasSuffix(word, "ing") {
		word = strings.TrimSuffix(word, "ing")
	}
	if len(word) > 5 && strings.HasSuffix(word, "ers") {
		word = strings.TrimSuffix(word, "ers")
	}
	if len(word) > 4 && strings.HasSuffix(word, "es") {
		word = strings.TrimSuffix(word, "es")
	} else if len(word) > 3 && strings.HasSuffix(word, "s") {
		word = strings.TrimSuffix(word, "s")
	}
	return word
}

func zoneNameUsesForbiddenLeadingRoot(name string, forbiddenRoots []string) bool {
	if len(forbiddenRoots) == 0 {
		return false
	}

	root := normalizedZoneLeadingRoot(name)
	if root == "" {
		return false
	}

	for _, forbidden := range forbiddenRoots {
		if root == forbidden {
			return true
		}
	}
	return false
}
