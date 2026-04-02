package processors

import "github.com/MaxBlaushild/poltergeist/pkg/models"

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
