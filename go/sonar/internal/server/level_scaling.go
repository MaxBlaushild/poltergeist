package server

import (
	"context"
	"math"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const (
	scenarioScaledDifficultyBuffer  = 10
	challengeScaledDifficultyBuffer = 20
)

func normalizeScaledLevel(level int) int {
	if level < 1 {
		return 1
	}
	return level
}

func expectedSpecializedStatForLevel(level int) int {
	normalizedLevel := normalizeScaledLevel(level)
	pointsGained := (normalizedLevel - 1) * models.CharacterStatPointsPerLevel
	return models.CharacterStatBaseValue + pointsGained
}

func scaledScenarioDifficultyForUserLevel(level int) int {
	return maxInt(0, expectedSpecializedStatForLevel(level)+scenarioScaledDifficultyBuffer)
}

func scaledChallengeDifficultyForUserLevel(level int) int {
	return maxInt(0, expectedSpecializedStatForLevel(level)+challengeScaledDifficultyBuffer)
}

func (s *server) currentUserLevel(ctx context.Context, userID uuid.UUID) (int, error) {
	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	return normalizeScaledLevel(userLevel.Level), nil
}

func (s *server) currentPartyMaxLevel(
	ctx context.Context,
	user *models.User,
) (int, error) {
	if user == nil {
		return 1, nil
	}
	maxLevel, err := s.currentUserLevel(ctx, user.ID)
	if err != nil {
		return 0, err
	}
	if user.PartyID == nil || *user.PartyID == uuid.Nil {
		return maxLevel, nil
	}

	partyMembers, err := s.dbClient.User().FindPartyMembers(ctx, user.ID)
	if err != nil {
		return 0, err
	}
	for _, member := range partyMembers {
		memberLevel, err := s.currentUserLevel(ctx, member.ID)
		if err != nil {
			return 0, err
		}
		if memberLevel > maxLevel {
			maxLevel = memberLevel
		}
	}
	return maxLevel, nil
}

func scenarioDifficultyForUserLevel(scenario *models.Scenario, userLevel int) int {
	if scenario == nil {
		return 0
	}
	if !scenario.ScaleWithUserLevel {
		return scenario.Difficulty
	}
	return scaledScenarioDifficultyForUserLevel(userLevel)
}

func challengeDifficultyForUserLevel(challenge *models.Challenge, userLevel int) int {
	if challenge == nil {
		return 0
	}
	if !challenge.ScaleWithUserLevel {
		return challenge.Difficulty
	}
	return scaledChallengeDifficultyForUserLevel(userLevel)
}

func scenarioWithScaledDifficulty(scenario models.Scenario, userLevel int) models.Scenario {
	scenario.Difficulty = scenarioDifficultyForUserLevel(&scenario, userLevel)
	return scenario
}

func challengeWithScaledDifficulty(challenge models.Challenge, userLevel int) models.Challenge {
	challenge.Difficulty = challengeDifficultyForUserLevel(&challenge, userLevel)
	return challenge
}

func scaledEncounterMonsterLevelForUserLevel(level int, memberCount int) int {
	normalizedLevel := normalizeScaledLevel(level)
	scaleFactor := 0.90
	switch {
	case memberCount <= 1:
		scaleFactor = 0.90
	case memberCount == 2:
		scaleFactor = 0.50
	default:
		scaleFactor = 0.35
	}
	scaled := int(math.Round(float64(normalizedLevel) * scaleFactor))
	return maxInt(1, scaled)
}
