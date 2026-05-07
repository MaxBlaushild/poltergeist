package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const questArchetypeGeneratedLocationArchetypePromptTemplate = `
You are designing one reusable location archetype for StreetSekai, an urban fantasy MMORPG layered onto real-world places.

This archetype is needed because a generated quest beat referenced a location flavor that did not cleanly match an existing archetype.

Quest context:
- preferred zone kind: %s
- draft name: %s
- draft hook: %s
- draft description: %s
- location concept: %s
- suggested location archetype name: %s
- location metadata tags: %s
- template concept: %s
- potential content: %s

Allowed Google place types:
%s

Existing archetype names to avoid duplicating:
%s

Return JSON only:
{
  "name": "2-5 word reusable archetype name",
  "includedTypes": ["exact_google_place_type"],
  "excludedTypes": ["exact_google_place_type"]
}

Rules:
- Output JSON only. No markdown.
- Create one reusable archetype that would actually help route a player to a fitting real-world place.
- Favor a name close to the suggested archetype or location concept when they are already good.
- includedTypes must contain 1-6 exact Google place types from the allowed list.
- excludedTypes may contain 0-6 exact Google place types from the allowed list.
- Keep excludedTypes sparse and only use them when they sharpen the concept.
- Do not create a one-off proper noun or quest-specific name.
- Make the archetype feel like a place category someone could visit in many cities.
`

type generatedQuestArchetypeLocationArchetypePayload struct {
	Name          string   `json:"name"`
	IncludedTypes []string `json:"includedTypes"`
	ExcludedTypes []string `json:"excludedTypes"`
}

type sanitizedQuestArchetypeGeneratedLocationArchetype struct {
	Name          string
	IncludedTypes googlemaps.PlaceTypeSlice
	ExcludedTypes googlemaps.PlaceTypeSlice
}

func (s *server) ensureQuestArchetypeSuggestionDraftLocationArchetypes(
	ctx context.Context,
	draft *models.QuestArchetypeSuggestionDraft,
	preferredZoneKind *models.ZoneKind,
	locationArchetypes []*models.LocationArchetype,
	requiredLocationArchetypeIDs []string,
) (bool, error) {
	if draft == nil {
		return false, nil
	}
	changed := repairQuestArchetypeSuggestionDraftLocationArchetypes(
		draft,
		locationArchetypes,
		requiredLocationArchetypeIDs,
	)

	if len(draft.Nodes) > 0 {
		for index := range draft.Nodes {
			node := &draft.Nodes[index]
			if node.Source != "location" {
				continue
			}
			if node.LocationArchetypeID != nil && *node.LocationArchetypeID != uuid.Nil {
				continue
			}
			archetype, err := s.ensureQuestArchetypeSuggestionGeneratedLocationArchetype(
				ctx,
				questArchetypeSuggestionStepFromNode(*node),
				draft,
				preferredZoneKind,
				locationArchetypes,
			)
			if err != nil {
				return changed, err
			}
			if archetype == nil {
				continue
			}
			locationID := archetype.ID
			node.LocationArchetypeID = &locationID
			node.LocationArchetypeName = strings.TrimSpace(archetype.Name)
			locationArchetypes = append(locationArchetypes, archetype)
			changed = true
		}
		if changed {
			draft.Steps = make(models.QuestArchetypeSuggestionSteps, 0, len(draft.Nodes))
			for _, node := range draft.Nodes {
				draft.Steps = append(draft.Steps, questArchetypeSuggestionStepFromNode(node))
			}
		}
		return changed, nil
	}

	for index := range draft.Steps {
		step := &draft.Steps[index]
		if step.Source != "location" {
			continue
		}
		if step.LocationArchetypeID != nil && *step.LocationArchetypeID != uuid.Nil {
			continue
		}
		archetype, err := s.ensureQuestArchetypeSuggestionGeneratedLocationArchetype(
			ctx,
			*step,
			draft,
			preferredZoneKind,
			locationArchetypes,
		)
		if err != nil {
			return changed, err
		}
		if archetype == nil {
			continue
		}
		locationID := archetype.ID
		step.LocationArchetypeID = &locationID
		step.LocationArchetypeName = strings.TrimSpace(archetype.Name)
		locationArchetypes = append(locationArchetypes, archetype)
		changed = true
	}

	return changed, nil
}

