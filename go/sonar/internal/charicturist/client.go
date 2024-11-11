package charicturist

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/imagine"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type Client interface {
	CreateCharacter(ctx context.Context, request CreateCharacterRequest) error
}

type client struct {
	imagine  imagine.ImagineClient
	dbClient db.DbClient
}

type CreateCharacterRequest struct {
	UserId            uuid.UUID
	ProfilePictureUrl string
}

const (
	imaginePrompt = " a pirate, profile picture, pixelated, retro video game style, white background"
)

func NewClient(imagine imagine.ImagineClient, dbClient db.DbClient) Client {
	return &client{imagine: imagine, dbClient: dbClient}
}

func (c *client) CreateCharacter(ctx context.Context, request CreateCharacterRequest) error {
	imagineResponse, err := c.imagine.InitiateImageGeneration(ctx, request.ProfilePictureUrl+imaginePrompt)
	if err != nil {
		return err
	}

	if err := c.dbClient.ImageGeneration().Create(ctx, &models.ImageGeneration{
		ID:                  uuid.New(),
		UserID:              request.UserId,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		GenerationID:        imagineResponse.Data.ID,
		GenerationBackendID: models.GenerationBackendImagine,
		Status:              models.GenerationStatusPending,
	}); err != nil {
		return err
	}

	return nil
}
