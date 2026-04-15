package poilocals

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
)

type ZoneContext struct {
	Name        string
	Description string
}

type PlaceContext struct {
	ID               string
	Name             string
	OriginalName     string
	Description      string
	Address          string
	EditorialSummary string
	Types            []string
}

type CharacterDraft struct {
	Name        string
	Description string
	PlaceID     string
	Dialogue    []string
}

type characterGenerationResponse struct {
	Characters []struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		PlaceID     string   `json:"placeId"`
		Dialogue    []string `json:"dialogue"`
	} `json:"characters"`
}

const characterGenerationPromptTemplate = `
You are designing memorable local NPCs for a fantasy roleplaying game.

Zone name: %s
Zone description: %s

Create exactly the requested number of locals for each point of interest below.
Most places should feel like they have one memorable regular, with a smaller number having two distinct locals.
Each local must:
- feel naturally tied to the point of interest
- include one quirky or vivid detail about the location, OR a made-up detail about their personal life
- have 2 short in-character dialogue lines
- avoid quests, missions, rewards, tutorials, or direct instructions to the player

Points of interest:
%s

Respond ONLY as JSON:
{
  "characters": [
    {
      "name": "string",
      "description": "string",
      "placeId": "string",
      "dialogue": ["string", "string"]
    }
  ]
}
`

func SeedKey(primary string, fallback string) string {
	trimmedPrimary := strings.TrimSpace(primary)
	if trimmedPrimary != "" {
		return trimmedPrimary
	}
	trimmedFallback := strings.TrimSpace(fallback)
	if trimmedFallback != "" {
		return trimmedFallback
	}
	return "poi-local"
}

func DesiredLocalCount(seedKey string) int {
	if stableHash(SeedKey(seedKey, ""))%5 == 0 {
		return 2
	}
	return 1
}

func GenerateDrafts(
	deepPriest deep_priest.DeepPriest,
	zone ZoneContext,
	places []PlaceContext,
) []CharacterDraft {
	if len(places) == 0 {
		return nil
	}

	desiredByPlaceID := map[string]int{}
	placeByID := map[string]PlaceContext{}
	promptLines := make([]string, 0, len(places))
	for _, place := range places {
		placeID := strings.TrimSpace(place.ID)
		if placeID == "" {
			continue
		}
		desiredByPlaceID[placeID] = DesiredLocalCount(SeedKey(place.ID, place.Name))
		placeByID[placeID] = place
		promptLines = append(promptLines, formatPlaceForPrompt(place, desiredByPlaceID[placeID]))
	}
	if len(promptLines) == 0 {
		return nil
	}

	generatedByPlaceID := map[string][]CharacterDraft{}
	if deepPriest != nil {
		prompt := fmt.Sprintf(
			characterGenerationPromptTemplate,
			displayZoneName(zone),
			displayZoneDescription(zone),
			strings.Join(promptLines, "\n"),
		)
		generatedByPlaceID = requestGeneratedDrafts(deepPriest, prompt, placeByID)
	}

	drafts := make([]CharacterDraft, 0, len(places)*2)
	for _, place := range places {
		placeID := strings.TrimSpace(place.ID)
		if placeID == "" {
			continue
		}
		targetCount := desiredByPlaceID[placeID]
		seenNames := map[string]struct{}{}
		validDrafts := make([]CharacterDraft, 0, targetCount)
		for _, draft := range generatedByPlaceID[placeID] {
			sanitized, ok := sanitizeCharacterDraft(draft, place, seenNames)
			if !ok {
				continue
			}
			validDrafts = append(validDrafts, sanitized)
			if len(validDrafts) >= targetCount {
				break
			}
		}
		if len(validDrafts) < targetCount {
			for _, fallback := range FallbackDrafts(zone, place) {
				sanitized, ok := sanitizeCharacterDraft(fallback, place, seenNames)
				if !ok {
					continue
				}
				validDrafts = append(validDrafts, sanitized)
				if len(validDrafts) >= targetCount {
					break
				}
			}
		}
		drafts = append(drafts, validDrafts...)
	}

	return drafts
}