func (s *server) ensureQuestArchetypeSuggestionGeneratedLocationArchetype(
	ctx context.Context,
	step models.QuestArchetypeSuggestionStep,
	draft *models.QuestArchetypeSuggestionDraft,
	preferredZoneKind *models.ZoneKind,
	locationArchetypes []*models.LocationArchetype,
) (*models.LocationArchetype, error) {
	if s == nil || s.deepPriest == nil {
		return nil, fmt.Errorf("%s step %q is missing a resolved location archetype and no archetype generator is available", step.Content, step.LocationConcept)
	}

	prompt := buildQuestArchetypeGeneratedLocationArchetypePrompt(
		step,
		draft,
		preferredZoneKind,
		locationArchetypes,
	)
	answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return nil, fmt.Errorf("failed to generate location archetype for %q: %w", step.LocationConcept, err)
	}

	var generated generatedQuestArchetypeLocationArchetypePayload
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &generated); err != nil {
		return nil, fmt.Errorf("failed to parse generated location archetype for %q: %w", step.LocationConcept, err)
	}

	sanitized, ok := sanitizeQuestArchetypeGeneratedLocationArchetypePayload(generated)
	if !ok {
		return nil, fmt.Errorf("generated location archetype for %q did not include a valid reusable name and place types", step.LocationConcept)
	}

	if existing := findQuestArchetypeSuggestionExistingLocationArchetypeMatch(sanitized, locationArchetypes); existing != nil {
		return existing, nil
	}

	now := time.Now()
	archetype := &models.LocationArchetype{
		ID:            uuid.New(),
		Name:          sanitized.Name,
		CreatedAt:     now,
		UpdatedAt:     now,
		IncludedTypes: sanitized.IncludedTypes,
		ExcludedTypes: sanitized.ExcludedTypes,
		Challenges:    models.LocationArchetypeChallenges{},
	}
	if err := s.dbClient.LocationArchetype().Create(ctx, archetype); err != nil {
		return nil, fmt.Errorf("failed to create generated location archetype %q: %w", archetype.Name, err)
	}
	return archetype, nil
}

func buildQuestArchetypeGeneratedLocationArchetypePrompt(
	step models.QuestArchetypeSuggestionStep,
	draft *models.QuestArchetypeSuggestionDraft,
	preferredZoneKind *models.ZoneKind,
	locationArchetypes []*models.LocationArchetype,
) string {
	allowedPlaceTypes := googlemaps.GetAllPlaceTypes()
	allowedNames := make([]string, 0, len(allowedPlaceTypes))
	for _, placeType := range allowedPlaceTypes {
		allowedNames = append(allowedNames, string(placeType))
	}
	sort.Strings(allowedNames)

	return fmt.Sprintf(
		questArchetypeGeneratedLocationArchetypePromptTemplate,
		renderQuestArchetypeGeneratedLocationArchetypeZoneKind(preferredZoneKind),
		questArchetypeGeneratedLocationArchetypeQuotedOrNone(draftNameOrEmpty(draft)),
		questArchetypeGeneratedLocationArchetypeQuotedOrNone(draftHookOrEmpty(draft)),
		questArchetypeGeneratedLocationArchetypeQuotedOrNone(draftDescriptionOrEmpty(draft)),
		questArchetypeGeneratedLocationArchetypeQuotedOrNone(step.LocationConcept),
		questArchetypeGeneratedLocationArchetypeQuotedOrNone(step.LocationArchetypeName),
		renderQuestArchetypeGeneratedLocationArchetypeTagList(step.LocationMetadataTags),
		questArchetypeGeneratedLocationArchetypeQuotedOrNone(step.TemplateConcept),
		questArchetypeGeneratedLocationArchetypeQuotedOrNone(strings.Join(step.PotentialContent, ", ")),
		strings.Join(allowedNames, ", "),
		joinQuestArchetypeGeneratedLocationArchetypeNames(locationArchetypes, 160),
	)
}

func sanitizeQuestArchetypeGeneratedLocationArchetypePayload(
	raw generatedQuestArchetypeLocationArchetypePayload,
) (sanitizedQuestArchetypeGeneratedLocationArchetype, bool) {
	name := normalizeQuestArchetypeGeneratedLocationArchetypeName(raw.Name)
	if name == "" {
		return sanitizedQuestArchetypeGeneratedLocationArchetype{}, false
	}
	allowedTypeIndex := buildQuestArchetypeGeneratedLocationArchetypePlaceTypeIndex(googlemaps.GetAllPlaceTypes())
	included := sanitizeQuestArchetypeGeneratedLocationArchetypePlaceTypes(raw.IncludedTypes, allowedTypeIndex, 6)
	if len(included) == 0 {
		return sanitizedQuestArchetypeGeneratedLocationArchetype{}, false
	}
	excluded := sanitizeQuestArchetypeGeneratedLocationArchetypePlaceTypes(raw.ExcludedTypes, allowedTypeIndex, 6)
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
	return sanitizedQuestArchetypeGeneratedLocationArchetype{
		Name:          name,
		IncludedTypes: sortQuestArchetypeGeneratedLocationPlaceTypes(included),
		ExcludedTypes: sortQuestArchetypeGeneratedLocationPlaceTypes(excluded),
	}, true
}

