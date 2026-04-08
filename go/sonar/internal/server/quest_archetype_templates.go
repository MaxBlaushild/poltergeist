package server

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type questArchetypeItemRewardPayload struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type questArchetypeSpellRewardPayload struct {
	SpellID string `json:"spellId"`
}

func normalizeQuestTemplateAcceptanceDialogue(
	input []models.DialogueMessage,
) models.DialogueSequence {
	return models.DialogueSequence(input)
}

func dialogueSequenceFromLines(input []string) models.DialogueSequence {
	return models.DialogueSequenceFromStringLines(input)
}

func normalizeExplicitQuestTemplateContent(
	name string,
	description string,
	acceptanceDialogue []models.DialogueMessage,
) (string, string, models.DialogueSequence, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", "", nil, fmt.Errorf("name is required")
	}
	trimmedDescription := strings.TrimSpace(description)
	if trimmedDescription == "" {
		return "", "", nil, fmt.Errorf("description is required")
	}
	normalizedDialogue := normalizeQuestTemplateAcceptanceDialogue(acceptanceDialogue)
	if len(normalizedDialogue) == 0 {
		return "", "", nil, fmt.Errorf("acceptanceDialogue must include at least one line")
	}
	return trimmedName, trimmedDescription, normalizedDialogue, nil
}

func questArchetypeNodeRequiresChallengeTemplate(node *models.QuestArchetypeNode) bool {
	if node == nil {
		return false
	}
	return models.NormalizeQuestArchetypeNodeType(string(node.NodeType)) == models.QuestArchetypeNodeTypeChallenge
}

func (s *server) validateQuestArchetypeChallengeTemplate(
	ctx context.Context,
	node *models.QuestArchetypeNode,
	templateID *uuid.UUID,
	requireTemplate bool,
) (*models.ChallengeTemplate, error) {
	if templateID == nil || *templateID == uuid.Nil {
		if requireTemplate {
			return nil, fmt.Errorf("challengeTemplateId is required for challenge nodes")
		}
		return nil, nil
	}
	if node != nil &&
		models.NormalizeQuestArchetypeNodeType(string(node.NodeType)) != models.QuestArchetypeNodeTypeChallenge {
		return nil, fmt.Errorf("challengeTemplateId can only be used on challenge nodes")
	}
	template, err := s.dbClient.ChallengeTemplate().FindByID(ctx, *templateID)
	if err != nil {
		return nil, fmt.Errorf("challengeTemplateId could not be loaded")
	}
	if template == nil {
		return nil, fmt.Errorf("challengeTemplateId could not be loaded")
	}
	if node != nil &&
		node.LocationArchetypeID != nil &&
		*node.LocationArchetypeID != uuid.Nil &&
		template.LocationArchetypeID != *node.LocationArchetypeID {
		return nil, fmt.Errorf("challengeTemplateId must match the node location archetype")
	}
	return template, nil
}

func normalizeQuestTemplateCharacterTags(input []string) models.StringArray {
	return parseInventoryInternalTags(input)
}

func normalizeQuestTemplateInternalTags(input []string) models.StringArray {
	return parseInventoryInternalTags(input)
}

func normalizeQuestTemplateRecurrenceFrequency(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	recurrence := models.NormalizeQuestRecurrenceFrequency(*value)
	if recurrence == "" {
		return nil, nil
	}
	if !models.IsValidQuestRecurrenceFrequency(recurrence) {
		return nil, fmt.Errorf("invalid recurrence frequency")
	}
	return &recurrence, nil
}

func (s *server) normalizeQuestArchetypeQuestGiverCharacterID(
	ctx context.Context,
	category string,
	rawID *uuid.UUID,
) (*uuid.UUID, error) {
	if rawID == nil || *rawID == uuid.Nil {
		if models.IsMainStoryQuestCategory(category) {
			return nil, fmt.Errorf("main story quest archetypes require questGiverCharacterId")
		}
		return nil, nil
	}

	character, err := s.dbClient.Character().FindByID(ctx, *rawID)
	if err != nil {
		return nil, fmt.Errorf("questGiverCharacterId could not be loaded")
	}
	if character == nil {
		return nil, fmt.Errorf("questGiverCharacterId could not be loaded")
	}

	return rawID, nil
}

