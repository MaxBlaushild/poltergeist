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
	NeighboringPointsOfInterest() NeighboringPointsOfInterestHandle
	TextVerificationCode() TextVerificationCodeHandle
	SentText() SentTextHandle
	HowManySubscription() HowManySubscriptionHandle
	SonarSurvey() SonarSurveyHandle
	SonarSurveySubmission() SonarSurveySubmissionHandle
	SonarActivity() SonarActivityHandle
	SonarCategory() SonarCategoryHandle
	Match() MatchHandle
	VerificationCode() VerificationCodeHandle
	PointOfInterestGroup() PointOfInterestGroupHandle
	PointOfInterestChallenge() PointOfInterestChallengeHandle
	InventoryItem() InventoryItemHandle
	AuditItem() AuditItemHandle
	ImageGeneration() ImageGenerationHandle
	PointOfInterestChildren() PointOfInterestChildrenHandle
	PointOfInterestDiscovery() PointOfInterestDiscoveryHandle
	MatchUser() MatchUserHandle
	Tag() TagHandle
	TagGroup() TagGroupHandle
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
	UpdateProfilePictureUrl(ctx context.Context, userID uuid.UUID, url string) error
}

type TeamHandle interface {
	GetAll(ctx context.Context) ([]models.Team, error)
	Create(ctx context.Context, userIDs []uuid.UUID, teamName string, matchID uuid.UUID) (*models.Team, error)
	AddUserToTeam(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
	RemoveUserFromMatch(ctx context.Context, matchID uuid.UUID, userID uuid.UUID) error
	UpdateTeamName(ctx context.Context, teamID uuid.UUID, name string) (*models.Team, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Team, error)
	GetByMatchID(ctx context.Context, matchID uuid.UUID) ([]models.Team, error)
}

type UserTeamHandle interface{}

type PointOfInterestHandle interface {
	FindAll(ctx context.Context) ([]models.PointOfInterest, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterest, error)
	Create(ctx context.Context, crystal models.PointOfInterest) error
	Unlock(ctx context.Context, pointOfInterestID uuid.UUID, teamID *uuid.UUID, userID *uuid.UUID) error
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.PointOfInterest, error)
	Edit(ctx context.Context, id uuid.UUID, name string, description string, lat string, lng string) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateImageUrl(ctx context.Context, id uuid.UUID, imageUrl string) error
	CreateForGroup(ctx context.Context, pointOfInterest *models.PointOfInterest, pointOfInterestGroupID uuid.UUID) error
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

type MatchHandle interface {
	Create(ctx context.Context, creatorID uuid.UUID, pointsOfInterestIDs []uuid.UUID) (*models.Match, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Match, error)
	StartMatch(ctx context.Context, matchID uuid.UUID) error
	EndMatch(ctx context.Context, matchID uuid.UUID) error
	FindCurrentMatchForUser(ctx context.Context, userId uuid.UUID) (*models.Match, error)
	FindForTeamID(ctx context.Context, teamID uuid.UUID) (*models.TeamMatch, error)
	FindCurrentMatchIDForUser(ctx context.Context, userId uuid.UUID) (*uuid.UUID, error)
}

type VerificationCodeHandle interface {
	Create(ctx context.Context) (*models.VerificationCode, error)
}

type PointOfInterestGroupHandle interface {
	Create(ctx context.Context, name string, description string, imageUrl string, typeValue models.PointOfInterestGroupType) (*models.PointOfInterestGroup, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestGroup, error)
	FindAll(ctx context.Context) ([]*models.PointOfInterestGroup, error)
	Edit(ctx context.Context, id uuid.UUID, name string, description string, typeValue models.PointOfInterestGroupType) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateImageUrl(ctx context.Context, id uuid.UUID, imageUrl string) error
	FindByType(ctx context.Context, typeValue models.PointOfInterestGroupType) ([]*models.PointOfInterestGroup, error)
	GetNearbyQuests(ctx context.Context, userID uuid.UUID, lat float64, lng float64, radiusInMeters float64, tags []string) ([]models.PointOfInterestGroup, error)
	GetStartedQuests(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestGroup, error)
}

type PointOfInterestChallengeHandle interface {
	Create(ctx context.Context, pointOfInterestID uuid.UUID, tier int, question string, inventoryItemID int) (*models.PointOfInterestChallenge, error)
	SubmitAnswerForChallenge(ctx context.Context, challengeID uuid.UUID, teamID *uuid.UUID, userID *uuid.UUID, text string, imageURL string, isCorrect bool) (*models.PointOfInterestChallengeSubmission, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestChallenge, error)
	GetChallengeForPointOfInterest(ctx context.Context, pointOfInterestID uuid.UUID, tier int) (*models.PointOfInterestChallenge, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Edit(ctx context.Context, id uuid.UUID, question string, inventoryItemID int, tier int) (*models.PointOfInterestChallenge, error)
	GetSubmissionsForMatch(ctx context.Context, matchID uuid.UUID) ([]models.PointOfInterestChallengeSubmission, error)
	GetSubmissionsForUser(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestChallengeSubmission, error)
}

type InventoryItemHandle interface {
	CreateOrIncrementInventoryItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, inventoryItemID int, quantity int) error
	UseInventoryItem(ctx context.Context, ownedInventoryItemID uuid.UUID) error
	ApplyInventoryItem(ctx context.Context, matchID uuid.UUID, inventoryItemID int, teamID uuid.UUID, duration time.Duration) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.OwnedInventoryItem, error)
	StealItems(ctx context.Context, thiefTeamID uuid.UUID, victimTeamID uuid.UUID) error
	GetItems(ctx context.Context, userOrTeam models.OwnedInventoryItem) ([]models.OwnedInventoryItem, error)
	StealItem(ctx context.Context, thiefTeamID uuid.UUID, victimTeamID uuid.UUID, inventoryItemID int) error
}

