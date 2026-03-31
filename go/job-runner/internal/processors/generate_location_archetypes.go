package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const locationArchetypeGenerationPromptTemplate = `
You are designing %d reusable location archetypes for a fantasy MMORPG layered onto real-world places.

Your job is to create destination categories that would feel genuinely appealing, atmospheric, and worth visiting for a player walking around in the real world.

Creative salt for this run: %s
Use that salt as a nudge toward a fresh angle, mood, and mix of destination types so this batch feels meaningfully different from past runs.

Allowed Google place types:
%s

Existing archetype names to avoid repeating:
%s

Existing included/excluded type combinations to avoid repeating:
%s

New names already planned in this batch:
%s

New included/excluded type combinations already planned in this batch:
%s

Return JSON only:
{
  "archetypes": [
    {
      "name": "2-5 word evocative archetype name",
      "includedTypes": ["1-6 exact place type values from the allowed list"],
      "excludedTypes": ["0-6 exact place type values from the allowed list"]
    }
  ]
}

Hard rules:
- Output exactly %d archetypes.
- Favor places that are charming, social, scenic, relaxing, delicious, surprising, playful, or culturally rich.
- Make the set diverse across food, parks, arts, nightlife, shopping, landmarks, sports, and cozy everyday discoveries.
- These must be reusable categories, not specific businesses, cities, or one-off landmarks.
- Every includedTypes and excludedTypes entry must match an allowed Google place type exactly.
- Do not reuse the same includedTypes/excludedTypes combination with a different name.
- Keep excludedTypes sparse and only use them when they sharpen the concept.
- The name should sound like a destination category a player would want to pursue.
`

type GenerateLocationArchetypesProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
}

type generatedLocationArchetypePayload struct {
	Name          string   `json:"name"`
	IncludedTypes []string `json:"includedTypes"`
	ExcludedTypes []string `json:"excludedTypes"`
}

type generatedLocationArchetypesResponse struct {
	Archetypes []generatedLocationArchetypePayload `json:"archetypes"`
}

type sanitizedGeneratedLocationArchetype struct {
	Name          string
	IncludedTypes googlemaps.PlaceTypeSlice
	ExcludedTypes googlemaps.PlaceTypeSlice
}

func NewGenerateLocationArchetypesProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
) GenerateLocationArchetypesProcessor {
	log.Println("Initializing GenerateLocationArchetypesProcessor")
	return GenerateLocationArchetypesProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateLocationArchetypesProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate location archetypes task: %v", task.Type())

	var payload jobs.GenerateLocationArchetypesTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	count := payload.Count
	if count <= 0 {
		count = 50
	}
	if count > 100 {
		count = 100
	}

	salt := collapseWhitespace(payload.Salt)
	if salt == "" {
		salt = fmt.Sprintf("auto-%d", time.Now().UnixNano())
	}

	return p.generateLocationArchetypes(ctx, count, salt)
}

