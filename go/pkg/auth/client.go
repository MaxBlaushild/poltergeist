package auth

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type GetUsersRequest struct {
	UserIDs []uint `json:"userIds" binding:"required"`
}

type RegisterByTextRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name"`
}

type LoginByTextRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

type User struct {
	gorm.Model
	Name        string `json:"name"`
	PhoneNumber string `json:"phoneNumber"`
}

type authClient struct{}

type AuthClient interface {
	GetUsers(userIDs []uint) ([]User, error)
	RegisterByText(request *RegisterByTextRequest) (*models.User, error)
	LoginByText(request *LoginByTextRequest) (*models.User, error)
}

const (
	baseUrl = "http://localhost:8089"
)

func NewAuthClient() AuthClient {
	return &authClient{}
}

func (d *authClient) GetUsers(userIDs []uint) ([]User, error) {
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

