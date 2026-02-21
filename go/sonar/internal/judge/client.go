package judge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const (
	JudgementMessageTemplate = `
	You are a judge on a gameshow. You are tasked with deciding whether 
	or not a team has successfully completed a challenge. Please be a 
	bit lenient towards, but still make sure the basic premise of the of the challenge
	is fulfilled. A challenge is not considered completed if the picture is not real.

	Here is the challenge: %s

	%s

	%s

	Please answer in the form of a JSON object with the following fields:
	
		{
			"judgement": true | false,
			"reason": "string"
		}
	`
	ScoreJudgementMessageTemplate = `
	You are a judge on a gameshow. You are tasked with scoring how well
	a team has completed a challenge. Be a bit lenient, but still make sure
	the basic premise of the challenge is fulfilled. A challenge should score
	0 if the picture is not real or the submission is unrelated.

	Here is the challenge: %s

	%s

	%s

	Please answer in the form of a JSON object with the following fields:

		{
			"score": 0-50,
			"reason": "string"
		}
	`
)

type Client interface {
	JudgeSubmission(ctx context.Context, request JudgeSubmissionRequest) (*JudgeSubmissionResponse, error)
	JudgeFreeform(ctx context.Context, request FreeformJudgeSubmissionRequest) (*FreeformJudgeSubmissionResponse, error)
}

type client struct {
	aws        aws.AWSClient
	db         db.DbClient
	deepPriest deep_priest.DeepPriest
}

type JudgeSubmissionRequest struct {
	Challenge          *models.PointOfInterestChallenge
	ImageSubmissionUrl string
	TextSubmission     string
	TeamID             *uuid.UUID
	UserID             *uuid.UUID
}

type SubmissionJudgement struct {
	Judgement bool   `json:"judgement"`
	Reason    string `json:"reason"`
}

type ScoreJudgement struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type JudgeSubmissionResponse struct {
	Challenge *models.PointOfInterestChallengeSubmission `json:"challenge"`
	Judgement SubmissionJudgement                        `json:"judgement"`
}

func (r *JudgeSubmissionResponse) IsSuccessful() bool {
	return r.Judgement.Judgement
}

type FreeformJudgeSubmissionRequest struct {
	Question           string
	ImageSubmissionUrl string
	TextSubmission     string
}

type FreeformJudgeSubmissionResponse struct {
	Judgement ScoreJudgement `json:"judgement"`
}

func NewClient(aws aws.AWSClient, db db.DbClient, deepPriest deep_priest.DeepPriest) Client {
	return &client{
		aws:        aws,
		db:         db,
		deepPriest: deepPriest,
	}
}

// getPartyMembers returns all party members if the user is in a party, otherwise returns just the user
func (c *client) getPartyMembers(ctx context.Context, userID *uuid.UUID) ([]models.User, error) {
	if userID == nil {
		return []models.User{}, nil
	}

	user, err := c.db.User().FindByID(ctx, *userID)
	if err != nil {
		return nil, err
	}

	// If user is in a party, return all party members
	if user.PartyID != nil {
		return c.db.User().FindPartyMembers(ctx, *userID)
	}

	// If not in a party, return just this user
	return []models.User{*user}, nil
}

