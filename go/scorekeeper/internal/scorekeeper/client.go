package scorekeeper

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
)

const (
	scorekeeperCharacterPrompt = "You are a worker at a click farm. "
	updateScorePrompt          = "You have just manually updated the score for a user in a game, and you are reporting the new score to the firm that hired you. You are grateful for your job. "
	findScoresPrompt           = "You have just been asked for an update on all of the current scores for all of the users in the game from your employers. The scores are as follows: "
	cheatingPrompt             = "You have just found out that one of your employers has tried to cheat at the game they have hired you to proctor. You are to politely chide them for cheating."

	praiseEmployersPrompt = "You should include praise for your employers. "
	metaphorPrompt        = "You should include a metaphor in your response. "

	messagePrompt              = "Please write a slack message to send to your employers. "
	slackMessageTruncatePrompt = "Please include only the message contents in your response. "
	sentenceTerminatorPrompt   = ". "
)

type scorekeeper struct {
	deepPriest deep_priest.DeepPriest
	dbClient   db.DbClient
}

type Scorekeeper interface {
	UpdateScore(ctx context.Context, username string) (string, error)
	GetScores(ctx context.Context) (string, error)
	ChideCheating(ctx context.Context) (string, error)
}

func NewScorekeeper(dbClient db.DbClient) Scorekeeper {
	deepPriest := deep_priest.SummonDeepPriest()

	return &scorekeeper{
		dbClient:   dbClient,
		deepPriest: deepPriest,
	}
}

func flavorPrompt(prompts []string) string {
	rand.Seed(time.Now().UnixNano())

	return prompts[rand.Intn(len(prompts))]
}

func (s *scorekeeper) UpdateScore(ctx context.Context, username string) (string, error) {
	score, err := s.dbClient.Score().Upsert(ctx, username)
	if err != nil {
		return "", err
	}

	prompt := scorekeeperCharacterPrompt
	prompt += updateScorePrompt
	prompt += fmt.Sprintf("You have just recorded the new score of %d for username %s. Make sure to mention this username explicitly in your response. ", score.Score, score.Username)
	prompt += messagePrompt
	prompt += flavorPrompt([]string{praiseEmployersPrompt, metaphorPrompt})
	prompt += slackMessageTruncatePrompt

	answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: prompt,
	})
	if err != nil {
		return "", err
	}

	return answer.Answer, nil
}

func (s *scorekeeper) GetScores(ctx context.Context) (string, error) {
	scores, err := s.dbClient.Score().FindAll(ctx)
	if err != nil {
		return "", err
	}

	prompt := scorekeeperCharacterPrompt
	prompt += findScoresPrompt

	for _, score := range scores {
		prompt += fmt.Sprintf("(username: %s, score: %d)", score.Username, score.Score)
	}

	prompt += sentenceTerminatorPrompt
	prompt += messagePrompt
	prompt += praiseEmployersPrompt
	prompt += slackMessageTruncatePrompt

	answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: prompt,
	})
	if err != nil {
		return "", err
	}

	return answer.Answer, nil
}

func (s *scorekeeper) ChideCheating(ctx context.Context) (string, error) {
	prompt := scorekeeperCharacterPrompt
	prompt += cheatingPrompt
	prompt += messagePrompt
	prompt += flavorPrompt([]string{praiseEmployersPrompt, metaphorPrompt})
	prompt += slackMessageTruncatePrompt

	answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: prompt,
	})
	if err != nil {
		return "", err
	}

	return answer.Answer, nil
}
