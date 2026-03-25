package server

import (
	"context"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func normalizeQuestDifficultySettings(
	rawMode string,
	input *int,
	fallbackMode models.QuestDifficultyMode,
	fallbackDifficulty int,
) (models.QuestDifficultyMode, int, error) {
	mode := fallbackMode
	if strings.TrimSpace(rawMode) != "" {
		mode = models.NormalizeQuestDifficultyMode(rawMode)
	}
	difficulty := models.NormalizeQuestDifficulty(fallbackDifficulty)
	if input != nil {
		if *input < 1 {
			return "", 0, errInvalidQuestDifficulty()
		}
		difficulty = *input
	}
	return models.NormalizeQuestDifficultyMode(string(mode)), models.NormalizeQuestDifficulty(difficulty), nil
}

func errInvalidQuestDifficulty() error {
	return &questDifficultyValidationError{message: "difficulty must be one or greater"}
}

type questDifficultyValidationError struct {
	message string
}

func (e *questDifficultyValidationError) Error() string {
	return e.message
}

func questUsesScaledDifficulty(mode models.QuestDifficultyMode) bool {
	return models.NormalizeQuestDifficultyMode(string(mode)) == models.QuestDifficultyModeScale
}

func syncQuestDifficultyConfiguration(
	ctx context.Context,
	s *server,
	quest *models.Quest,
) error {
	if s == nil || quest == nil {
		return nil
	}

	scaleWithUserLevel := questUsesScaledDifficulty(quest.DifficultyMode)
	difficulty := models.NormalizeQuestDifficulty(quest.Difficulty)

	nodes, err := s.dbClient.QuestNode().FindByQuestID(ctx, quest.ID)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		for _, challenge := range node.Challenges {
			updatedChallenge := challenge
			updatedChallenge.ScaleWithUserLevel = scaleWithUserLevel
			updatedChallenge.Difficulty = difficulty
			updatedChallenge.UpdatedAt = time.Now()
			if _, err := s.dbClient.QuestNodeChallenge().Update(ctx, challenge.ID, &updatedChallenge); err != nil {
				return err
			}
		}

		if node.ChallengeID != nil && *node.ChallengeID != uuid.Nil {
			challenge, err := s.dbClient.Challenge().FindByID(ctx, *node.ChallengeID)
			if err != nil {
				return err
			}
			if challenge != nil {
				challenge.ScaleWithUserLevel = scaleWithUserLevel
				challenge.Difficulty = difficulty
				challenge.UpdatedAt = time.Now()
				if err := s.dbClient.Challenge().Update(ctx, challenge.ID, challenge); err != nil {
					return err
				}
			}
		}

		if node.ScenarioID != nil && *node.ScenarioID != uuid.Nil {
			scenario, err := s.dbClient.Scenario().FindByID(ctx, *node.ScenarioID)
			if err != nil {
				return err
			}
			if scenario != nil {
				scenario.ScaleWithUserLevel = scaleWithUserLevel
				scenario.Difficulty = difficulty
				scenario.UpdatedAt = time.Now()
				if err := s.dbClient.Scenario().Update(ctx, scenario.ID, scenario); err != nil {
					return err
				}
				if err := s.dbClient.Scenario().ReplaceOptions(
					ctx,
					scenario.ID,
					scenarioOptionsWithQuestDifficulty(scenario.Options, scaleWithUserLevel, difficulty),
				); err != nil {
					return err
				}
			}
		}

		if node.MonsterEncounterID != nil && *node.MonsterEncounterID != uuid.Nil {
			encounter, err := s.dbClient.MonsterEncounter().FindByID(ctx, *node.MonsterEncounterID)
			if err != nil {
				return err
			}
			if encounter != nil {
				encounter.ScaleWithUserLevel = scaleWithUserLevel
				encounter.UpdatedAt = time.Now()
				if err := s.dbClient.MonsterEncounter().Update(ctx, encounter.ID, encounter); err != nil {
					return err
				}
				if !scaleWithUserLevel {
					for _, member := range encounter.Members {
						monster := member.Monster
						monster.Level = difficulty
						monster.UpdatedAt = time.Now()
						if err := s.dbClient.Monster().Update(ctx, monster.ID, &monster); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func scenarioOptionsWithQuestDifficulty(
	input []models.ScenarioOption,
	scaleWithUserLevel bool,
	difficulty int,
) []models.ScenarioOption {
	options := make([]models.ScenarioOption, 0, len(input))
	for _, option := range input {
		cloned := option
		if scaleWithUserLevel {
			cloned.Difficulty = nil
		} else {
			value := models.NormalizeQuestDifficulty(difficulty)
			cloned.Difficulty = &value
		}
		options = append(options, cloned)
	}
	return options
}
