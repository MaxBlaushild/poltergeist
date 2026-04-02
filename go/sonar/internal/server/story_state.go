package server

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func normalizeStoryFlagKeys(input []string) models.StringArray {
	return parseInventoryInternalTags(input)
}

func userStoryFlagMap(flags []models.UserStoryFlag) map[string]bool {
	result := map[string]bool{}
	for _, flag := range flags {
		key := strings.ToLower(strings.TrimSpace(flag.FlagKey))
		if key == "" {
			continue
		}
		result[key] = flag.Value
	}
	return result
}

func (s *server) loadUserStoryFlagMap(
	ctx context.Context,
	userID uuid.UUID,
) (map[string]bool, error) {
	flags, err := s.dbClient.UserStoryFlag().FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return userStoryFlagMap(flags), nil
}

func hasRequiredStoryFlags(required []string, active map[string]bool) bool {
	if len(required) == 0 {
		return true
	}
	for _, raw := range required {
		key := strings.ToLower(strings.TrimSpace(raw))
		if key == "" {
			continue
		}
		if !active[key] {
			return false
		}
	}
	return true
}

func scenarioAvailableForStoryFlags(
	scenario *models.Scenario,
	active map[string]bool,
) bool {
	if scenario == nil {
		return false
	}
	return hasRequiredStoryFlags(scenario.RequiredStoryFlags, active)
}

func challengeAvailableForStoryFlags(
	challenge *models.Challenge,
	active map[string]bool,
) bool {
	if challenge == nil {
		return false
	}
	return hasRequiredStoryFlags(challenge.RequiredStoryFlags, active)
}

func monsterEncounterAvailableForStoryFlags(
	encounter *models.MonsterEncounter,
	active map[string]bool,
) bool {
	if encounter == nil {
		return false
	}
	return hasRequiredStoryFlags(encounter.RequiredStoryFlags, active)
}

func (s *server) applyQuestStoryFlagsOnTurnIn(
	ctx context.Context,
	userID uuid.UUID,
	quest *models.Quest,
) error {
	if quest == nil {
		return nil
	}
	for _, flag := range normalizeStoryFlagKeys([]string(quest.SetStoryFlags)) {
		if err := s.dbClient.UserStoryFlag().Upsert(ctx, userID, flag, true); err != nil {
			return err
		}
	}
	for _, flag := range normalizeStoryFlagKeys([]string(quest.ClearStoryFlags)) {
		if err := s.dbClient.UserStoryFlag().DeleteByUserAndKeys(ctx, userID, []string{flag}); err != nil {
			return err
		}
	}
	return nil
}

func matchingCharacterStoryVariant(
	character *models.Character,
	activeFlags map[string]bool,
) *models.CharacterStoryVariant {
	if character == nil || len(character.StoryVariants) == 0 {
		return nil
	}
	var matches []models.CharacterStoryVariant
	for _, variant := range character.StoryVariants {
		if !hasRequiredStoryFlags(variant.RequiredStoryFlags, activeFlags) {
			continue
		}
		matches = append(matches, variant)
	}
	if len(matches) == 0 {
		return nil
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Priority != matches[j].Priority {
			return matches[i].Priority > matches[j].Priority
		}
		if len(matches[i].RequiredStoryFlags) != len(matches[j].RequiredStoryFlags) {
			return len(matches[i].RequiredStoryFlags) > len(matches[j].RequiredStoryFlags)
		}
		return matches[i].UpdatedAt.After(matches[j].UpdatedAt)
	})
	return &matches[0]
}

func applyCharacterStoryVariant(
	character *models.Character,
	activeFlags map[string]bool,
) *models.CharacterStoryVariant {
	variant := matchingCharacterStoryVariant(character, activeFlags)
	if character == nil || variant == nil {
		return nil
	}
	if strings.TrimSpace(variant.Description) != "" {
		character.Description = strings.TrimSpace(variant.Description)
	}
	return variant
}

