package gameengine

import (
	"context"
	"encoding/json"

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
	Successful     bool   `json:"successful"`
	Reason         string `json:"reason"`
	QuestCompleted bool   `json:"questCompleted"`
}

type GameEngineClient interface {
	ProcessSuccessfulSubmission(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge) (*SubmissionResult, error)
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

// getPartyMembers returns all party members if the user is in a party, otherwise returns just the user
func (c *gameEngineClient) getPartyMembers(ctx context.Context, userID *uuid.UUID) ([]models.User, error) {
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

	return c.ProcessSuccessfulSubmission(ctx, submission, challenge)
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

func (c *gameEngineClient) ProcessSuccessfulSubmission(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge) (*SubmissionResult, error) {
	var partyID *uuid.UUID
	if submission.UserID != nil {
		user, err := c.db.User().FindByID(ctx, *submission.UserID)
		if err == nil && user.PartyID != nil {
			partyID = user.PartyID
		}
	}

	questCompleted, err := c.HasCompletedQuest(ctx, challenge)
	if err != nil {
		return nil, err
	}

	submissionResult := SubmissionResult{
		QuestCompleted: questCompleted,
		Successful:     true,
		Reason:         "Challenge completed successfully!",
	}

	// Create activity for challenge completed
	challengeActivityData, err := json.Marshal(models.ChallengeCompletedActivity{
		ChallengeID: challenge.ID,
		Successful:  true,
		Reason:      "Challenge completed successfully!",
		SubmitterID: submission.UserID,
	})
	if err != nil {
		return nil, err
	}
	if err := c.db.Activity().CreateActivitiesForPartyMembers(ctx, partyID, submission.UserID, models.ActivityTypeChallengeCompleted, challengeActivityData); err != nil {
		return nil, err
	}

	// Create activity for quest completed if applicable
	if questCompleted {
		// For now, use the challenge ID as a placeholder for quest ID
		// In a more complete implementation, we'd look up the actual quest ID
		questActivityData, err := json.Marshal(models.QuestCompletedActivity{
			QuestID: challenge.ID,
		})
		if err != nil {
			return nil, err
		}
		if err := c.db.Activity().CreateActivitiesForPartyMembers(ctx, partyID, submission.UserID, models.ActivityTypeQuestCompleted, questActivityData); err != nil {
			return nil, err
		}
	}

	if err = c.awardItems(ctx, submission, challenge); err != nil {
		return nil, err
	}

	if err = c.awardExperiencePoints(ctx, submission, questCompleted); err != nil {
		return nil, err
	}

	if err = c.awardReputationPoints(ctx, submission, challenge, questCompleted); err != nil {
		return nil, err
	}

	if err := c.addTaskCompleteMessage(ctx, submission, challenge, &submissionResult); err != nil {
		return nil, err
	}

	return &submissionResult, nil
}

func (c *gameEngineClient) awardItems(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge) error {
	// Get all party members or just the submitter
	partyMembers, err := c.getPartyMembers(ctx, submission.UserID)
	if err != nil {
		return err
	}

	// Award items to each party member
	for _, member := range partyMembers {
		memberID := member.ID

		if challenge.InventoryItemID == 0 {
			item, err := c.quartermaster.GetItem(ctx, submission.TeamID, &memberID)
			if err != nil {
				return err
			}

			// Create activity for item received for this specific member
			activityData, err := json.Marshal(models.ItemReceivedActivity{
				ItemID:   item.ID,
				ItemName: item.Name,
			})
			if err != nil {
				return err
			}
			if err := c.db.Activity().CreateActivity(ctx, models.Activity{
				UserID:       memberID,
				ActivityType: models.ActivityTypeItemReceived,
				Data:         activityData,
				Seen:         false,
			}); err != nil {
				return err
			}
		}

		item, err := c.quartermaster.GetItemSpecificItem(ctx, submission.TeamID, &memberID, challenge.InventoryItemID)
		if err != nil {
			return err
		}

		// Create activity for item received for this specific member
		activityData, err := json.Marshal(models.ItemReceivedActivity{
			ItemID:   item.ID,
			ItemName: item.Name,
		})
		if err != nil {
			return err
		}
		if err := c.db.Activity().CreateActivity(ctx, models.Activity{
			UserID:       memberID,
			ActivityType: models.ActivityTypeItemReceived,
			Data:         activityData,
			Seen:         false,
		}); err != nil {
			return err
		}
	}

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

func (c *gameEngineClient) awardExperiencePoints(ctx context.Context, submission Submission, questCompleted bool) error {
	// Get all party members or just the submitter
	partyMembers, err := c.getPartyMembers(ctx, submission.UserID)
	if err != nil {
		return err
	}

	experiencePoints := BaseExperiencePointsAwardedForSuccessfulSubmission
	if questCompleted {
		experiencePoints += BaseExperiencePointsAwardedForFinishedQuest
	}

	// Award experience points to each party member
	for _, member := range partyMembers {
		userLevel, err := c.db.UserLevel().ProcessExperiencePointAdditions(ctx, member.ID, experiencePoints)
		if err != nil {
			return err
		}

		// Only create level-up activity for this member if they actually leveled up
		if userLevel.LevelsGained > 0 {
			activityData, err := json.Marshal(models.LevelUpActivity{
				NewLevel: userLevel.Level,
			})
			if err != nil {
				return err
			}
			if err := c.db.Activity().CreateActivity(ctx, models.Activity{
				UserID:       member.ID,
				ActivityType: models.ActivityTypeLevelUp,
				Data:         activityData,
				Seen:         false,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *gameEngineClient) awardReputationPoints(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge, questCompleted bool) error {
	// Get all party members or just the submitter
	partyMembers, err := c.getPartyMembers(ctx, submission.UserID)
	if err != nil {
		return err
	}

	reputationPoints := BaseReputationPointsAwardedForSuccessfulSubmission
	if questCompleted {
		reputationPoints += BaseReputationPointsAwardedForFinishedQuest
	}

	zone, err := c.db.PointOfInterest().FindZoneForPointOfInterest(ctx, challenge.PointOfInterestID)
	if err != nil {
		return err
	}

	// Award reputation points to each party member
	for _, member := range partyMembers {
		userZoneReputation, err := c.db.UserZoneReputation().ProcessReputationPointAdditions(ctx, member.ID, zone.ZoneID, reputationPoints)
		if err != nil {
			return err
		}

		// Only create reputation-up activity for this member if they actually gained reputation levels
		if userZoneReputation.LevelsGained > 0 {
			// Get full zone details
			fullZone, err := c.db.Zone().FindByID(ctx, zone.ZoneID)
			if err != nil {
				return err
			}

			activityData, err := json.Marshal(models.ReputationUpActivity{
				NewLevel: userZoneReputation.Level,
				ZoneName: fullZone.Name,
				ZoneID:   zone.ZoneID,
			})
			if err != nil {
				return err
			}
			if err := c.db.Activity().CreateActivity(ctx, models.Activity{
				UserID:       member.ID,
				ActivityType: models.ActivityTypeReputationUp,
				Data:         activityData,
				Seen:         false,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}
