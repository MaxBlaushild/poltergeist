package db

import (
	"context"

	"github.com/MaxBlaushild/authenticator/internal/models"
)

type DbClient interface {
	Migrate(ctx context.Context, models ...interface{}) error
	User() UserHandle
	Challenge() ChallengeHandle
	Credential() CredentialHandle
}

type UserHandle interface {
	Insert(ctx context.Context, name string, phoneNumber string) (*models.AuthUser, error)
	FindByID(ctx context.Context, id uint) (*models.AuthUser, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*models.AuthUser, error)
	FindUsersByIDs(ctx context.Context, userIDs []uint) ([]models.AuthUser, error)
	FindAll(ctx context.Context) ([]models.AuthUser, error)
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
