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
- Return JSON only.

Respond as:
{
  "templates": [
    {
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

const minimumQuestMonsterTemplateMatchScore = 2

type questMonsterTemplateRequest struct {
	Count             int
	MonsterType       models.MonsterTemplateType
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
		if score < minimumQuestMonsterTemplateMatchScore {
			continue
		}
		scored = append(scored, questMonsterTemplateMatch{template: template, score: score})
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if scored[i].template.MonsterType != scored[j].template.MonsterType {
			return scored[i].template.MonsterType < scored[j].template.MonsterType
		}
		return strings.ToLower(scored[i].template.Name) < strings.ToLower(scored[j].template.Name)
	})
	if len(scored) > limit {
		scored = scored[:limit]
	}
	return scored
}

func buildQuestMonsterTemplateQuerySet(request questMonsterTemplateRequest) map[string]struct{} {
	queryTokens := generatedQuestTemplateTokens(request.ThemePrompt)
	queryTokens = append(queryTokens, generatedQuestTemplateTokens(request.EncounterConcept)...)
	queryTokens = append(queryTokens, generatedQuestTemplateTokens(request.LocationConcept)...)
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

	specs, err := s.buildQuestSpecificMonsterTemplateSpecs(count, usedNames, existingNames, request)
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
		if err := s.dbClient.MonsterTemplate().Create(ctx, template); err != nil {
			return nil, fmt.Errorf("failed to create quest-specific monster template: %w", err)
		}
		created = append(created, *template)
	}
	return created, nil
}

func (s *server) buildQuestSpecificMonsterTemplateSpecs(
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
		fallback := buildBulkMonsterTemplateSpecsFromSeeds(remaining, usedNames, request.MonsterType)
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
