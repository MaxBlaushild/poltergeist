package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/crystal-crisis-api/internal/models"
)

type DbClient interface {
	Team() TeamHandle
	UserTeam() UserTeamHandle
	Crystal() CrystalHandle
	CrystalUnlocking() CrystalUnlockingHandle
	Neighbor() NeighborHandle
	Migrate(ctx context.Context, models ...interface{}) error
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
