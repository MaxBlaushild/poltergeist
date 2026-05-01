package server

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const questSpecificMonsterTemplatesPromptTemplate = `
You are designing %d original %s monster templates for a fantasy action RPG.

These monsters are specifically being generated to support a reusable quest archetype.

Quest context:
- Theme prompt: %s
- Encounter concept: %s
- Location concept: %s
- Location archetype: %s
- Location place types: %s
- Encounter tone tags: %s
- Seed hints: %s

Zone kind classification:
%s

Template role guidance:
%s

Avoid these existing monster template names:
%s

Hard constraints:
- Output exactly %d templates.
- Use unique names (2-4 words) that are NOT in the existing names list.
- Keep descriptions concise and practical (8-18 words), focused on monster behavior/combat role.
- Make the monsters feel naturally suited to the quest context, place type, and encounter tone.
- Do not reference DnD, tabletop, or copyrighted franchises.
- All base stats must be integers from 1 to 20.
- Every template must include a "zoneKind" slug chosen from the allowed zone kinds.
- Return JSON only.

Respond as:
{
  "templates": [
    {
      "zoneKind": "string",
      "name": "string",
      "description": "string",
      "baseStrength": 10,
      "baseDexterity": 10,
      "baseConstitution": 10,
      "baseIntelligence": 10,
      "baseWisdom": 10,
      "baseCharisma": 10
    }
  ]
}
`

const (
	minimumQuestMonsterTemplateMatchScore = 2
	questMonsterTemplateVeryRecentWindow  = 7 * 24 * time.Hour
	questMonsterTemplateRecentWindow      = 30 * 24 * time.Hour
	questMonsterTemplateAgingWindow       = 90 * 24 * time.Hour
)

type questMonsterTemplateRequest struct {
	Count             int
	MonsterType       models.MonsterTemplateType
	PreferredZoneKind *models.ZoneKind
	ThemePrompt       string
	EncounterConcept  string
	LocationConcept   string
	LocationArchetype *models.LocationArchetype
	EncounterTone     []string
	SeedHints         []string
}

type questMonsterTemplateMatch struct {
	template models.MonsterTemplate
	score    int
}

func (s *server) ensureQuestMonsterTemplateIDs(
	ctx context.Context,
	templates *[]models.MonsterTemplate,
	request questMonsterTemplateRequest,
	existingIDs []string,
) ([]string, error) {
	if templates == nil {
		return nil, fmt.Errorf("monster template pool is required")
	}
	request.Count = maxInt(1, request.Count)
	request.MonsterType = models.NormalizeMonsterTemplateType(string(request.MonsterType))
	if request.MonsterType == "" {
		request.MonsterType = models.MonsterTemplateTypeMonster
	}

	selected := make([]string, 0, request.Count)
	seen := map[string]struct{}{}
	templatesByID := make(map[string]models.MonsterTemplate, len(*templates))
	for _, template := range *templates {
		templatesByID[template.ID.String()] = template
	}
	for _, rawID := range existingIDs {
		templateID := strings.TrimSpace(rawID)
		if templateID == "" {
			continue
		}
		if _, ok := templatesByID[templateID]; !ok {
			continue
		}
		if _, exists := seen[templateID]; exists {
			continue
		}
		seen[templateID] = struct{}{}
		selected = append(selected, templateID)
		if len(selected) >= request.Count {
			return selected[:request.Count], nil
		}
	}

	for _, match := range selectQuestMonsterTemplateMatches(*templates, request, request.Count) {
		templateID := match.template.ID.String()
		if _, exists := seen[templateID]; exists {
			continue
		}
		seen[templateID] = struct{}{}
		selected = append(selected, templateID)
		if len(selected) >= request.Count {
			return selected[:request.Count], nil
		}
	}

	missingCount := request.Count - len(selected)
	if missingCount <= 0 {
		return selected, nil
	}

	createdTemplates, err := s.createQuestSpecificMonsterTemplates(ctx, *templates, request, missingCount)
	if err != nil {
		return nil, err
	}
	*templates = append(*templates, createdTemplates...)
	for _, template := range createdTemplates {
		selected = append(selected, template.ID.String())
		if len(selected) >= request.Count {
			break
		}
	}
	return selected, nil
}

