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

func buildZoneNameDiversityGuidance(
	ctx context.Context,
	dbClient db.DbClient,
	currentZoneID uuid.UUID,
) string {
	zones, err := dbClient.Zone().FindAll(ctx)
	if err != nil {
		log.Printf("Failed to load zones for naming guidance: %v", err)
		return "- Existing zone names unavailable.\n- Avoid repeating the same opening word or cadence across multiple names."
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

		firstWord := normalizedZoneLeadingWord(name)
		if firstWord != "" {
			leadingWordCounts[firstWord]++
		}
	}

	overusedLeadingWords := make([]string, 0)
	for word, count := range leadingWordCounts {
		if count >= 2 {
			overusedLeadingWords = append(overusedLeadingWords, word)
		}
	}
	sort.Strings(overusedLeadingWords)

	lines := []string{}
	if len(existingNames) > 0 {
		lines = append(lines, fmt.Sprintf("- Existing zone names: %s", strings.Join(existingNames, ", ")))
	} else {
		lines = append(lines, "- Existing zone names: none yet")
	}

	if len(overusedLeadingWords) > 0 {
		lines = append(
			lines,
			fmt.Sprintf("- Avoid starting the new name with these overused opening words: %s", strings.Join(overusedLeadingWords, ", ")),
		)
	} else {
		lines = append(lines, "- No repeated opening word is currently dominant, but still avoid cliché repeated prefixes.")
	}

	return strings.Join(lines, "\n")
}

func normalizedZoneLeadingWord(name string) string {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return ""
	}

	word := strings.Trim(parts[0], ".,;:!?\"'`()[]{}")
	return strings.ToLower(word)
}