func (p *GenerateLocationArchetypesProcessor) generateLocationArchetypes(ctx context.Context, count int, salt string) error {
	existing, err := p.dbClient.LocationArchetype().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load existing location archetypes: %w", err)
	}

	existingNameKeys := make(map[string]struct{}, len(existing))
	existingNames := make([]string, 0, len(existing))
	existingSignatures := make(map[string]struct{}, len(existing))
	existingSignatureLabels := make([]string, 0, len(existing))
	for _, archetype := range existing {
		if archetype == nil {
			continue
		}
		name := collapseWhitespace(archetype.Name)
		if name == "" {
			continue
		}
		existingNameKeys[normalizeLocationArchetypeNameKey(name)] = struct{}{}
		existingNames = append(existingNames, name)
		signature := buildLocationArchetypeSignature(archetype.IncludedTypes, archetype.ExcludedTypes)
		if signature != "" {
			existingSignatures[signature] = struct{}{}
			existingSignatureLabels = append(existingSignatureLabels, humanizeLocationArchetypeSignature(signature))
		}
	}
	sort.Strings(existingNames)
	sort.Strings(existingSignatureLabels)

	allowedPlaceTypes := googlemaps.GetAllPlaceTypes()
	allowedIndex := buildLocationArchetypePlaceTypeIndex(allowedPlaceTypes)
	allowedNames := make([]string, 0, len(allowedPlaceTypes))
	for _, placeType := range allowedPlaceTypes {
		allowedNames = append(allowedNames, string(placeType))
	}

	plannedNameKeys := make(map[string]struct{}, len(existingNameKeys)+count)
	for key := range existingNameKeys {
		plannedNameKeys[key] = struct{}{}
	}
	plannedSignatures := make(map[string]struct{}, len(existingSignatures)+count)
	for key := range existingSignatures {
		plannedSignatures[key] = struct{}{}
	}
	planned := make([]sanitizedGeneratedLocationArchetype, 0, count)

	for attempt := 0; attempt < 4 && len(planned) < count; attempt++ {
		remaining := count - len(planned)
		attemptSalt := fmt.Sprintf("%s|attempt-%d", salt, attempt+1)
		prompt := fmt.Sprintf(
			locationArchetypeGenerationPromptTemplate,
			remaining,
			attemptSalt,
			strings.Join(allowedNames, ", "),
			joinLocationArchetypeAvoidanceNames(existingNames, 200),
			joinLocationArchetypeAvoidanceNames(existingSignatureLabels, 200),
			joinLocationArchetypeAvoidanceNames(sanitizedLocationArchetypeNames(planned), 100),
			joinLocationArchetypeAvoidanceNames(sanitizedLocationArchetypeSignatures(planned), 100),
			remaining,
		)

		answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
		if err != nil {
			return fmt.Errorf("failed to generate location archetypes: %w", err)
		}

		var generated generatedLocationArchetypesResponse
		if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &generated); err != nil {
			return fmt.Errorf("failed to parse generated location archetypes payload: %w", err)
		}

		sanitized := sanitizeGeneratedLocationArchetypes(generated.Archetypes, allowedIndex, remaining)
		for _, spec := range sanitized {
			nameKey := normalizeLocationArchetypeNameKey(spec.Name)
			if nameKey == "" {
				continue
			}
			signature := buildLocationArchetypeSignature(spec.IncludedTypes, spec.ExcludedTypes)
			if signature == "" {
				continue
			}
			if _, exists := plannedNameKeys[nameKey]; exists {
				continue
			}
			if _, exists := plannedSignatures[signature]; exists {
				continue
			}
			plannedNameKeys[nameKey] = struct{}{}
			plannedSignatures[signature] = struct{}{}
			planned = append(planned, spec)
			if len(planned) >= count {
				break
			}
		}
	}

	if len(planned) == 0 {
		return fmt.Errorf("no valid location archetypes were generated")
	}

	createdCount := 0
	now := time.Now()
	for _, spec := range planned {
		archetype := &models.LocationArchetype{
			ID:            uuid.New(),
			Name:          spec.Name,
			CreatedAt:     now,
			UpdatedAt:     now,
			IncludedTypes: spec.IncludedTypes,
			ExcludedTypes: spec.ExcludedTypes,
			Challenges:    models.LocationArchetypeChallenges{},
		}
		if err := p.dbClient.LocationArchetype().Create(ctx, archetype); err != nil {
			return fmt.Errorf("failed to create location archetype %q: %w", spec.Name, err)
		}
		createdCount++
	}

	log.Printf(
		"GenerateLocationArchetypesProcessor created %d location archetypes (requested %d)",
		createdCount,
		count,
	)
	return nil
}

func sanitizeGeneratedLocationArchetypes(
	raw []generatedLocationArchetypePayload,
	allowedTypeIndex map[string]string,
	limit int,
) []sanitizedGeneratedLocationArchetype {
	if limit <= 0 {
		return nil
	}

	result := make([]sanitizedGeneratedLocationArchetype, 0, limit)
	seenNames := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		name := collapseWhitespace(item.Name)
		if name == "" {
			continue
		}
		nameKey := normalizeLocationArchetypeNameKey(name)
		if nameKey == "" {
			continue
		}
		if _, exists := seenNames[nameKey]; exists {
			continue
		}

		included := sanitizeGeneratedLocationArchetypePlaceTypes(item.IncludedTypes, allowedTypeIndex, 6)
		if len(included) == 0 {
			continue
		}
		excluded := sanitizeGeneratedLocationArchetypePlaceTypes(item.ExcludedTypes, allowedTypeIndex, 6)
		if len(excluded) > 0 {
			includedSet := make(map[string]struct{}, len(included))
			for _, placeType := range included {
				includedSet[string(placeType)] = struct{}{}
			}
			filteredExcluded := make(googlemaps.PlaceTypeSlice, 0, len(excluded))
			for _, placeType := range excluded {
				if _, overlap := includedSet[string(placeType)]; overlap {
					continue
				}
				filteredExcluded = append(filteredExcluded, placeType)
			}
			excluded = filteredExcluded
		}
		included = sortPlaceTypes(included)
		excluded = sortPlaceTypes(excluded)

		seenNames[nameKey] = struct{}{}
		result = append(result, sanitizedGeneratedLocationArchetype{
			Name:          name,
			IncludedTypes: included,
			ExcludedTypes: excluded,
		})
		if len(result) >= limit {
			break
		}
	}
	return result
}