type AuditItemHandle interface {
	Create(ctx context.Context, matchID *uuid.UUID, userID *uuid.UUID, message string) error
	GetAuditItemsForMatch(ctx context.Context, matchID uuid.UUID) ([]*models.AuditItem, error)
	GetAuditItemsForUser(ctx context.Context, userID uuid.UUID) ([]*models.AuditItem, error)
}

type ImageGenerationHandle interface {
	Create(ctx context.Context, imageGeneration *models.ImageGeneration) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.ImageGeneration, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.ImageGeneration, error)
	UpdateState(ctx context.Context, imageGenerationID uuid.UUID, state models.GenerationStatus) error
	FindByState(ctx context.Context, state models.GenerationStatus) ([]models.ImageGeneration, error)
	SetOptions(ctx context.Context, imageGenerationID uuid.UUID, options []string) error
	Updates(ctx context.Context, imageGenerationID uuid.UUID, updates *models.ImageGeneration) error
	GetCompleteGenerationsForUser(ctx context.Context, userID uuid.UUID) ([]models.ImageGeneration, error)
}

type PointOfInterestChildrenHandle interface {
	Create(ctx context.Context, pointOfInterestGroupMemberID uuid.UUID, pointOfInterestID uuid.UUID, pointOfInterestChallengeID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PointOfInterestDiscoveryHandle interface {
	GetDiscoveriesForTeam(teamID uuid.UUID) ([]models.PointOfInterestDiscovery, error)
	GetDiscoveriesForUser(userID uuid.UUID) ([]models.PointOfInterestDiscovery, error)
}

type MatchUserHandle interface {
	Create(ctx context.Context, matchUser *models.MatchUser) error
	FindByMatchID(ctx context.Context, matchID uuid.UUID) ([]models.MatchUser, error)
	FindUsersForMatch(ctx context.Context, matchID uuid.UUID) ([]models.User, error)
}

type TagHandle interface {
	FindAll(ctx context.Context) ([]*models.Tag, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Tag, error)
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*models.Tag, error)
	Create(ctx context.Context, tag *models.Tag) error
	Update(ctx context.Context, tag *models.Tag) error
	AddTagToPointOfInterest(ctx context.Context, tagID uuid.UUID, pointOfInterestID uuid.UUID) error
	RemoveTagFromPointOfInterest(ctx context.Context, tagID uuid.UUID, pointOfInterestID uuid.UUID) error
}

type TagGroupHandle interface {
	FindAll(ctx context.Context) ([]*models.TagGroup, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.TagGroup, error)
	Create(ctx context.Context, tagGroup *models.TagGroup) error
	Update(ctx context.Context, tagGroup *models.TagGroup) error
}