func FallbackDrafts(zone ZoneContext, place PlaceContext) []CharacterDraft {
	targetCount := DesiredLocalCount(SeedKey(place.ID, place.Name))
	drafts := make([]CharacterDraft, 0, targetCount)
	for ordinal := 0; ordinal < targetCount; ordinal++ {
		name := fallbackCharacterName(place, ordinal)
		description := fallbackCharacterDescription(zone, place, ordinal)
		dialogue := fallbackDialogueLines(place, ordinal)
		drafts = append(drafts, CharacterDraft{
			Name:        name,
			Description: description,
			PlaceID:     strings.TrimSpace(place.ID),
			Dialogue:    dialogue,
		})
	}
	return drafts
}

func requestGeneratedDrafts(
	deepPriest deep_priest.DeepPriest,
	prompt string,
	placeByID map[string]PlaceContext,
) map[string][]CharacterDraft {
	generatedByPlaceID := map[string][]CharacterDraft{}
	requestPrompt := prompt
	for attempt := 1; attempt <= 2; attempt++ {
		answer, err := deepPriest.PetitionTheFount(&deep_priest.Question{Question: requestPrompt})
		if err != nil {
			log.Printf("POI local generation failed: %v", err)
			return generatedByPlaceID
		}
		response, err := parseCharacterGenerationResponse(answer.Answer)
		if err == nil {
			for _, raw := range response.Characters {
				placeID := strings.TrimSpace(raw.PlaceID)
				if _, ok := placeByID[placeID]; !ok {
					continue
				}
				generatedByPlaceID[placeID] = append(generatedByPlaceID[placeID], CharacterDraft{
					Name:        raw.Name,
					Description: raw.Description,
					PlaceID:     placeID,
					Dialogue:    raw.Dialogue,
				})
			}
			return generatedByPlaceID
		}
		log.Printf("POI local generation response parse failed: %v", err)
		requestPrompt = prompt + "\n\nReturn ONLY valid JSON with all braces and quotes closed. No markdown."
	}
	return generatedByPlaceID
}

func parseCharacterGenerationResponse(raw string) (characterGenerationResponse, error) {
	var response characterGenerationResponse
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return response, fmt.Errorf("empty response")
	}
	candidate := strings.TrimSpace(extractJSON(trimmed))
	if candidate == "" {
		candidate = trimmed
	}
	if err := json.Unmarshal([]byte(candidate), &response); err != nil {
		return response, err
	}
	return response, nil
}

func sanitizeCharacterDraft(
	draft CharacterDraft,
	place PlaceContext,
	seenNames map[string]struct{},
) (CharacterDraft, bool) {
	name := truncate(strings.TrimSpace(draft.Name), 80)
	if name == "" {
		return CharacterDraft{}, false
	}
	normalizedName := strings.ToLower(name)
	if _, exists := seenNames[normalizedName]; exists {
		return CharacterDraft{}, false
	}
	seenNames[normalizedName] = struct{}{}

	description := truncate(strings.TrimSpace(draft.Description), 280)
	if description == "" {
		description = fallbackCharacterDescription(ZoneContext{}, place, len(seenNames)-1)
	}

	dialogue := sanitizeDialogueLines(draft.Dialogue, fallbackDialogueLines(place, len(seenNames)-1))
	if len(dialogue) == 0 {
		return CharacterDraft{}, false
	}

	return CharacterDraft{
		Name:        name,
		Description: description,
		PlaceID:     strings.TrimSpace(place.ID),
		Dialogue:    dialogue,
	}, true
}

