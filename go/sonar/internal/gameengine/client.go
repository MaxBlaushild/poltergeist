package gameengine

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/chat"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/judge"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/quartermaster"
	"github.com/google/uuid"
)

const (
	BaseReputationPointsAwardedForSuccessfulSubmission = 100
	BaseExperiencePointsAwardedForSuccessfulSubmission = 100
	BaseExperiencePointsAwardedForFinishedQuest        = 250
	BaseReputationPointsAwardedForFinishedQuest        = 250
)

type Submission struct {
	ChallengeID uuid.UUID
	TeamID      *uuid.UUID
	UserID      *uuid.UUID
	ImageURL    string
	Text        string
}

type SubmissionResult struct {
	Successful        bool                          `json:"successful"`
	Reason            string                        `json:"reason"`
	QuestCompleted    bool                          `json:"questCompleted"`
	ItemsAwarded      []quartermaster.InventoryItem `json:"itemsAwarded"`
	ExperienceAwarded int                           `json:"experienceAwarded"`
	ReputationAwarded int                           `json:"reputationAwarded"`
	ZoneID            uuid.UUID                     `json:"zoneID"`
}

type GameEngineClient interface {
	ProcessSubmission(ctx context.Context, submission Submission) (*SubmissionResult, error)
}

type gameEngineClient struct {
	db            db.DbClient
	judge         judge.Client
	quartermaster quartermaster.Quartermaster
	chatClient    chat.Client
}

func NewGameEngineClient(
	db db.DbClient,
	judge judge.Client,
	quartermaster quartermaster.Quartermaster,
	chatClient chat.Client,
) GameEngineClient {
	return &gameEngineClient{db: db, judge: judge, quartermaster: quartermaster, chatClient: chatClient}
}

func (c *gameEngineClient) ProcessSubmission(ctx context.Context, submission Submission) (*SubmissionResult, error) {
	challenge, err := c.db.PointOfInterestChallenge().FindByID(ctx, submission.ChallengeID)
	if err != nil {
		return nil, err
	}

	judgementResult, err := c.judgeSubmission(ctx, submission, challenge)
	if err != nil {
		return nil, err
	}

	if !judgementResult.IsSuccessful() {
		return &SubmissionResult{
			Successful: false,
			Reason:     judgementResult.Judgement.Reason,
		}, nil
	}

	return c.processSuccessfulSubmission(ctx, submission, challenge)
}

func (c *gameEngineClient) judgeSubmission(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge) (*judge.JudgeSubmissionResponse, error) {
	judgementResult, err := c.judge.JudgeSubmission(ctx, judge.JudgeSubmissionRequest{
		Challenge:          challenge,
		TeamID:             submission.TeamID,
		UserID:             submission.UserID,
		ImageSubmissionUrl: submission.ImageURL,
		TextSubmission:     submission.Text,
	})
	if err != nil {
		return nil, err
	}

	return judgementResult, nil
}

func (c *gameEngineClient) processSuccessfulSubmission(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge) (*SubmissionResult, error) {
	questCompleted, err := c.HasCompletedQuest(ctx, challenge)
	if err != nil {
		return nil, err
	}

	submissionResult := SubmissionResult{
		QuestCompleted: questCompleted,
		Successful:     true,
	}

	if err = c.awardItems(ctx, submission, challenge, &submissionResult); err != nil {
		return nil, err
	}

	if err := c.addTaskCompleteMessage(ctx, submission, challenge, &submissionResult); err != nil {
		return nil, err
	}

	return &submissionResult, nil
}

func (c *gameEngineClient) awardItems(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge, submissionResult *SubmissionResult) error {
	if challenge.InventoryItemID == 0 {
		item, err := c.quartermaster.GetItem(ctx, submission.TeamID, submission.UserID)
		if err != nil {
			return err
		}

		submissionResult.ItemsAwarded = append(submissionResult.ItemsAwarded, item)
	}

	item, err := c.quartermaster.GetItemSpecificItem(ctx, submission.TeamID, submission.UserID, challenge.InventoryItemID)
	if err != nil {
		return err
	}

	submissionResult.ItemsAwarded = append(submissionResult.ItemsAwarded, item)

	return nil
}

func (c *gameEngineClient) HasCompletedQuest(ctx context.Context, challenge *models.PointOfInterestChallenge) (bool, error) {
	children, err := c.db.PointOfInterestChallenge().GetChildrenForChallenge(ctx, challenge.ID)
	if err != nil {
		return false, err
	}
	return len(children) == 0, nil
}

func (c *gameEngineClient) addTaskCompleteMessage(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge, submissionResult *SubmissionResult) error {
	if err := c.chatClient.AddCaptureMessage(ctx, submission.TeamID, submission.UserID, challenge); err != nil {
		return err
	}

	if submissionResult.QuestCompleted {
		return c.chatClient.AddCompletedQuestMessage(ctx, submission.TeamID, submission.UserID, challenge)
	}

	return nil
}

func (c *gameEngineClient) awardExperiencePoints(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge, submissionResult *SubmissionResult) error {
	experiencePoints := BaseExperiencePointsAwardedForSuccessfulSubmission
	if submissionResult.QuestCompleted {
		experiencePoints += BaseExperiencePointsAwardedForFinishedQuest
	}

	submissionResult.ExperienceAwarded = experiencePoints

	if submission.UserID != nil {
		if _, err := c.db.UserLevel().ProcessExperiencePointAdditions(ctx, *submission.UserID, experiencePoints); err != nil {
			return err
		}
	}

	return nil
}

func (c *gameEngineClient) awardReputationPoints(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge, submissionResult *SubmissionResult) error {
	reputationPoints := BaseReputationPointsAwardedForSuccessfulSubmission
	if submissionResult.QuestCompleted {
		reputationPoints += BaseReputationPointsAwardedForFinishedQuest
	}

	submissionResult.ReputationAwarded = reputationPoints

	zone, err := c.db.PointOfInterest().FindZoneForPointOfInterest(ctx, challenge.PointOfInterestID)
	if err != nil {
		return err
	}

	submissionResult.ZoneID = zone.ZoneID

	if submission.UserID != nil {
		if _, err := c.db.UserZoneReputation().ProcessReputationPointAdditions(ctx, *submission.UserID, zone.ZoneID, reputationPoints); err != nil {
			return err
		}
	}

	return nil
}
