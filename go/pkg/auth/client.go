package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/http"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type GetUsersRequest struct {
	UserIDs []uuid.UUID `json:"userIds" binding:"required"`
}

type RegisterByTextRequest struct {
	PhoneNumber string  `json:"phoneNumber" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name"`
	UserID      *string `json:"userId"`
}

type LoginByTextRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

type User struct {
	ID          uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phoneNumber"`
}

type client struct {
	httpClient http.Client
}

type Client interface {
	GetUsers(ctx context.Context, userIDs []uuid.UUID) ([]User, error)
	RegisterByText(ctx context.Context, request *RegisterByTextRequest) (*models.User, error)
	LoginByText(ctx context.Context, request *LoginByTextRequest) (*models.User, error)
}

const (
	baseUrl = "http://localhost:8089"
)

func NewClient() Client {
	httpClient := http.NewClient(baseUrl, http.ApplicationJson)
	return &client{httpClient: httpClient}
}

func (c *client) GetUsers(ctx context.Context, userIDs []uuid.UUID) ([]User, error) {
	request := GetUsersRequest{
		UserIDs: userIDs,
	}

	respBytes, err := c.httpClient.Post(ctx, "/authenticator/get-users", &request)
	if err != nil {
		return nil, err
	}

	var users []User
	err = json.Unmarshal(respBytes, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (c *client) RegisterByText(ctx context.Context, request *RegisterByTextRequest) (*models.User, error) {
	respBytes, err := c.httpClient.Post(ctx, "/authenticator/text/register", request)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = json.Unmarshal(respBytes, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *client) LoginByText(ctx context.Context, request *LoginByTextRequest) (*models.User, error) {
	respBytes, err := c.httpClient.Post(ctx, "/authenticator/text/login", request)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = json.Unmarshal(respBytes, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