func buildQuestArchetypeItemRewards(
	questArchetypeID uuid.UUID,
	payloads []questArchetypeItemRewardPayload,
) []models.QuestArchetypeItemReward {
	rewards := make([]models.QuestArchetypeItemReward, 0, len(payloads))
	now := time.Now()
	for _, reward := range payloads {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		rewards = append(rewards, models.QuestArchetypeItemReward{
			ID:               uuid.New(),
			CreatedAt:        now,
			UpdatedAt:        now,
			QuestArchetypeID: questArchetypeID,
			InventoryItemID:  reward.InventoryItemID,
			Quantity:         reward.Quantity,
		})
	}
	return rewards
}

func buildQuestArchetypeSpellRewards(
	questArchetypeID uuid.UUID,
	payloads []questArchetypeSpellRewardPayload,
) ([]models.QuestArchetypeSpellReward, error) {
	rewards := make([]models.QuestArchetypeSpellReward, 0, len(payloads))
	now := time.Now()
	seen := map[uuid.UUID]struct{}{}
	for idx, reward := range payloads {
		spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
		if err != nil || spellID == uuid.Nil {
			return nil, fmt.Errorf("spellRewards[%d].spellId must be a valid UUID", idx)
		}
		if _, exists := seen[spellID]; exists {
			continue
		}
		seen[spellID] = struct{}{}
		rewards = append(rewards, models.QuestArchetypeSpellReward{
			ID:               uuid.New(),
			CreatedAt:        now,
			UpdatedAt:        now,
			QuestArchetypeID: questArchetypeID,
			SpellID:          spellID,
		})
	}
	return rewards, nil
}

func (s *server) resolveQuestTemplateCharacterID(
	ctx context.Context,
	zoneID uuid.UUID,
	questArchetype *models.QuestArchetype,
) (*uuid.UUID, error) {
	if questArchetype != nil && questArchetype.QuestGiverCharacterID != nil && *questArchetype.QuestGiverCharacterID != uuid.Nil {
		return questArchetype.QuestGiverCharacterID, nil
	}
	if questArchetype != nil && models.IsMainStoryQuestCategory(questArchetype.Category) {
		return nil, fmt.Errorf("main story quest archetype requires questGiverCharacterId")
	}
	if questArchetype == nil || len(questArchetype.CharacterTags) == 0 {
		return nil, nil
	}

	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		return nil, err
	}
	if zone == nil {
		return nil, nil
	}

	pointsOfInterest, err := s.dbClient.PointOfInterest().FindAllForZone(ctx, zoneID)
	if err != nil {
		return nil, err
	}
	pointOfInterestIDs := make(map[uuid.UUID]struct{}, len(pointsOfInterest))
	for _, poi := range pointsOfInterest {
		pointOfInterestIDs[poi.ID] = struct{}{}
	}

	characters, err := s.dbClient.Character().FindAll(ctx)
	if err != nil {
		return nil, err
	}

	desiredTags := make(map[string]struct{}, len(questArchetype.CharacterTags))
	for _, tag := range questArchetype.CharacterTags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		desiredTags[normalized] = struct{}{}
	}
	if len(desiredTags) == 0 {
		return nil, nil
	}

	type candidate struct {
		character *models.Character
		priority  int
	}
	candidates := make([]candidate, 0)
	for _, character := range characters {
		if character == nil || !characterMatchesQuestTemplateTags(character, desiredTags) {
			continue
		}
		if character.PointOfInterestID != nil {
			if _, ok := pointOfInterestIDs[*character.PointOfInterestID]; ok {
				candidates = append(candidates, candidate{character: character, priority: 0})
				continue
			}
		}
		if characterInZoneBoundary(zone, character) {
			candidates = append(candidates, candidate{character: character, priority: 1})
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority < candidates[j].priority
		}
		iName := strings.ToLower(strings.TrimSpace(candidates[i].character.Name))
		jName := strings.ToLower(strings.TrimSpace(candidates[j].character.Name))
		if iName != jName {
			return iName < jName
		}
		return candidates[i].character.ID.String() < candidates[j].character.ID.String()
	})

	selectedID := candidates[0].character.ID
	return &selectedID, nil
}