func sanitizeDialogueLines(lines []string, fallback []string) []string {
	blockedTokens := []string{
		"quest", "mission", "reward", "objective", "tutorial", "click", "player",
	}
	seen := map[string]struct{}{}
	sanitized := make([]string, 0, 2)
	for _, raw := range lines {
		line := strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
		if line == "" {
			continue
		}
		if len(line) > 180 {
			line = strings.TrimSpace(line[:180])
		}
		lower := strings.ToLower(line)
		blocked := false
		for _, token := range blockedTokens {
			if strings.Contains(lower, token) {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		sanitized = append(sanitized, line)
		if len(sanitized) >= 2 {
			break
		}
	}
	if len(sanitized) >= 2 {
		return sanitized
	}
	for _, raw := range fallback {
		line := strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		sanitized = append(sanitized, line)
		if len(sanitized) >= 2 {
			break
		}
	}
	return sanitized
}

func formatPlaceForPrompt(place PlaceContext, desiredCount int) string {
	summary := strings.TrimSpace(place.Description)
	if summary == "" {
		summary = strings.TrimSpace(place.EditorialSummary)
	}
	if summary == "" {
		summary = strings.Join(place.Types, ", ")
	}
	if summary == "" {
		summary = "local landmark"
	}

	parts := []string{
		fmt.Sprintf("- placeId=%s", strings.TrimSpace(place.ID)),
		fmt.Sprintf("desiredCharacters=%d", desiredCount),
		fmt.Sprintf("name=%s", truncate(placeDisplayName(place), 80)),
		fmt.Sprintf("summary=%s", truncate(summary, 140)),
	}
	if originalName := strings.TrimSpace(place.OriginalName); originalName != "" {
		parts = append(parts, fmt.Sprintf("originalName=%s", truncate(originalName, 80)))
	}
	if address := strings.TrimSpace(place.Address); address != "" {
		parts = append(parts, fmt.Sprintf("address=%s", truncate(address, 100)))
	}
	if len(place.Types) > 0 {
		parts = append(parts, fmt.Sprintf("types=%s", truncate(strings.Join(place.Types, ","), 120)))
	}
	return strings.Join(parts, " | ")
}

func displayZoneName(zone ZoneContext) string {
	if strings.TrimSpace(zone.Name) == "" {
		return "Unnamed district"
	}
	return strings.TrimSpace(zone.Name)
}

func displayZoneDescription(zone ZoneContext) string {
	if strings.TrimSpace(zone.Description) == "" {
		return "No zone description provided."
	}
	return strings.TrimSpace(zone.Description)
}

func placeDisplayName(place PlaceContext) string {
	if strings.TrimSpace(place.Name) != "" {
		return strings.TrimSpace(place.Name)
	}
	if strings.TrimSpace(place.OriginalName) != "" {
		return strings.TrimSpace(place.OriginalName)
	}
	return "Unnamed point of interest"
}

func fallbackCharacterName(place PlaceContext, ordinal int) string {
	firstNames := []string{
		"Bria", "Cormac", "Delia", "Enzo", "Farah", "Gideon", "Ivo", "Junia",
		"Kestrel", "Lena", "Milo", "Nora", "Orrin", "Pia", "Rafi", "Sabine",
		"Tobin", "Veda", "Willa", "Yorin",
	}
	lastNames := []string{
		"Bell", "Brass", "Clover", "Dawn", "Fable", "Fenn", "Hearth", "Keel",
		"Mallow", "Pike", "Quill", "Reed", "Rook", "Sable", "Thimble", "Vale",
		"Vane", "Wick", "Yarrow", "Zeal",
	}

	hash := stableHash(fmt.Sprintf("%s|%d|name", SeedKey(place.ID, placeDisplayName(place)), ordinal))
	first := firstNames[hash%len(firstNames)]
	last := lastNames[(hash/len(firstNames))%len(lastNames)]
	return first + " " + last
}

func fallbackCharacterDescription(zone ZoneContext, place PlaceContext, ordinal int) string {
	role := pickByKind(place, ordinal,
		[]string{"counter sorcerer", "late-shift regular", "keeper of neighborhood gossip"},
		[]string{"shelf-straightening bibliophile", "quiet annotator", "keeper of borrowed rumors"},
		[]string{"bench-side birdwatcher", "park-path dreamer", "caretaker of small omens"},
		[]string{"gallery murmurer", "ticket-stub collector", "back-row critic"},
		[]string{"market haggler", "window-display perfectionist", "collector of unusual receipts"},
		[]string{"timetable memorizer", "platform philosopher", "rain-soaked commuter oracle"},
	)
	detail := personalDetail(place, ordinal)
	placeName := placeDisplayName(place)
	zoneName := strings.TrimSpace(zone.Name)
	if zoneName == "" {
		return fmt.Sprintf("%s is a %s who lingers around %s and %s.", fallbackCharacterName(place, ordinal), role, placeName, detail)
	}
	return fmt.Sprintf("%s is a %s from %s who treats %s like a second living room and %s.", fallbackCharacterName(place, ordinal), role, zoneName, placeName, detail)
}

func fallbackDialogueLines(place PlaceContext, ordinal int) []string {
	placeName := placeDisplayName(place)
	observations := []string{
		fmt.Sprintf("%s gets interesting once the noise drops and people start telling the truth.", placeName),
		fmt.Sprintf("I can tell what kind of day it will be by the first sound %s makes.", placeName),
		fmt.Sprintf("There is always one corner of %s that feels luckier than the rest.", placeName),
		fmt.Sprintf("I swear %s smells different when someone arrives with a secret.", placeName),
	}
	personal := []string{
		"I repair old umbrellas at night, so I notice every impatient pair of hands.",
		"My sister says I collect trivial facts the way dragons collect gold.",
		"I keep a notebook of overheard promises and grade them for sincerity.",
		"I still send my aunt a postcard after every particularly strange afternoon.",
	}
	hash := stableHash(fmt.Sprintf("%s|%d|dialogue", SeedKey(place.ID, placeName), ordinal))
	return []string{
		observations[hash%len(observations)],
		personal[(hash/len(observations))%len(personal)],
	}
}

func personalDetail(place PlaceContext, ordinal int) string {
	details := []string{
		"still mails elaborate birthday cards to three ex-roommates",
		"swears their best ideas arrive while polishing chipped teacups",
		"keeps a private ranking of the district's most dramatic pigeons",
		"has been teaching a younger cousin how to bluff at cards",
		"writes down every odd superstition they hear before breakfast",
		"never misses the chance to compare strangers' shoes like omens",
	}
	hash := stableHash(fmt.Sprintf("%s|%d|detail", SeedKey(place.ID, placeDisplayName(place)), ordinal))
	return details[hash%len(details)]
}

func pickByKind(
	place PlaceContext,
	ordinal int,
	foodRoles []string,
	bookRoles []string,
	parkRoles []string,
	artsRoles []string,
	shopRoles []string,
	transitRoles []string,
) string {
	rolePool := []string{"street-corner storyteller", "habitual regular", "watchful local"}
	switch placeKind(place) {
	case "food":
		rolePool = foodRoles
	case "books":
		rolePool = bookRoles
	case "park":
		rolePool = parkRoles
	case "arts":
		rolePool = artsRoles
	case "shop":
		rolePool = shopRoles
	case "transit":
		rolePool = transitRoles
	}
	hash := stableHash(fmt.Sprintf("%s|%d|role", SeedKey(place.ID, placeDisplayName(place)), ordinal))
	return rolePool[hash%len(rolePool)]
}

func placeKind(place PlaceContext) string {
	haystack := strings.ToLower(strings.Join([]string{
		place.Name,
		place.OriginalName,
		place.Description,
		place.EditorialSummary,
		strings.Join(place.Types, " "),
	}, " "))
	switch {
	case containsAny(haystack, "cafe", "coffee", "bakery", "restaurant", "bar", "brew", "diner"):
		return "food"
	case containsAny(haystack, "book", "library"):
		return "books"
	case containsAny(haystack, "park", "garden", "trail", "playground"):
		return "park"
	case containsAny(haystack, "museum", "gallery", "cinema", "theater", "music", "art"):
		return "arts"
	case containsAny(haystack, "market", "shop", "store", "boutique"):
		return "shop"
	case containsAny(haystack, "station", "transit", "train", "subway", "bus"):
		return "transit"
	default:
		return "default"
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func stableHash(value string) int {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(strings.TrimSpace(value)))
	return int(hasher.Sum32())
}

func truncate(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if max <= 0 || len(trimmed) <= max {
		return trimmed
	}
	return strings.TrimSpace(trimmed[:max])
}

func extractJSON(raw string) string {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end < start {
		return raw
	}
	return raw[start : end+1]
}
