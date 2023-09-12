package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
)

type client struct {
	db                     *gorm.DB
	teamHandle             *teamHandle
	userTeamHandle         *userTeamHandle
	crystalHandle          *crystalHandle
	crystalUnlockingHandle *crystalUnlockingHandle
	neighborHandle         *neighborHandle
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
		db:                     db,
		teamHandle:             &teamHandle{db: db},
		userTeamHandle:         &userTeamHandle{db: db},
		crystalHandle:          &crystalHandle{db: db},
		crystalUnlockingHandle: &crystalUnlockingHandle{db: db},
		neighborHandle:         &neighborHandle{db: db},
	}, err
}

func (c *client) Team() TeamHandle {
	return c.teamHandle
}

func (c *client) UserTeam() UserTeamHandle {
	return c.userTeamHandle
}

func (c *client) Crystal() CrystalHandle {
	return c.crystalHandle
}

func (c *client) CrystalUnlocking() CrystalUnlockingHandle {
	return c.crystalUnlockingHandle
}

func (c *client) Neighbor() NeighborHandle {
	return c.neighborHandle
}

func (c *client) Migrate(ctx context.Context, m ...interface{}) error {
	return c.db.WithContext(ctx).AutoMigrate(m...)
}
