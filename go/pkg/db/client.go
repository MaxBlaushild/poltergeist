package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
)

type client struct {
	db                                *gorm.DB
	scoreHandle                       *scoreHandler
	userHandle                        *userHandle
	howManyQuestionHandle             *howManyQuestionHandle
	howManyAnswerHandle               *howManyAnswerHandle
	teamHandle                        *teamHandle
	userTeamHandle                    *userTeamHandle
	pointOfInterestHandle             *pointOfInterestHandle
	pointOfInterestTeamHandle         *pointOfInterestTeamHandle
	neighboringPointsOfInterestHandle *neighboringPointsOfInterestHandle
	textVerificationCodeHandle        *textVerificationCodeHandle
	sentTextHandle                    *sentTextHandle
	howManySubscriptionHandle         *howManySubscriptionHandle
	sonarSurveyHandle                 *sonarSurveyHandle
	sonarSurveySubmissionHandle       *sonarSurveySubmissionHandle
	sonarActivityHandle               *sonarActivityHandle
	sonarCategoryHandle               *sonarCategoryHandle
	sonarUserHandle                   *sonarUserHandle
	matchHandle                       *matchHandle
	verificationCodeHandle            *verificationCodeHandler
	pointOfInterestGroupHandle        *pointOfInterestGroupHandle
	pointOfInterestChallengeHandle    *pointOfInterestChallengeHandle
	inventoryItemHandle               *inventoryItemHandler
	auditItemHandle                   *auditItemHandler
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
		db:                                db,
		scoreHandle:                       &scoreHandler{db: db},
		userHandle:                        &userHandle{db: db},
		howManyQuestionHandle:             &howManyQuestionHandle{db: db},
		howManyAnswerHandle:               &howManyAnswerHandle{db: db},
		teamHandle:                        &teamHandle{db: db},
		userTeamHandle:                    &userTeamHandle{db: db},
		pointOfInterestHandle:             &pointOfInterestHandle{db: db},
		pointOfInterestTeamHandle:         &pointOfInterestTeamHandle{db: db},
		neighboringPointsOfInterestHandle: &neighboringPointsOfInterestHandle{db: db},
		textVerificationCodeHandle:        &textVerificationCodeHandle{db: db},
		sentTextHandle:                    &sentTextHandle{db: db},
		howManySubscriptionHandle:         &howManySubscriptionHandle{db: db},
		sonarSurveyHandle:                 &sonarSurveyHandle{db: db},
		sonarSurveySubmissionHandle:       &sonarSurveySubmissionHandle{db: db},
		sonarActivityHandle:               &sonarActivityHandle{db: db},
		sonarCategoryHandle:               &sonarCategoryHandle{db: db},
		sonarUserHandle:                   &sonarUserHandle{db: db},
		matchHandle:                       &matchHandle{db: db},
		verificationCodeHandle:            &verificationCodeHandler{db: db},
		pointOfInterestGroupHandle:        &pointOfInterestGroupHandle{db: db},
		pointOfInterestChallengeHandle:    &pointOfInterestChallengeHandle{db: db},
		inventoryItemHandle:               &inventoryItemHandler{db: db},
		auditItemHandle:                   &auditItemHandler{db: db},
	}, err
}

func (c *client) AuditItem() AuditItemHandle {
	return c.auditItemHandle
}

func (c *client) InventoryItem() InventoryItemHandle {
	return c.inventoryItemHandle
}

func (c *client) PointOfInterestChallenge() PointOfInterestChallengeHandle {
	return c.pointOfInterestChallengeHandle
}

func (c *client) PointOfInterestGroup() PointOfInterestGroupHandle {
	return c.pointOfInterestGroupHandle
}

func (c *client) VerificationCode() VerificationCodeHandle {
	return c.verificationCodeHandle
}

func (c *client) Match() MatchHandle {
	return c.matchHandle
}

func (c *client) Score() ScoreHandle {
	return c.scoreHandle
}

func (c *client) Exec(ctx context.Context, q string) error {
	return c.db.WithContext(ctx).Exec(q).Error
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

func (c *client) PointOfInterest() PointOfInterestHandle {
	return c.pointOfInterestHandle
}

func (c *client) NeighboringPointsOfInterest() NeighboringPointsOfInterestHandle {
	return c.neighboringPointsOfInterestHandle
}

func (c *client) Team() TeamHandle {
	return c.teamHandle
}

func (c *client) UserTeam() UserTeamHandle {
	return c.userTeamHandle
}

func (c *client) PointOfInterestTeam() PointOfInterestTeamHandle {
	return c.pointOfInterestTeamHandle
}

func (c *client) TextVerificationCode() TextVerificationCodeHandle {
	return c.textVerificationCodeHandle
}

func (c *client) SentText() SentTextHandle {
	return c.sentTextHandle
}

func (c *client) HowManySubscription() HowManySubscriptionHandle {
	return c.howManySubscriptionHandle
}

func (c *client) SonarSurvey() SonarSurveyHandle {
	return c.sonarSurveyHandle
}

func (c *client) SonarSurveySubmission() SonarSurveySubmissionHandle {
	return c.sonarSurveySubmissionHandle
}

func (c *client) SonarActivity() SonarActivityHandle {
	return c.sonarActivityHandle
}

func (c *client) SonarCategory() SonarCategoryHandle {
	return c.sonarCategoryHandle
}

func (c *client) SonarUser() SonarUserHandle {
	return c.sonarUserHandle
}
