package auth

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"gorm.io/gorm"
)

type GetUsersRequest struct {
	UserIDs []uint `json:"userIds" binding:"required"`
}

type User struct {
	gorm.Model
	Name        string `json:"name"`
	PhoneNumber string `json:"phoneNumber"`
}

type authClient struct{}

type AuthClient interface {
	GetUsers(userIDs []uint) ([]User, error)
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