func selectQuestMonsterTemplateMatches(
	templates []models.MonsterTemplate,
	request questMonsterTemplateRequest,
	limit int,
) []questMonsterTemplateMatch {
	if len(templates) == 0 || limit <= 0 {
		return nil
	}
	querySet := buildQuestMonsterTemplateQuerySet(request)
	scored := make([]questMonsterTemplateMatch, 0, len(templates))
	for _, template := range templates {
		score := 0
		for _, token := range generatedQuestTemplateTokens(template.Name + " " + template.Description) {
			if _, ok := querySet[token]; ok {
				score++
			}
		}
		if template.MonsterType == request.MonsterType {
			score++
		}
		if preferredZoneKind := models.ZoneKindPromptSlug(request.PreferredZoneKind); preferredZoneKind != "" &&
			models.NormalizeZoneKind(template.ZoneKind) == preferredZoneKind {
			score += 3
		}
		score -= questMonsterTemplateFreshnessPenalty(template.CreatedAt)
		if score < minimumQuestMonsterTemplateMatchScore {
			continue
		}
		scored = append(scored, questMonsterTemplateMatch{template: template, score: score})
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		leftPenalty := questMonsterTemplateFreshnessPenalty(scored[i].template.CreatedAt)
		rightPenalty := questMonsterTemplateFreshnessPenalty(scored[j].template.CreatedAt)
		if leftPenalty != rightPenalty {
			return leftPenalty < rightPenalty
		}
		if scored[i].template.MonsterType != scored[j].template.MonsterType {
			return scored[i].template.MonsterType < scored[j].template.MonsterType
		}
		if !scored[i].template.CreatedAt.Equal(scored[j].template.CreatedAt) {
			return scored[i].template.CreatedAt.Before(scored[j].template.CreatedAt)
		}
		return strings.ToLower(scored[i].template.Name) < strings.ToLower(scored[j].template.Name)
	})
	if len(scored) > limit {
		scored = scored[:limit]
	}
	return scored
}

func questMonsterTemplateFreshnessPenalty(createdAt time.Time) int {
	if createdAt.IsZero() {
		return 0
	}
	age := time.Since(createdAt)
	switch {
	case age < questMonsterTemplateVeryRecentWindow:
		return 4
	case age < questMonsterTemplateRecentWindow:
		return 2
	case age < questMonsterTemplateAgingWindow:
		return 1
	default:
		return 0
	}
}

func buildQuestMonsterTemplateQuerySet(request questMonsterTemplateRequest) map[string]struct{} {
	queryTokens := generatedQuestTemplateTokens(request.ThemePrompt)
	queryTokens = append(queryTokens, generatedQuestTemplateTokens(request.EncounterConcept)...)
	queryTokens = append(queryTokens, generatedQuestTemplateTokens(request.LocationConcept)...)
	if request.PreferredZoneKind != nil {
		queryTokens = append(
			queryTokens,
			generatedQuestTemplateTokens(models.ZoneKindPromptLabel(request.PreferredZoneKind))...,
		)
		queryTokens = append(
			queryTokens,
			generatedQuestTemplateTokens(models.ZoneKindPromptSlug(request.PreferredZoneKind))...,
		)
		queryTokens = append(
			queryTokens,
			generatedQuestTemplateTokens(models.ZoneKindPromptSeed(request.PreferredZoneKind))...,
		)
	}
	for _, tone := range request.EncounterTone {
		queryTokens = append(queryTokens, generatedQuestTemplateTokens(tone)...)
	}
	for _, seed := range request.SeedHints {
		queryTokens = append(queryTokens, generatedQuestTemplateTokens(seed)...)
	}
	if request.LocationArchetype != nil {
		queryTokens = append(queryTokens, generatedQuestTemplateTokens(request.LocationArchetype.Name)...)
		for _, placeType := range request.LocationArchetype.IncludedTypes {
			queryTokens = append(queryTokens, generatedQuestTemplateTokens(string(placeType))...)
		}
	}
	querySet := map[string]struct{}{}
	for _, token := range queryTokens {
		querySet[token] = struct{}{}
	}
	return querySet
}

