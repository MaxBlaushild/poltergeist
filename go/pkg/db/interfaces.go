package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type DbClient interface {
	Score() ScoreHandle
	User() UserHandle
	HowManyQuestion() HowManyQuestionHandle
	HowManyAnswer() HowManyAnswerHandle
	Team() TeamHandle
	UserTeam() UserTeamHandle
	PointOfInterest() PointOfInterestHandle
	PointOfInterestTeam() PointOfInterestTeamHandle
	NeighboringPointsOfInterest() NeighboringPointsOfInterestHandle
	TextVerificationCode() TextVerificationCodeHandle
	SentText() SentTextHandle
	HowManySubscription() HowManySubscriptionHandle
	SonarSurvey() SonarSurveyHandle
	SonarSurveySubmission() SonarSurveySubmissionHandle
	SonarActivity() SonarActivityHandle
	SonarCategory() SonarCategoryHandle
	SonarUser() SonarUserHandle
	Match() MatchHandle
	VerificationCode() VerificationCodeHandle
	PointOfInterestGroup() PointOfInterestGroupHandle
	PointOfInterestChallenge() PointOfInterestChallengeHandle
	InventoryItem() InventoryItemHandle
	Exec(ctx context.Context, q string) error
}

type ScoreHandle interface {
	Upsert(ctx context.Context, username string) (*models.Score, error)
	FindAll(ctx context.Context) ([]models.Score, error)
}

type HowManyAnswerHandle interface {
	FindByQuestionIDAndUserID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.HowManyAnswer, error)
	Insert(ctx context.Context, a *models.HowManyAnswer) (*models.HowManyAnswer, error)
}

type HowManyQuestionHandle interface {
	Insert(ctx context.Context, text string, explanation string, howMany int, promptSeedIndex int, prompt string) (*models.HowManyQuestion, error)
	FindAll(ctx context.Context) ([]*models.HowManyQuestion, error)
	MarkValid(ctx context.Context, howManyQuestionID uuid.UUID) error
	MarkDone(ctx context.Context, howManyQuestionID uuid.UUID) error
	FindTodaysQuestion(ctx context.Context) (*models.HowManyQuestion, error)
	FindById(ctx context.Context, id uuid.UUID) (*models.HowManyQuestion, error)
	ValidQuestionsRemaining(ctx context.Context) (int64, error)
}

type UserHandle interface {
	Insert(ctx context.Context, name string, phoneNumber string, id *uuid.UUID) (*models.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error)
	FindUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]models.User, error)
	FindAll(ctx context.Context) ([]models.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	DeleteAll(ctx context.Context) error
}

type TeamHandle interface {
	GetAll(ctx context.Context) ([]models.Team, error)
	Create(ctx context.Context, userIDs []uuid.UUID, teamName string, matchID uuid.UUID) (*models.Team, error)
	AddUserToTeam(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
	RemoveUserFromMatch(ctx context.Context, matchID uuid.UUID, userID uuid.UUID) error
	UpdateTeamName(ctx context.Context, teamID uuid.UUID, name string) (*models.Team, error)
}

type UserTeamHandle interface{}

type PointOfInterestHandle interface {
	FindAll(ctx context.Context) ([]models.PointOfInterest, error)
	Capture(ctx context.Context, pointOfInterestID uuid.UUID, teamID uuid.UUID, tier int) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterest, error)
	Create(ctx context.Context, crystal models.PointOfInterest) error
	Unlock(ctx context.Context, crystalID uuid.UUID, teamID uuid.UUID) error
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.PointOfInterest, error)
}

type PointOfInterestTeamHandle interface {
	FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.PointOfInterestTeam, error)
}

type NeighboringPointsOfInterestHandle interface {
	Create(ctx context.Context, crystalOneID uuid.UUID, crystalTwoID uuid.UUID) error
	FindAll(ctx context.Context) ([]models.NeighboringPointsOfInterest, error)
}

type TextVerificationCodeHandle interface {
	Insert(ctx context.Context, phoneNumber string) (*models.TextVerificationCode, error)
	Find(ctx context.Context, phoneNumber string, code string) (*models.TextVerificationCode, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}

type SentTextHandle interface {
	GetCount(ctx context.Context, phoneNumber string, textType string) (int64, error)
	Insert(ctx context.Context, textType string, phoneNumber string, text string) (*models.SentText, error)
}

type HowManySubscriptionHandle interface {
	Insert(ctx context.Context, userID uuid.UUID) (*models.HowManySubscription, error)
	FindAll(ctx context.Context) ([]models.HowManySubscription, error)
	IncrementNumFreeQuestions(ctx context.Context, userID uuid.UUID) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.HowManySubscription, error)
	SetSubscribed(ctx context.Context, userID uuid.UUID, stripeID string) error
	DeleteByStripeID(ctx context.Context, stripeID string) error
}

