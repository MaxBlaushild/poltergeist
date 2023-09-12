package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
)

type client struct {
	db               *gorm.DB
	scoreHandle      *userHandle
	challengeHandle  *challengeHandle
	credentialHandle *credentialHandle
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
		db:               db,
		scoreHandle:      &userHandle{db: db},
		challengeHandle:  &challengeHandle{db: db},
		credentialHandle: &credentialHandle{db: db},
	}, err
}

func (c *client) User() UserHandle {
	return c.scoreHandle
}
func (c *client) Challenge() ChallengeHandle {
	return c.challengeHandle
}

func (c *client) Migrate(ctx context.Context, m ...interface{}) error {
	return c.db.WithContext(ctx).AutoMigrate(m...)
}

func (c *client) Credential() CredentialHandle {
	return c.credentialHandle
}