func (c *client) JudgeSubmission(ctx context.Context, request JudgeSubmissionRequest) (*JudgeSubmissionResponse, error) {
	prompt := c.makeJudgementMessage(request.Challenge, request)

	var answer *deep_priest.Answer
	var err error
	if request.ImageSubmissionUrl != "" {
		answer, err = c.deepPriest.PetitionTheFountWithImage(&deep_priest.QuestionWithImage{
			Question: prompt,
			Image:    request.ImageSubmissionUrl,
		})
	} else {
		answer, err = c.deepPriest.PetitionTheFount(&deep_priest.Question{
			Question: prompt,
		})
	}
	if err != nil {
		return nil, err
	}

	judgementResult := SubmissionJudgement{}
	err = json.Unmarshal([]byte(answer.Answer), &judgementResult)
	if err != nil {
		return nil, fmt.Errorf("error decoding judgement response (%s): %w", answer.Answer, err)
	}

	// Get all party members (or just the user if not in a party)
	partyMembers, err := c.getPartyMembers(ctx, request.UserID)
	if err != nil {
		return nil, err
	}

	// Create submission records for all party members
	var challengeSubmission *models.PointOfInterestChallengeSubmission
	for _, member := range partyMembers {
		submission, err := c.db.PointOfInterestChallenge().SubmitAnswerForChallenge(ctx, request.Challenge.ID, request.TeamID, &member.ID, request.TextSubmission, request.ImageSubmissionUrl, judgementResult.Judgement)
		if err != nil {
			return nil, err
		}

		// Keep the original submitter's submission as the response
		if request.UserID != nil && member.ID == *request.UserID {
			challengeSubmission = submission
		}
	}

	// If no party members found, create a submission for the original user
	if challengeSubmission == nil && request.UserID != nil {
		challengeSubmission, err = c.db.PointOfInterestChallenge().SubmitAnswerForChallenge(ctx, request.Challenge.ID, request.TeamID, request.UserID, request.TextSubmission, request.ImageSubmissionUrl, judgementResult.Judgement)
		if err != nil {
			return nil, err
		}
	}

	return &JudgeSubmissionResponse{
		Challenge: challengeSubmission,
		Judgement: judgementResult,
	}, nil
}

func (c *client) JudgeFreeform(ctx context.Context, request FreeformJudgeSubmissionRequest) (*FreeformJudgeSubmissionResponse, error) {
	prompt := c.makeScoreJudgementMessageForQuestion(request.Question, request.TextSubmission, request.ImageSubmissionUrl)

	var answer *deep_priest.Answer
	var err error
	if request.ImageSubmissionUrl != "" {
		answer, err = c.deepPriest.PetitionTheFountWithImage(&deep_priest.QuestionWithImage{
			Question: prompt,
			Image:    request.ImageSubmissionUrl,
		})
	} else {
		answer, err = c.deepPriest.PetitionTheFount(&deep_priest.Question{
			Question: prompt,
		})
	}
	if err != nil {
		return nil, err
	}

	judgementResult := ScoreJudgement{}
	if err := json.Unmarshal([]byte(answer.Answer), &judgementResult); err != nil {
		return nil, fmt.Errorf("error decoding judgement response (%s): %w", answer.Answer, err)
	}

	return &FreeformJudgeSubmissionResponse{
		Judgement: judgementResult,
	}, nil
}

func (c *client) makeJudgementMessage(challenge *models.PointOfInterestChallenge, request JudgeSubmissionRequest) string {
	textMessage := ""
	imageMessage := ""

	if request.TextSubmission != "" {
		textMessage = fmt.Sprintf("Here is the text part of the submission: '%s'", request.TextSubmission)
	}

	if request.ImageSubmissionUrl != "" {
		imageMessage = "You should also look at the image included as part of the submission."
	}

	return fmt.Sprintf(JudgementMessageTemplate, challenge.Question, textMessage, imageMessage)
}

func (c *client) makeJudgementMessageForQuestion(question string, textSubmission string, imageSubmissionUrl string) string {
	textMessage := ""
	imageMessage := ""

	if textSubmission != "" {
		textMessage = fmt.Sprintf("Here is the text part of the submission: '%s'", textSubmission)
	}
	if imageSubmissionUrl != "" {
		imageMessage = "You should also look at the image included as part of the submission."
	}
	return fmt.Sprintf(JudgementMessageTemplate, question, textMessage, imageMessage)
}

func (c *client) makeScoreJudgementMessageForQuestion(question string, textSubmission string, imageSubmissionUrl string) string {
	textMessage := ""
	imageMessage := ""

	if textSubmission != "" {
		textMessage = fmt.Sprintf("Here is the text part of the submission: '%s'", textSubmission)
	}
	if imageSubmissionUrl != "" {
		imageMessage = "You should also look at the image included as part of the submission."
	}
	return fmt.Sprintf(ScoreJudgementMessageTemplate, question, textMessage, imageMessage)
}