type SonarSurveyHandle interface {
	GetSurveys(ctx context.Context, userID uuid.UUID) ([]models.SonarSurvey, error)
	CreateSurvey(ctx context.Context, userID uuid.UUID, title string, activityIDs []uuid.UUID) (*models.SonarSurvey, error)
	GetSurveyByID(ctx context.Context, surveyID uuid.UUID) (*models.SonarSurvey, error)
}

type SonarSurveySubmissionHandle interface {
	CreateSubmission(ctx context.Context, surveryID uuid.UUID, userID uuid.UUID, activityIDS []uuid.UUID, downs []bool) (*models.SonarSurveySubmission, error)
	GetUserSubmissionForSurvey(ctx context.Context, userID uuid.UUID, surveyID uuid.UUID) (*models.SonarSurveySubmission, error)
	GetAllSubmissionsForUser(ctx context.Context, userID uuid.UUID) ([]models.SonarSurveySubmission, error)
	GetSubmissionByID(ctx context.Context, submissionID uuid.UUID) (*models.SonarSurveySubmission, error)
}

type SonarActivityHandle interface {
	GetAllActivities(ctx context.Context) ([]models.SonarActivity, error)
	CreateActivity(ctx context.Context, activity models.SonarActivity) (models.SonarActivity, error)
	UpdateActivity(ctx context.Context, activity models.SonarActivity) (models.SonarActivity, error)
	DeleteActivity(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetActivityByID(ctx context.Context, id uuid.UUID) (models.SonarActivity, error)
}

type SonarCategoryHandle interface {
	GetCategoriesByUserID(ctx context.Context, userID uuid.UUID) ([]models.SonarCategory, error)
	GetAllCategoriesWithActivities(ctx context.Context) ([]models.SonarCategory, error)
	CreateCategory(ctx context.Context, category models.SonarCategory) (models.SonarCategory, error)
	UpdateCategory(ctx context.Context, category models.SonarCategory) (models.SonarCategory, error)
	DeleteCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetCategoryByID(ctx context.Context, id uuid.UUID) (models.SonarCategory, error)
}

type SonarUserHandle interface {
	FindOrCreateSonarUser(ctx context.Context, viewerID uuid.UUID, vieweeID uuid.UUID) (*models.SonarUser, error)
	GetSonarUserCount(ctx context.Context, viewerID uuid.UUID) (int64, error)
	FindUserByViewerAndViewee(ctx context.Context, viewerID uuid.UUID, vieweeID uuid.UUID) (*models.SonarUser, error)
	GetSonarUserProfileIcon(ctx context.Context, viewerID uuid.UUID) (string, error)
}

type MatchHandle interface {
	Create(ctx context.Context, creatorID uuid.UUID, pointsOfInterestIDs []uuid.UUID) (*models.Match, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Match, error)
	StartMatch(ctx context.Context, matchID uuid.UUID) error
	EndMatch(ctx context.Context, matchID uuid.UUID) error
	FindCurrentMatchForUser(ctx context.Context, userId uuid.UUID) (*models.Match, error)
	FindForTeamID(ctx context.Context, teamID uuid.UUID) (*models.TeamMatch, error)
}

type VerificationCodeHandle interface {
	Create(ctx context.Context) (*models.VerificationCode, error)
}

type PointOfInterestGroupHandle interface {
	Create(ctx context.Context, pointOfInterestIDs []uuid.UUID, name string) (*models.PointOfInterestGroup, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestGroup, error)
	FindAll(ctx context.Context) ([]*models.PointOfInterestGroup, error)
}

type PointOfInterestChallengeHandle interface {
	SubmitAnswerForChallenge(ctx context.Context, challengeID uuid.UUID, teamID uuid.UUID, text string, imageURL string, isCorrect bool) (*models.PointOfInterestChallengeSubmission, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestChallenge, error)
	GetChallengeForPointOfInterest(ctx context.Context, pointOfInterestID uuid.UUID, tier int) (*models.PointOfInterestChallenge, error)
}

type InventoryItemHandle interface {
	CreateOrIncrementInventoryItem(ctx context.Context, teamID uuid.UUID, inventoryItemID int, quantity int) error
	GetInventoryItem(ctx context.Context, teamID uuid.UUID, inventoryItemID int) (*models.TeamInventoryItem, error)
	UseInventoryItem(ctx context.Context, teamInventoryItemID uuid.UUID) error
	ApplyInventoryItem(ctx context.Context, matchID uuid.UUID, inventoryItemID int, teamID uuid.UUID, duration time.Duration) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.TeamInventoryItem, error)
	StealItems(ctx context.Context, thiefTeamID uuid.UUID, victimTeamID uuid.UUID) error
	GetTeamsItems(ctx context.Context, teamID uuid.UUID) ([]models.TeamInventoryItem, error)
}