func (s *server) createQuestSpecificMonsterTemplates(
	ctx context.Context,
	existingTemplates []models.MonsterTemplate,
	request questMonsterTemplateRequest,
	count int,
) ([]models.MonsterTemplate, error) {
	if count <= 0 {
		return []models.MonsterTemplate{}, nil
	}

	usedNames := make(map[string]struct{}, len(existingTemplates)+count)
	existingNames := make([]string, 0, len(existingTemplates))
	for _, template := range existingTemplates {
		name := strings.TrimSpace(template.Name)
		if name == "" {
			continue
		}
		normalized := strings.ToLower(name)
		usedNames[normalized] = struct{}{}
		existingNames = append(existingNames, name)
	}

	specs, err := s.buildQuestSpecificMonsterTemplateSpecs(ctx, count, usedNames, existingNames, request)
	if err != nil {
		return nil, err
	}

	created := make([]models.MonsterTemplate, 0, len(specs))
	now := time.Now()
	for _, spec := range specs {
		emptyError := ""
		template := &models.MonsterTemplate{
			ID:                    uuid.New(),
			CreatedAt:             now,
			UpdatedAt:             now,
			Archived:              false,
			MonsterType:           models.NormalizeMonsterTemplateType(spec.MonsterType),
			ZoneKind:              models.NormalizeZoneKind(spec.ZoneKind),
			Name:                  strings.TrimSpace(spec.Name),
			Description:           strings.TrimSpace(spec.Description),
			BaseStrength:          spec.BaseStrength,
			BaseDexterity:         spec.BaseDexterity,
			BaseConstitution:      spec.BaseConstitution,
			BaseIntelligence:      spec.BaseIntelligence,
			BaseWisdom:            spec.BaseWisdom,
			BaseCharisma:          spec.BaseCharisma,
			ImageGenerationStatus: models.MonsterTemplateImageGenerationStatusNone,
			ImageGenerationError:  &emptyError,
			Spells:                []models.MonsterTemplateSpell{},
			Progressions:          []models.MonsterTemplateProgression{},
		}
		if trimmedGenreID := strings.TrimSpace(spec.GenreID); trimmedGenreID != "" {
			genreID, err := uuid.Parse(trimmedGenreID)
			if err != nil || genreID == uuid.Nil {
				return nil, fmt.Errorf("failed to parse generated monster genre: %s", trimmedGenreID)
			}
			template.GenreID = genreID
		}
		if err := s.dbClient.MonsterTemplate().Create(ctx, template); err != nil {
			return nil, fmt.Errorf("failed to create quest-specific monster template: %w", err)
		}
		created = append(created, *template)
	}
	return created, nil
}