func sortPlaceTypes(items googlemaps.PlaceTypeSlice) googlemaps.PlaceTypeSlice {
	if len(items) == 0 {
		return nil
	}
	sorted := append(googlemaps.PlaceTypeSlice(nil), items...)
	sort.Slice(sorted, func(i, j int) bool {
		return string(sorted[i]) < string(sorted[j])
	})
	return sorted
}

func sanitizeGeneratedLocationArchetypePlaceTypes(
	raw []string,
	allowedTypeIndex map[string]string,
	limit int,
) googlemaps.PlaceTypeSlice {
	if limit <= 0 {
		return nil
	}

	result := make(googlemaps.PlaceTypeSlice, 0, limit)
	seen := make(map[string]struct{}, len(raw))
	for _, candidate := range raw {
		key := normalizeGeneratedPlaceTypeKey(candidate)
		if key == "" {
			continue
		}
		canonical, ok := allowedTypeIndex[key]
		if !ok {
			continue
		}
		if _, exists := seen[canonical]; exists {
			continue
		}
		seen[canonical] = struct{}{}
		result = append(result, googlemaps.PlaceType(canonical))
		if len(result) >= limit {
			break
		}
	}
	return result
}

func buildLocationArchetypePlaceTypeIndex(placeTypes []googlemaps.PlaceType) map[string]string {
	index := make(map[string]string, len(placeTypes))
	for _, placeType := range placeTypes {
		canonical := string(placeType)
		index[normalizeGeneratedPlaceTypeKey(canonical)] = canonical
	}
	return index
}

func normalizeGeneratedPlaceTypeKey(raw string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return ""
	}
	trimmed = strings.ReplaceAll(trimmed, "-", "_")
	trimmed = strings.ReplaceAll(trimmed, " ", "_")
	return trimmed
}

func normalizeLocationArchetypeNameKey(raw string) string {
	cleaned := collapseWhitespace(strings.ToLower(raw))
	if cleaned == "" {
		return ""
	}

	var builder strings.Builder
	lastWasSpace := false
	for _, r := range cleaned {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(r)
			lastWasSpace = false
		case unicode.IsSpace(r):
			if !lastWasSpace && builder.Len() > 0 {
				builder.WriteByte(' ')
				lastWasSpace = true
			}
		default:
			if !lastWasSpace && builder.Len() > 0 {
				builder.WriteByte(' ')
				lastWasSpace = true
			}
		}
	}

	return strings.TrimSpace(builder.String())
}

func collapseWhitespace(raw string) string {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func joinLocationArchetypeAvoidanceNames(names []string, limit int) string {
	if len(names) == 0 {
		return "none"
	}
	if limit > 0 && len(names) > limit {
		names = names[:limit]
	}
	return strings.Join(names, "; ")
}

func sanitizedLocationArchetypeNames(items []sanitizedGeneratedLocationArchetype) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if item.Name == "" {
			continue
		}
		names = append(names, item.Name)
	}
	return names
}

func sanitizedLocationArchetypeSignatures(items []sanitizedGeneratedLocationArchetype) []string {
	signatures := make([]string, 0, len(items))
	for _, item := range items {
		signature := buildLocationArchetypeSignature(item.IncludedTypes, item.ExcludedTypes)
		if signature == "" {
			continue
		}
		signatures = append(signatures, humanizeLocationArchetypeSignature(signature))
	}
	return signatures
}

func buildLocationArchetypeSignature(
	includedTypes googlemaps.PlaceTypeSlice,
	excludedTypes googlemaps.PlaceTypeSlice,
) string {
	included := sortPlaceTypes(includedTypes)
	if len(included) == 0 {
		return ""
	}
	excluded := sortPlaceTypes(excludedTypes)

	includedParts := make([]string, 0, len(included))
	for _, item := range included {
		includedParts = append(includedParts, string(item))
	}
	excludedParts := make([]string, 0, len(excluded))
	for _, item := range excluded {
		excludedParts = append(excludedParts, string(item))
	}

	return "include:" + strings.Join(includedParts, ",") + "|exclude:" + strings.Join(excludedParts, ",")
}

func humanizeLocationArchetypeSignature(signature string) string {
	if signature == "" {
		return ""
	}
	parts := strings.SplitN(signature, "|exclude:", 2)
	if len(parts) != 2 {
		return signature
	}
	included := strings.TrimPrefix(parts[0], "include:")
	excluded := parts[1]
	if excluded == "" {
		return "include [" + included + "]"
	}
	return "include [" + included + "] exclude [" + excluded + "]"
}
