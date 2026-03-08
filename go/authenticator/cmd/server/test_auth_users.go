package main

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	demoPhone       = "+14407858475"
	secondTestPhone = "+12025550101"
	testAuthCode    = "123456"
)

type testAuthUser struct {
	ID          uuid.UUID
	Name        string
	PhoneNumber string
	Code        string
}

var testAuthUsersByPhone = map[string]testAuthUser{
	demoPhone: {
		ID:          uuid.MustParse("d8d28ec1-2162-4d87-97d6-c8f8b6e6a801"),
		Name:        "Demo User",
		PhoneNumber: demoPhone,
		Code:        testAuthCode,
	},
	secondTestPhone: {
		ID:          uuid.MustParse("be105033-b2e6-4d4d-a91c-70037dbcc0f3"),
		Name:        "Test User 2",
		PhoneNumber: secondTestPhone,
		Code:        testAuthCode,
	},
}

func lookupTestAuthUser(phoneNumber string) (*testAuthUser, bool) {
	user, ok := testAuthUsersByPhone[formatPhoneNumber(phoneNumber)]
	if !ok {
		return nil, false
	}

	return &user, true
}

func lookupTestAuthUserForLogin(phoneNumber string, code string) (*testAuthUser, bool) {
	user, ok := lookupTestAuthUser(phoneNumber)
	if !ok || user.Code != code {
		return nil, false
	}

	return user, true
}

func ensureTestAuthUser(ctx context.Context, dbClient db.DbClient, user *testAuthUser) (*models.User, error) {
	foundUser, err := dbClient.User().FindByPhoneNumber(ctx, user.PhoneNumber)
	if err == nil && foundUser != nil {
		return foundUser, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return dbClient.User().Insert(
		ctx,
		user.Name,
		user.PhoneNumber,
		&user.ID,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}
