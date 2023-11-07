package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
)

type client struct {
	db                             *gorm.DB
	scoreHandle                    *scoreHandler
	userHandle                     *userHandle
	questionSetHandle              *questionSetHandle
	matchHandle                    *matchHandle
	userSubmissionHandle           *userSubmissionHandle
	questionHandle                 *questionHandle
	howManyQuestionHandle          *howManyQuestionHandle
	howManyAnswerHandle            *howManyAnswerHandle
	challengeHandle                *challengeHandle
	credentialHandle               *credentialHandle
	teamHandle                     *teamHandle
	userTeamHandle                 *userTeamHandle
	crystalHandle                  *crystalHandle
	crystalUnlockingHandle         *crystalUnlockingHandle
	neighborHandle                 *neighborHandle
	textVerificationCodeHandle     *textVerificationCodeHandle
	sentTextHandle                 *sentTextHandle
	guessHowManuSubscriptionHandle *guessHowManySubscriptionHandle
}

type ClientConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func NewClient(cfg ClientConfig) (DbClient, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return &client{
		db:                             db,
		scoreHandle:                    &scoreHandler{db: db},
		userHandle:                     &userHandle{db: db},
		questionSetHandle:              &questionSetHandle{db: db},
		matchHandle:                    &matchHandle{db: db},
		userSubmissionHandle:           &userSubmissionHandle{db: db},
		questionHandle:                 &questionHandle{db: db},
		howManyQuestionHandle:          &howManyQuestionHandle{db: db},
		howManyAnswerHandle:            &howManyAnswerHandle{db: db},
		challengeHandle:                &challengeHandle{db: db},
		credentialHandle:               &credentialHandle{db: db},
		teamHandle:                     &teamHandle{db: db},
		userTeamHandle:                 &userTeamHandle{db: db},
		crystalHandle:                  &crystalHandle{db: db},
		crystalUnlockingHandle:         &crystalUnlockingHandle{db: db},
		neighborHandle:                 &neighborHandle{db: db},
		textVerificationCodeHandle:     &textVerificationCodeHandle{db: db},
		sentTextHandle:                 &sentTextHandle{db: db},
		guessHowManuSubscriptionHandle: &guessHowManySubscriptionHandle{db: db},
	}, err
}

func (c *client) Score() ScoreHandle {
	return c.scoreHandle
}

func (c *client) Exec(ctx context.Context, q string) error {
	return c.db.WithContext(ctx).Exec(q).Error
}

func (c *client) Migrate(ctx context.Context, m ...interface{}) error {
	return c.db.WithContext(ctx).AutoMigrate(m...)
}

func (c *client) QuestionSet() QuestionSetHandle {
	return c.questionSetHandle
}

func (c *client) Match() MatchHandle {
	return c.matchHandle
}

func (c *client) UserSubmission() UserSubmissionHandle {
	return c.userSubmissionHandle
}

func (c *client) Question() QuestionHandle {
	return c.questionHandle
}

func (c *client) HowManyAnswer() HowManyAnswerHandle {
	return c.howManyAnswerHandle
}

func (c *client) HowManyQuestion() HowManyQuestionHandle {
	return c.howManyQuestionHandle
}

func (c *client) User() UserHandle {
	return c.userHandle
}
func (c *client) Challenge() ChallengeHandle {
	return c.challengeHandle
}

func (c *client) Credential() CredentialHandle {
	return c.credentialHandle
}

func (c *client) Crystal() CrystalHandle {
	return c.crystalHandle
}

func (c *client) Neighbor() NeighborHandle {
	return c.neighborHandle
}

func (c *client) Team() TeamHandle {
	return c.teamHandle
}

func (c *client) UserTeam() UserTeamHandle {
	return c.userTeamHandle
}

func (c *client) CrystalUnlocking() CrystalUnlockingHandle {
	return c.crystalUnlockingHandle
}

func (c *client) TextVerificationCode() TextVerificationCodeHandle {
	return c.textVerificationCodeHandle
}

func (c *client) SentText() SentTextHandle {
	return c.sentTextHandle
}

func (c *client) GuessHowManySubscription() GuessHowManySubscriptionHandle {
	return c.guessHowManuSubscriptionHandle
}