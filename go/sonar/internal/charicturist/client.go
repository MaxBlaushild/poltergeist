package charicturist

import (
	"context"
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/useapi"
	"github.com/google/uuid"
)

type Client interface {
	CreateCharacter(ctx context.Context, request CreateCharacterRequest) error
}

type client struct {
	useApi   useapi.Client
	dbClient db.DbClient
}

type CreateCharacterRequest struct {
	UserId            uuid.UUID
	ProfilePictureUrl string
	Gender            string
}

const (
	imaginePrompt = " a pirate, %sprofile picture, pixelated, retro video game style, white background"
)

func NewClient(useApi useapi.Client, dbClient db.DbClient) Client {
	return &client{useApi: useApi, dbClient: dbClient}
}

func (c *client) CreateCharacter(ctx context.Context, request CreateCharacterRequest) error {
	genderQualifier := "gender is " + request.Gender + ", "
	prompt := fmt.Sprintf(imaginePrompt, genderQualifier)
	imagineResponse, err := c.useApi.GenerateImageOptions(ctx, request.ProfilePictureUrl+prompt)
	if err != nil {
		return err
	}

	if err := c.dbClient.ImageGeneration().Create(ctx, &models.ImageGeneration{
		ID:                  uuid.New(),
		UserID:              request.UserId,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		GenerationID:        imagineResponse.Hash,
		GenerationBackendID: models.GenerationBackendUseApi,
		Status:              models.GenerationStatusPending,
	}); err != nil {
		return err
	}

	return nil
}
