package server

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func normalizeCharacterRelationshipValue(value int) int {
	if value > 3 {
		return 3
	}
	if value < -3 {
		return -3
	}
	return value
}

func normalizeCharacterRelationshipState(
	state models.CharacterRelationshipState,
) models.CharacterRelationshipState {
	return models.CharacterRelationshipState{
		Trust:   normalizeCharacterRelationshipValue(state.Trust),
		Respect: normalizeCharacterRelationshipValue(state.Respect),
		Fear:    normalizeCharacterRelationshipValue(state.Fear),
		Debt:    normalizeCharacterRelationshipValue(state.Debt),
	}
}

func characterRelationshipStateMap(
	relationships []models.UserCharacterRelationship,
) map[uuid.UUID]models.CharacterRelationshipState {
	relationshipMap := make(map[uuid.UUID]models.CharacterRelationshipState, len(relationships))
	for _, relationship := range relationships {
		if relationship.CharacterID == uuid.Nil {
			continue
		}
		relationshipMap[relationship.CharacterID] = normalizeCharacterRelationshipState(relationship.State())
	}
	return relationshipMap
}

func (s *server) loadUserCharacterRelationshipMap(
	ctx context.Context,
	userID uuid.UUID,
	characterIDs []uuid.UUID,
) (map[uuid.UUID]models.CharacterRelationshipState, error) {
	relationships, err := s.dbClient.UserCharacterRelationship().FindByUserAndCharacterIDs(ctx, userID, characterIDs)
	if err != nil {
		return nil, err
	}
	return characterRelationshipStateMap(relationships), nil
}

func applyCharacterRelationship(
	character *models.Character,
	relationshipMap map[uuid.UUID]models.CharacterRelationshipState,
) {
	if character == nil {
		return
	}
	relationship, ok := relationshipMap[character.ID]
	if !ok || relationship.IsZero() {
		character.Relationship = nil
		return
	}
	normalized := normalizeCharacterRelationshipState(relationship)
	character.Relationship = &normalized
}

func collectCharacterIDsFromPointsOfInterest(
	pointsOfInterest []models.PointOfInterest,
) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	characterIDs := make([]uuid.UUID, 0)
	for _, pointOfInterest := range pointsOfInterest {
		for _, character := range pointOfInterest.Characters {
			if character.ID == uuid.Nil {
				continue
			}
			if _, exists := seen[character.ID]; exists {
				continue
			}
			seen[character.ID] = struct{}{}
			characterIDs = append(characterIDs, character.ID)
		}
	}
	return characterIDs
}

func (s *server) applyQuestGiverRelationshipEffectsOnTurnIn(
	ctx context.Context,
	userID uuid.UUID,
	quest *models.Quest,
) error {
	if quest == nil || quest.QuestGiverCharacterID == nil {
		return nil
	}
	return s.applyQuestGiverRelationshipDelta(
		ctx,
		userID,
		*quest.QuestGiverCharacterID,
		quest.QuestGiverRelationshipEffects,
	)
}

func (s *server) applyQuestGiverRelationshipDelta(
	ctx context.Context,
	userID uuid.UUID,
	characterID uuid.UUID,
	delta models.CharacterRelationshipState,
) error {
	if characterID == uuid.Nil {
		return nil
	}
	delta = normalizeCharacterRelationshipState(delta)
	if delta.IsZero() {
		return nil
	}
	_, err := s.dbClient.UserCharacterRelationship().ApplyDelta(
		ctx,
		userID,
		characterID,
		delta,
	)
	return err
}
