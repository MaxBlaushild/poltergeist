package server

import (
	"context"
	"math"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const (
	scenarioScaledStatGainPerLevel     = 2
	challengeScaledDifficultyBuffer    = 20
	challengeScaledStatGainPerLevel    = 3
	challengeScaledNoStatDifficultyCap = 50
	standardEncounterScaleFactor       = 0.90
	raidEncounterRecommendedPartySize  = 5.0
	maxMonsterAbilityLeadLevels        = 1
	maxBossEncounterBonusLevels        = 5
)

func normalizeScaledLevel(level int) int {
	if level < 1 {
		return 1
	}
	return level
}

func expectedScenarioStatForLevel(level int) int {
	normalizedLevel := normalizeScaledLevel(level)
	pointsGained := (normalizedLevel - 1) * scenarioScaledStatGainPerLevel
	return models.CharacterStatBaseValue + pointsGained
}

func scaledScenarioDifficultyForUserLevel(level int) int {
	return maxInt(0, expectedScenarioStatForLevel(level))
}

func expectedFocusedChallengeStatForLevel(level int) int {
	normalizedLevel := normalizeScaledLevel(level)
	pointsGained := (normalizedLevel - 1) * challengeScaledStatGainPerLevel
	return models.CharacterStatBaseValue + pointsGained
}

func scaledChallengeDifficultyForUserLevel(level int) int {
	return maxInt(
		0,
		expectedFocusedChallengeStatForLevel(level)+challengeScaledDifficultyBuffer,
	)
}

func scaledChallengeDifficultyForStatTags(level int, statTags []string) int {
	difficulty := scaledChallengeDifficultyForUserLevel(level)
	hasStatTags := false
	for _, tag := range statTags {
		if tag != "" {
			hasStatTags = true
			break
		}
	}
	if !hasStatTags && difficulty > challengeScaledNoStatDifficultyCap {
		return challengeScaledNoStatDifficultyCap
	}
	return difficulty
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
	return scaledChallengeDifficultyForStatTags(userLevel, []string(challenge.StatTags))
}

func scenarioWithScaledDifficulty(scenario models.Scenario, userLevel int) models.Scenario {
	scenario.Difficulty = scenarioDifficultyForUserLevel(&scenario, userLevel)
	return scenario
}

func challengeWithScaledDifficulty(challenge models.Challenge, userLevel int) models.Challenge {
	challenge.Difficulty = challengeDifficultyForUserLevel(&challenge, userLevel)
	return challenge
}

func scaledStandardEncounterMonsterLevelForUserLevel(level int, memberCount int) int {
	normalizedLevel := normalizeScaledLevel(level)
	scaleFactor := standardEncounterScaleFactor
	switch {
	case memberCount <= 1:
		scaleFactor = standardEncounterScaleFactor
	case memberCount == 2:
		scaleFactor = 0.50
	default:
		scaleFactor = 0.35
	}
	scaled := int(math.Round(float64(normalizedLevel) * scaleFactor))
	return maxInt(1, scaled)
}

func scaledRaidEncounterMonsterLevelForUserLevel(level int, memberCount int) int {
	normalizedLevel := normalizeScaledLevel(level)
	normalizedMemberCount := maxInt(1, memberCount)
	scaleFactor := (standardEncounterScaleFactor * raidEncounterRecommendedPartySize) / float64(normalizedMemberCount)
	scaled := int(math.Round(float64(normalizedLevel) * scaleFactor))
	return maxInt(1, scaled)
}

func scaledBossEncounterBonusLevels(level int) int {
	normalizedLevel := normalizeScaledLevel(level)
	bonusLevels := normalizedLevel / 3
	if bonusLevels > maxBossEncounterBonusLevels {
		return maxBossEncounterBonusLevels
	}
	return maxInt(0, bonusLevels)
}

func cappedMonsterAbilityLevelForUserLevel(monsterLevel int, userLevel int) int {
	normalizedMonsterLevel := normalizeScaledLevel(monsterLevel)
	if userLevel <= 0 {
		return normalizedMonsterLevel
	}
	normalizedUserLevel := normalizeScaledLevel(userLevel)
	cappedLevel := normalizedUserLevel + maxMonsterAbilityLeadLevels
	if cappedLevel < normalizedMonsterLevel {
		return cappedLevel
	}
	return normalizedMonsterLevel
}

func scaledEncounterMonsterLevelForUserLevelAndType(level int, memberCount int, encounterType models.MonsterEncounterType) int {
	switch models.NormalizeMonsterEncounterType(string(encounterType)) {
	case models.MonsterEncounterTypeBoss:
		return scaledStandardEncounterMonsterLevelForUserLevel(
			level+scaledBossEncounterBonusLevels(level),
			memberCount,
		)
	case models.MonsterEncounterTypeRaid:
		return scaledRaidEncounterMonsterLevelForUserLevel(level, memberCount)
	default:
		return scaledStandardEncounterMonsterLevelForUserLevel(level, memberCount)
	}
}

func scaledEncounterMonsterLevelForUserLevel(level int, memberCount int) int {
	return scaledEncounterMonsterLevelForUserLevelAndType(level, memberCount, models.MonsterEncounterTypeMonster)
}
