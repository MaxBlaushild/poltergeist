package server

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	questlogpkg "github.com/MaxBlaushild/poltergeist/sonar/internal/questlog"
	"github.com/google/uuid"
)

func (s *server) hydrateQuestLogQuestGivers(
	ctx context.Context,
	userID uuid.UUID,
	questLog *questlogpkg.QuestLog,
) error {
	if questLog == nil {
		return nil
	}

	characterIDs := collectQuestLogQuestGiverCharacterIDs(questLog)
	if len(characterIDs) == 0 {
		return nil
	}

	activeStoryFlags, err := s.loadUserStoryFlagMap(ctx, userID)
	if err != nil {
		return err
	}
	relationshipMap, err := s.loadUserCharacterRelationshipMap(
		ctx,
		userID,
		characterIDs,
	)
	if err != nil {
		return err
	}

	characterByID := make(map[uuid.UUID]*models.Character, len(characterIDs))
	loadCharacter := func(characterID uuid.UUID) (*models.Character, error) {
		if characterID == uuid.Nil {
			return nil, nil
		}
		if cached, ok := characterByID[characterID]; ok {
			return cached, nil
		}
		character, err := s.dbClient.Character().FindByID(ctx, characterID)
		if err != nil || character == nil {
			return nil, err
		}

		characterCopy := *character
		if character.PointOfInterest != nil {
			pointOfInterestCopy := *character.PointOfInterest
			applyPointOfInterestStoryVariant(&pointOfInterestCopy, activeStoryFlags)
			characterCopy.PointOfInterest = &pointOfInterestCopy
		}
		applyCharacterStoryVariant(&characterCopy, activeStoryFlags)
		applyCharacterRelationship(&characterCopy, relationshipMap)
		characterByID[characterID] = &characterCopy
		return &characterCopy, nil
	}

	for index := range questLog.Quests {
		if err := hydrateQuestLogQuestGiverForEntry(
			&questLog.Quests[index],
			loadCharacter,
		); err != nil {
			return err
		}
	}
	for index := range questLog.CompletedQuests {
		if err := hydrateQuestLogQuestGiverForEntry(
			&questLog.CompletedQuests[index],
			loadCharacter,
		); err != nil {
			return err
		}
	}

	return nil
}

func collectQuestLogQuestGiverCharacterIDs(
	questLog *questlogpkg.QuestLog,
) []uuid.UUID {
	if questLog == nil {
		return nil
	}

	seen := make(map[uuid.UUID]struct{})
	ids := make([]uuid.UUID, 0, len(questLog.Quests)+len(questLog.CompletedQuests))
	appendID := func(raw *uuid.UUID) {
		if raw == nil || *raw == uuid.Nil {
			return
		}
		if _, exists := seen[*raw]; exists {
			return
		}
		seen[*raw] = struct{}{}
		ids = append(ids, *raw)
	}

	for _, quest := range questLog.Quests {
		appendID(quest.QuestGiverCharacterID)
	}
	for _, quest := range questLog.CompletedQuests {
		appendID(quest.QuestGiverCharacterID)
	}

	return ids
}

func hydrateQuestLogQuestGiverForEntry(
	entry *questlogpkg.Quest,
	loadCharacter func(uuid.UUID) (*models.Character, error),
) error {
	if entry == nil || entry.QuestGiverCharacterID == nil {
		return nil
	}

	character, err := loadCharacter(*entry.QuestGiverCharacterID)
	if err != nil || character == nil {
		return err
	}

	entry.QuestGiverCharacter = character
	if character.PointOfInterest != nil {
		pointOfInterestCopy := *character.PointOfInterest
		entry.QuestGiverPointOfInterest = &pointOfInterestCopy
	}

	return nil
}
