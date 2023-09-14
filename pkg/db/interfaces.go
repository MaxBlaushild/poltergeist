package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

type DbClient interface {
	Migrate(ctx context.Context, models ...interface{}) error
	Score() ScoreHandle
	User() UserHandle
	QuestionSet() QuestionSetHandle
	Match() MatchHandle
	UserSubmission() UserSubmissionHandle
	Question() QuestionHandle
	HowManyQuestion() HowManyQuestionHandle
	HowManyAnswer() HowManyAnswerHandle
	Challenge() ChallengeHandle
	Credential() CredentialHandle
	Team() TeamHandle
	UserTeam() UserTeamHandle
	Crystal() CrystalHandle
	CrystalUnlocking() CrystalUnlockingHandle
	Neighbor() NeighborHandle
	TextVerificationCode() TextVerificationCodeHandle
}

type ScoreHandle interface {
	Upsert(ctx context.Context, username string) (*models.Score, error)
	FindAll(ctx context.Context) ([]models.Score, error)
}

type QuestionHandle interface {
	FindByQuestionSetID(ctx context.Context, questionSetID uint) ([]models.Question, error)
	GetAllQuestions(ctx context.Context) ([]models.Question, error)
}

type QuestionSetHandle interface {
	Insert(ctx context.Context, questions []models.Question) (*models.QuestionSet, error)
}

type MatchHandle interface {
	Insert(ctx context.Context, match *models.Match) error
	GetCurrentMatchForUser(ctx context.Context, userID uint) (*models.Match, error)
}

type UserSubmissionHandle interface {
	Insert(ctx context.Context, questionSetID uint, userID uint, userAnswers []models.UserAnswer) (*models.UserSubmission, error)
	FindByUserAndQuestionSetID(ctx context.Context, userID uint, questionSetID uint) (*models.UserSubmission, error)
}

type HowManyAnswerHandle interface {
	FindByQuestionIDAndUserID(ctx context.Context, id uint, userID uint) (*models.HowManyAnswer, error)
	FindByQuestionIDAndEphemeralUserID(ctx context.Context, id uint, userID string) (*models.HowManyAnswer, error)
	Insert(ctx context.Context, a *models.HowManyAnswer) (*models.HowManyAnswer, error)
}

type HowManyQuestionHandle interface {
	Insert(ctx context.Context, text string, explanation string, howMany int) (*models.HowManyQuestion, error)
	FindAll(ctx context.Context) ([]*models.HowManyQuestion, error)
	MarkValid(ctx context.Context, howManyQuestionID string) error
	MarkDone(ctx context.Context, howManyQuestionID uint) error
	FindTodaysQuestion(ctx context.Context) (*models.HowManyQuestion, error)
	FindById(ctx context.Context, id uint) (*models.HowManyQuestion, error)
}

type UserHandle interface {
	Insert(ctx context.Context, name string, phoneNumber string) (*models.User, error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error)
	FindUsersByIDs(ctx context.Context, userIDs []uint) ([]models.User, error)
	FindAll(ctx context.Context) ([]models.User, error)
	Delete(ctx context.Context, userID uint) error
	DeleteAll(ctx context.Context) error
}

type ChallengeHandle interface {
	Insert(ctx context.Context, challenge string, userID uint) error
	Find(ctx context.Context, challenge string) (*models.Challenge, error)
}

type CredentialHandle interface {
	Insert(ctx context.Context, credentialID string, publicKey string, userID uint) (*models.Credential, error)
	FindAll(ctx context.Context) ([]models.Credential, error)
	Delete(ctx context.Context, credentialID uint) error
	DeleteAll(ctx context.Context) error
}

type TeamHandle interface {
	GetAll(ctx context.Context) ([]models.Team, error)
	Create(ctx context.Context, userIDs []uint, teamName string) error
}

type UserTeamHandle interface{}

type CrystalHandle interface {
	FindAll(ctx context.Context) ([]models.Crystal, error)
	Capture(ctx context.Context, crystalID uint, teamID uint, attune bool) error
	FindByID(ctx context.Context, id uint) (*models.Crystal, error)
	Create(ctx context.Context, crystal models.Crystal) error
	Unlock(ctx context.Context, crystalID uint, teamID uint) error
}

type CrystalUnlockingHandle interface {
	FindByTeamID(ctx context.Context, teamID string) ([]models.CrystalUnlocking, error)
}

type NeighborHandle interface {
	Create(ctx context.Context, crystalOneID uint, crystalTwoID uint) error
	FindAll(ctx context.Context) ([]models.Neighbor, error)
}

type TextVerificationCodeHandle interface {
	Insert(ctx context.Context, phoneNumber string) (*models.TextVerificationCode, error)
	Find(ctx context.Context, phoneNumber string, code string) (*models.TextVerificationCode, error)
	MarkUsed(ctx context.Context, id uint) error
}
