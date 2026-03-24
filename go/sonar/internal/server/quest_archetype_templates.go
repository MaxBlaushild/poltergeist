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

func normalizeQuestTemplateAcceptanceDialogue(input []string) models.StringArray {
	lines := make(models.StringArray, 0, len(input))
	for _, line := range input {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lines = append(lines, trimmed)
	}
	if lines == nil {
		return models.StringArray{}
	}
	return lines
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

func (s *server) resolveZoneQuestArchetypeCharacterID(
	ctx context.Context,
	zoneQuestArchetype *models.ZoneQuestArchetype,
) (*uuid.UUID, error) {
	if zoneQuestArchetype == nil {
		return nil, nil
	}
	if zoneQuestArchetype.CharacterID != nil {
		return zoneQuestArchetype.CharacterID, nil
	}

	questArchetype := &zoneQuestArchetype.QuestArchetype
	if questArchetype == nil || questArchetype.ID == uuid.Nil {
		loaded, err := s.dbClient.QuestArchetype().FindByID(ctx, zoneQuestArchetype.QuestArchetypeID)
		if err != nil {
			return nil, err
		}
		questArchetype = loaded
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
