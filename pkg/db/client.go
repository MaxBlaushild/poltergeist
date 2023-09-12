package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
)

type client struct {
	db                    *gorm.DB
	scoreHandle           *scoreHandler
	userhandle            *userHandler
	questionSetHandle     *questionSetHandle
	matchHandle           *matchHandle
	userSubmissionHandle  *userSubmissionHandle
	questionHandle        *questionHandle
	howManyQuestionHandle *howManyQuestionHandle
	howManyAnswerHandle   *howManyAnswerHandle
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
		db:                    db,
		scoreHandle:           &scoreHandler{db: db},
		userhandle:            &userHandler{db: db},
		questionSetHandle:     &questionSetHandle{db: db},
		matchHandle:           &matchHandle{db: db},
		userSubmissionHandle:  &userSubmissionHandle{db: db},
		questionHandle:        &questionHandle{db: db},
		howManyQuestionHandle: &howManyQuestionHandle{db: db},
		howManyAnswerHandle:   &howManyAnswerHandle{db: db},
	}, err
}

func (c *client) Score() ScoreHandle {
	return c.scoreHandle
}

func (c *client) Migrate(ctx context.Context, m ...interface{}) error {
	return c.db.WithContext(ctx).AutoMigrate(m...)
}

func (c *client) User() UserHandle {
	return c.userhandle
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
