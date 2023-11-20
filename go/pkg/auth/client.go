package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

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

type authClient struct{}

type AuthClient interface {
	GetUsers(userIDs []uuid.UUID) ([]User, error)
	RegisterByText(request *RegisterByTextRequest) (*models.User, error)
	LoginByText(request *LoginByTextRequest) (*models.User, error)
}

const (
	baseUrl = "http://localhost:8089"
)

func NewAuthClient() AuthClient {
	return &authClient{}
}

func (d *authClient) GetUsers(userIDs []uuid.UUID) ([]User, error) {
	request := GetUsersRequest{
		UserIDs: userIDs,
	}
	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseUrl+"/authenticator/get-users", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, errors.New("error making request to authenticator")
	}

	var users []User
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (d *authClient) RegisterByText(request *RegisterByTextRequest) (*models.User, error) {
	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseUrl+"/authenticator/text/register", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, errors.New("error making request to authenticator")
	}

	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *authClient) LoginByText(request *LoginByTextRequest) (*models.User, error) {
	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseUrl+"/authenticator/text/login", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, errors.New("error making request to authenticator")
	}

	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