func matchingPointOfInterestStoryVariant(
	poi *models.PointOfInterest,
	activeFlags map[string]bool,
) *models.PointOfInterestStoryVariant {
	if poi == nil || len(poi.StoryVariants) == 0 {
		return nil
	}
	var matches []models.PointOfInterestStoryVariant
	for _, variant := range poi.StoryVariants {
		if !hasRequiredStoryFlags(variant.RequiredStoryFlags, activeFlags) {
			continue
		}
		matches = append(matches, variant)
	}
	if len(matches) == 0 {
		return nil
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Priority != matches[j].Priority {
			return matches[i].Priority > matches[j].Priority
		}
		if len(matches[i].RequiredStoryFlags) != len(matches[j].RequiredStoryFlags) {
			return len(matches[i].RequiredStoryFlags) > len(matches[j].RequiredStoryFlags)
		}
		return matches[i].UpdatedAt.After(matches[j].UpdatedAt)
	})
	return &matches[0]
}

func applyPointOfInterestStoryVariant(
	poi *models.PointOfInterest,
	activeFlags map[string]bool,
) *models.PointOfInterestStoryVariant {
	variant := matchingPointOfInterestStoryVariant(poi, activeFlags)
	if poi == nil || variant == nil {
		return nil
	}
	if strings.TrimSpace(variant.Description) != "" {
		poi.Description = strings.TrimSpace(variant.Description)
	}
	if strings.TrimSpace(variant.Clue) != "" {
		poi.Clue = strings.TrimSpace(variant.Clue)
	}
	return variant
}

func normalizeCharacterStoryVariants(
	input []models.CharacterStoryVariant,
) models.CharacterStoryVariants {
	now := time.Now()
	variants := make(models.CharacterStoryVariants, 0, len(input))
	for _, raw := range input {
		requiredFlags := normalizeStoryFlagKeys([]string(raw.RequiredStoryFlags))
		description := strings.TrimSpace(raw.Description)
		dialogue := make(models.DialogueSequence, 0, len(raw.Dialogue))
		for _, message := range raw.Dialogue {
			text := strings.TrimSpace(message.Text)
			if text == "" {
				continue
			}
			speaker := strings.TrimSpace(strings.ToLower(message.Speaker))
			if speaker == "" {
				speaker = "character"
			}
			dialogue = append(dialogue, models.DialogueMessage{
				Speaker: speaker,
				Text:    text,
				Order:   len(dialogue),
			})
		}
		if len(requiredFlags) == 0 && description == "" && len(dialogue) == 0 {
			continue
		}
		id := raw.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		createdAt := raw.CreatedAt
		if createdAt.IsZero() {
			createdAt = now
		}
		variants = append(variants, models.CharacterStoryVariant{
			ID:                 id,
			CreatedAt:          createdAt,
			UpdatedAt:          now,
			Priority:           raw.Priority,
			RequiredStoryFlags: requiredFlags,
			Description:        description,
			Dialogue:           dialogue,
		})
	}
	return variants
}

func normalizePointOfInterestStoryVariants(
	input []models.PointOfInterestStoryVariant,
) models.PointOfInterestStoryVariants {
	now := time.Now()
	variants := make(models.PointOfInterestStoryVariants, 0, len(input))
	for _, raw := range input {
		requiredFlags := normalizeStoryFlagKeys([]string(raw.RequiredStoryFlags))
		description := strings.TrimSpace(raw.Description)
		clue := strings.TrimSpace(raw.Clue)
		if len(requiredFlags) == 0 && description == "" && clue == "" {
			continue
		}
		id := raw.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		createdAt := raw.CreatedAt
		if createdAt.IsZero() {
			createdAt = now
		}
		variants = append(variants, models.PointOfInterestStoryVariant{
			ID:                 id,
			CreatedAt:          createdAt,
			UpdatedAt:          now,
			Priority:           raw.Priority,
			RequiredStoryFlags: requiredFlags,
			Description:        description,
			Clue:               clue,
		})
	}
	return variants
}