func findQuestArchetypeSuggestionExistingLocationArchetypeMatch(
	spec sanitizedQuestArchetypeGeneratedLocationArchetype,
	locationArchetypes []*models.LocationArchetype,
) *models.LocationArchetype {
	if len(locationArchetypes) == 0 {
		return nil
	}
	nameKey := normalizeQuestArchetypeGeneratedLocationArchetypeNameKey(spec.Name)
	signature := buildQuestArchetypeGeneratedLocationArchetypeSignature(spec.IncludedTypes, spec.ExcludedTypes)
	for _, archetype := range locationArchetypes {
		if archetype == nil {
			continue
		}
		if nameKey != "" &&
			nameKey == normalizeQuestArchetypeGeneratedLocationArchetypeNameKey(archetype.Name) {
			return archetype
		}
		if signature != "" &&
			signature == buildQuestArchetypeGeneratedLocationArchetypeSignature(archetype.IncludedTypes, archetype.ExcludedTypes) {
			return archetype
		}
	}
	return nil
}

func buildQuestArchetypeGeneratedLocationArchetypePlaceTypeIndex(
	placeTypes []googlemaps.PlaceType,
) map[string]string {
	index := make(map[string]string, len(placeTypes))
	for _, placeType := range placeTypes {
		canonical := string(placeType)
		index[normalizeQuestArchetypeGeneratedLocationPlaceTypeKey(canonical)] = canonical
	}
	return index
}

func sanitizeQuestArchetypeGeneratedLocationArchetypePlaceTypes(
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
		key := normalizeQuestArchetypeGeneratedLocationPlaceTypeKey(candidate)
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

func sortQuestArchetypeGeneratedLocationPlaceTypes(
	items googlemaps.PlaceTypeSlice,
) googlemaps.PlaceTypeSlice {
	if len(items) == 0 {
		return nil
	}
	sorted := append(googlemaps.PlaceTypeSlice(nil), items...)
	sort.Slice(sorted, func(i, j int) bool {
		return string(sorted[i]) < string(sorted[j])
	})
	return sorted
}

func buildQuestArchetypeGeneratedLocationArchetypeSignature(
	includedTypes googlemaps.PlaceTypeSlice,
	excludedTypes googlemaps.PlaceTypeSlice,
) string {
	included := sortQuestArchetypeGeneratedLocationPlaceTypes(includedTypes)
	if len(included) == 0 {
		return ""
	}
	excluded := sortQuestArchetypeGeneratedLocationPlaceTypes(excludedTypes)

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

func normalizeQuestArchetypeGeneratedLocationArchetypeName(raw string) string {
	return questArchetypeGeneratedLocationArchetypeCollapseWhitespace(raw)
}

func normalizeQuestArchetypeGeneratedLocationArchetypeNameKey(raw string) string {
	cleaned := questArchetypeGeneratedLocationArchetypeCollapseWhitespace(strings.ToLower(raw))
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

func normalizeQuestArchetypeGeneratedLocationPlaceTypeKey(raw string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return ""
	}
	trimmed = strings.ReplaceAll(trimmed, "-", "_")
	trimmed = strings.ReplaceAll(trimmed, " ", "_")
	return trimmed
}

func joinQuestArchetypeGeneratedLocationArchetypeNames(
	locationArchetypes []*models.LocationArchetype,
	limit int,
) string {
	if len(locationArchetypes) == 0 {
		return "none"
	}
	names := make([]string, 0, len(locationArchetypes))
	seen := map[string]struct{}{}
	for _, archetype := range locationArchetypes {
		if archetype == nil {
			continue
		}
		name := questArchetypeGeneratedLocationArchetypeCollapseWhitespace(archetype.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, name)
	}
	sort.Strings(names)
	if limit > 0 && len(names) > limit {
		names = names[:limit]
	}
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, "; ")
}

func questArchetypeGeneratedLocationArchetypeCollapseWhitespace(raw string) string {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func questArchetypeGeneratedLocationArchetypeQuotedOrNone(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "none"
	}
	return trimmed
}

func renderQuestArchetypeGeneratedLocationArchetypeZoneKind(
	zoneKind *models.ZoneKind,
) string {
	if zoneKind == nil {
		return "none"
	}
	label := strings.TrimSpace(models.ZoneKindPromptLabel(zoneKind))
	slug := strings.TrimSpace(models.ZoneKindPromptSlug(zoneKind))
	if label != "" && slug != "" {
		return fmt.Sprintf("%s (%s)", label, slug)
	}
	if label != "" {
		return label
	}
	if slug != "" {
		return slug
	}
	return "none"
}

func renderQuestArchetypeGeneratedLocationArchetypeTagList(tags []string) string {
	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		parts = append(parts, trimmed)
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

func draftNameOrEmpty(draft *models.QuestArchetypeSuggestionDraft) string {
	if draft == nil {
		return ""
	}
	return draft.Name
}

func draftHookOrEmpty(draft *models.QuestArchetypeSuggestionDraft) string {
	if draft == nil {
		return ""
	}
	return draft.Hook
}

func draftDescriptionOrEmpty(draft *models.QuestArchetypeSuggestionDraft) string {
	if draft == nil {
		return ""
	}
	return draft.Description
}