func (s *server) buildQuestSpecificMonsterTemplateSpecs(
	ctx context.Context,
	count int,
	usedNames map[string]struct{},
	existingNames []string,
	request questMonsterTemplateRequest,
) ([]jobs.MonsterTemplateCreationSpec, error) {
	request.Count = maxInt(1, count)
	request.MonsterType = models.NormalizeMonsterTemplateType(string(request.MonsterType))
	if request.MonsterType == "" {
		request.MonsterType = models.MonsterTemplateTypeMonster
	}

	specs := make([]jobs.MonsterTemplateCreationSpec, 0, request.Count)
	if s.deepPriest != nil {
		zoneKinds, err := s.dbClient.ZoneKind().FindAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load zone kinds for monster template generation: %w", err)
		}
		prompt := fmt.Sprintf(
			questSpecificMonsterTemplatesPromptTemplate,
			request.Count,
			monsterTemplateTypePromptLabel(request.MonsterType),
			questMonsterPromptValue(request.ThemePrompt),
			questMonsterPromptValue(request.EncounterConcept),
			questMonsterPromptValue(request.LocationConcept),
			questMonsterPromptLocationName(request.LocationArchetype),
			questMonsterPromptLocationTypes(request.LocationArchetype),
			questMonsterPromptList(request.EncounterTone),
			questMonsterPromptList(request.SeedHints),
			questMonsterZoneKindPromptBlock(zoneKinds, request.PreferredZoneKind),
			monsterTemplateTypePromptGuidance(request.MonsterType),
			formatMonsterTemplateNamesForPrompt(existingNames),
			request.Count,
		)
		answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
		if err == nil {
			candidates, parseErr := parseGeneratedMonsterTemplates(answer.Answer)
			if parseErr == nil {
				for _, candidate := range candidates {
					if len(specs) >= request.Count {
						break
					}
					candidate = sanitizeMonsterTemplateSpec(candidate)
					candidate.MonsterType = string(request.MonsterType)
					candidate.ZoneKind = normalizeGeneratedMonsterTemplateZoneKind(
						candidate.ZoneKind,
						zoneKinds,
						request.PreferredZoneKind,
					)
					if candidate.Name == "" {
						continue
					}
					normalized := strings.ToLower(candidate.Name)
					if _, exists := usedNames[normalized]; exists {
						continue
					}
					usedNames[normalized] = struct{}{}
					existingNames = append(existingNames, candidate.Name)
					specs = append(specs, candidate)
				}
			}
		}
	}

	if remaining := request.Count - len(specs); remaining > 0 {
		fallback := buildQuestMonsterFallbackSpecsFromRequest(
			remaining,
			usedNames,
			request,
		)
		specs = append(specs, fallback...)
	}
	if len(specs) == 0 {
		return nil, fmt.Errorf("failed to prepare quest-specific monster templates")
	}
	if len(specs) > request.Count {
		specs = specs[:request.Count]
	}
	return specs, nil
}

func buildQuestMonsterFallbackSpecsFromRequest(
	count int,
	usedNames map[string]struct{},
	request questMonsterTemplateRequest,
) []jobs.MonsterTemplateCreationSpec {
	specs := make([]jobs.MonsterTemplateCreationSpec, 0, count)
	if count <= 0 {
		return specs
	}
	orderedSeedPool := orderQuestMonsterFallbackSeedPool(
		genericMonsterTemplateRoleSeeds,
		request,
		usedNames,
	)
	if len(orderedSeedPool) == 0 {
		return specs
	}
	prefixes := questMonsterFallbackPrefixes(request)
	preferredZoneKind := models.ZoneKindPromptSlug(request.PreferredZoneKind)
	for index := 0; index < count; index++ {
		seed := orderedSeedPool[index%len(orderedSeedPool)]
		baseName := strings.TrimSpace(seed.Name)
		if len(prefixes) > 0 {
			baseName = fmt.Sprintf("%s %s", prefixes[index%len(prefixes)], baseName)
		}
		specs = append(specs, jobs.MonsterTemplateCreationSpec{
			MonsterType:      string(request.MonsterType),
			ZoneKind:         preferredZoneKind,
			Name:             nextUniqueMonsterTemplateName(baseName, usedNames, prefixes),
			Description:      buildMonsterTemplateFallbackDescription(seed, nil, request.PreferredZoneKind),
			BaseStrength:     seed.BaseStrength,
			BaseDexterity:    seed.BaseDexterity,
			BaseConstitution: seed.BaseConstitution,
			BaseIntelligence: seed.BaseIntelligence,
			BaseWisdom:       seed.BaseWisdom,
			BaseCharisma:     seed.BaseCharisma,
		})
	}
	return specs
}

func orderQuestMonsterFallbackSeedPool(
	seedPool []dndMonsterTemplateSeed,
	request questMonsterTemplateRequest,
	usedNames map[string]struct{},
) []dndMonsterTemplateSeed {
	ordered := append([]dndMonsterTemplateSeed(nil), seedPool...)
	querySet := buildQuestMonsterTemplateQuerySet(request)
	sort.SliceStable(ordered, func(left, right int) bool {
		leftScore := questMonsterFallbackSeedScore(ordered[left], querySet)
		rightScore := questMonsterFallbackSeedScore(ordered[right], querySet)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		leftUsage := monsterTemplateSeedUsageCount(ordered[left].Name, usedNames)
		rightUsage := monsterTemplateSeedUsageCount(ordered[right].Name, usedNames)
		if leftUsage != rightUsage {
			return leftUsage < rightUsage
		}
		return ordered[left].Name < ordered[right].Name
	})
	return ordered
}

