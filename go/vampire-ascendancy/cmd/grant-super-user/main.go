// Command grant-super-user grants a user edit access to the shared content
// library (characters, houses, items, quiz questions) — the bootstrap step
// for the very first super user, since the /admin dashboard's own "grant"
// action requires already being a super user. Every super user after the
// first can be granted from that dashboard instead.
//
//	go run ./cmd/grant-super-user --config-name local --email you@example.com
package main

import (
	"context"
	"flag"
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/vampire-ascendancy/internal/config"
)

func main() {
	email := flag.String("email", "", "Look up the user by email.")
	phone := flag.String("phone", "", "Look up the user by phone number, e.g. +15555550123.")

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if (*email == "") == (*phone == "") {
		log.Fatalf("pass exactly one of --email or --phone")
	}

	dbClient, err := db.NewClient(db.ClientConfig{
		Name:     cfg.Public.DbName,
		Host:     cfg.Public.DbHost,
		Port:     cfg.Public.DbPort,
		User:     cfg.Public.DbUser,
		Password: cfg.Secret.DbPassword,
	})
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	ctx := context.Background()

	var user *models.User
	if *email != "" {
		user, err = dbClient.User().FindByEmail(ctx, *email)
	} else {
		user, err = dbClient.User().FindByPhoneNumber(ctx, *phone)
	}
	if err != nil {
		log.Fatalf("failed to look up user: %v", err)
	}
	if user == nil {
		log.Fatalf("no user found — they need an account (sign up first) before they can be granted access")
	}

	if err := dbClient.Vampire().AddSuperUser(ctx, user.ID, nil); err != nil {
		log.Fatalf("failed to grant super user: %v", err)
	}

	log.Printf("%s (%s) can now edit the shared content library", user.Name, user.ID)
}