func (s *server) resolveMainStoryQuestTemplateCharacterID(
	ctx context.Context,
	rawTags []string,
) (*uuid.UUID, error) {
	normalizedTags := normalizeQuestTemplateCharacterTags(rawTags)
	if len(normalizedTags) == 0 {
		return nil, fmt.Errorf("main story quest archetypes require a specific quest giver")
	}

	characters, err := s.dbClient.Character().FindAll(ctx)
	if err != nil {
		return nil, err
	}

	desiredTags := make(map[string]struct{}, len(normalizedTags))
	for _, tag := range normalizedTags {
		trimmed := strings.ToLower(strings.TrimSpace(tag))
		if trimmed == "" {
			continue
		}
		desiredTags[trimmed] = struct{}{}
	}
	if len(desiredTags) == 0 {
		return nil, fmt.Errorf("main story quest archetypes require a specific quest giver")
	}

	matches := make([]*models.Character, 0)
	for _, character := range characters {
		if character == nil || !characterMatchesQuestTemplateTags(character, desiredTags) {
			continue
		}
		matches = append(matches, character)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf(
			"could not resolve a specific quest giver from character tags: %s",
			strings.Join([]string(normalizedTags), ", "),
		)
	}

	sort.Slice(matches, func(i, j int) bool {
		iName := strings.ToLower(strings.TrimSpace(matches[i].Name))
		jName := strings.ToLower(strings.TrimSpace(matches[j].Name))
		if iName != jName {
			return iName < jName
		}
		return matches[i].ID.String() < matches[j].ID.String()
	})

	if len(matches) > 1 {
		names := make([]string, 0, len(matches))
		for _, match := range matches {
			names = append(names, strings.TrimSpace(match.Name))
		}
		return nil, fmt.Errorf(
			"character tags resolve to multiple possible quest givers (%s); choose a specific character",
			strings.Join(names, ", "),
		)
	}

	selectedID := matches[0].ID
	return &selectedID, nil
}

func (s *server) resolveZoneQuestArchetypeCharacterID(
	ctx context.Context,
	zoneQuestArchetype *models.ZoneQuestArchetype,
) (*uuid.UUID, error) {
	if zoneQuestArchetype == nil {
		return nil, nil
	}

	questArchetype := &zoneQuestArchetype.QuestArchetype
	if questArchetype == nil || questArchetype.ID == uuid.Nil {
		loaded, err := s.dbClient.QuestArchetype().FindByID(ctx, zoneQuestArchetype.QuestArchetypeID)
		if err != nil {
			return nil, err
		}
		questArchetype = loaded
	}
	if questArchetype != nil && models.IsMainStoryQuestCategory(questArchetype.Category) {
		if questArchetype.QuestGiverCharacterID != nil && *questArchetype.QuestGiverCharacterID != uuid.Nil {
			return questArchetype.QuestGiverCharacterID, nil
		}
		return nil, fmt.Errorf("main story quest archetype requires questGiverCharacterId")
	}
	if zoneQuestArchetype.CharacterID != nil {
		return zoneQuestArchetype.CharacterID, nil
	}
	return s.resolveQuestTemplateCharacterID(ctx, zoneQuestArchetype.ZoneID, questArchetype)
}

func characterMatchesQuestTemplateTags(character *models.Character, desiredTags map[string]struct{}) bool {
	if character == nil || len(character.InternalTags) == 0 || len(desiredTags) == 0 {
		return false
	}
	for _, tag := range character.InternalTags {
		if _, ok := desiredTags[strings.ToLower(strings.TrimSpace(tag))]; ok {
			return true
		}
	}
	return false
}

func characterInZoneBoundary(zone *models.Zone, character *models.Character) bool {
	if zone == nil || character == nil {
		return false
	}
	for _, location := range character.Locations {
		if zone.IsPointInBoundary(location.Latitude, location.Longitude) {
			return true
		}
	}
	return false
}