func questMonsterFallbackSeedScore(
	seed dndMonsterTemplateSeed,
	querySet map[string]struct{},
) int {
	score := 0
	for _, token := range generatedQuestTemplateTokens(seed.Name + " " + seed.Description) {
		if _, ok := querySet[token]; ok {
			score++
		}
	}
	return score
}

func questMonsterFallbackPrefixes(
	request questMonsterTemplateRequest,
) []string {
	prefixes := monsterTemplateFallbackPrefixes(nil, request.PreferredZoneKind)
	seen := make(map[string]struct{}, len(prefixes))
	for _, prefix := range prefixes {
		seen[strings.ToLower(strings.TrimSpace(prefix))] = struct{}{}
	}
	appendWords := func(raw string) {
		for _, word := range strings.Fields(strings.TrimSpace(raw)) {
			trimmed := strings.Trim(word, " -_,.;:!?/\\")
			if len(trimmed) < 4 {
				continue
			}
			if _, blocked := monsterTemplateFallbackPrefixStopWords[strings.ToLower(trimmed)]; blocked {
				continue
			}
			appendUniqueMonsterTemplatePrefix(
				&prefixes,
				seen,
				questMonsterFallbackPrefixLabel(trimmed),
			)
		}
	}

	for _, tone := range request.EncounterTone {
		appendWords(tone)
	}
	appendWords(request.LocationConcept)
	appendWords(request.EncounterConcept)
	if request.LocationArchetype != nil {
		appendWords(request.LocationArchetype.Name)
		for _, placeType := range request.LocationArchetype.IncludedTypes {
			appendWords(string(placeType))
		}
	}
	for _, seedHint := range request.SeedHints {
		appendWords(seedHint)
	}
	appendWords(request.ThemePrompt)
	return prefixes
}

func questMonsterFallbackPrefixLabel(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	if len(lower) == 1 {
		return strings.ToUpper(lower)
	}
	return strings.ToUpper(lower[:1]) + lower[1:]
}

func questMonsterPromptValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "none"
	}
	return trimmed
}

func questMonsterPromptList(values []string) string {
	parts := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		parts = append(parts, trimmed)
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

func questMonsterZoneKindPromptBlock(
	zoneKinds []models.ZoneKind,
	preferred *models.ZoneKind,
) string {
	lines := make([]string, 0, len(zoneKinds)+8)
	if preferred != nil {
		if label := strings.TrimSpace(models.ZoneKindPromptLabel(preferred)); label != "" {
			lines = append(lines, fmt.Sprintf("- preferred zone kind: %s", label))
		}
		if slug := strings.TrimSpace(models.ZoneKindPromptSlug(preferred)); slug != "" {
			lines = append(lines, fmt.Sprintf("- preferred zone kind slug: %s", slug))
		}
		if seed := strings.TrimSpace(models.ZoneKindPromptSeed(preferred)); seed != "" {
			lines = append(lines, fmt.Sprintf("- preferred zone kind creative seed: %s", seed))
		}
		lines = append(lines, "")
	}
	lines = append(lines, "Allowed zone kinds:", models.ZoneKindsPromptOptions(zoneKinds))
	lines = append(lines, "", "Additional rules:")
	lines = append(lines, "- Return zoneKind as one allowed slug exactly as written.")
	lines = append(lines, "- Choose the strongest environmental fit for the quest encounter and location context.")
	if preferred != nil {
		lines = append(lines, "- If the preferred zone kind is still a strong fit, keep it.")
	}
	return strings.Join(lines, "\n")
}

func questMonsterPromptLocationName(location *models.LocationArchetype) string {
	if location == nil {
		return "none"
	}
	return questMonsterPromptValue(location.Name)
}

func questMonsterPromptLocationTypes(location *models.LocationArchetype) string {
	if location == nil || len(location.IncludedTypes) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(location.IncludedTypes))
	for _, placeType := range location.IncludedTypes {
		parts = append(parts, string(placeType))
	}
	return strings.Join(parts, ", ")
}
