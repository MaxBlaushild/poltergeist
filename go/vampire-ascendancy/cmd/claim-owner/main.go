// Command claim-owner grants a user the Host (owner) role on an instance —
// the post-migration bootstrap step called out in
// go/vampire-ascendancy/docs/MULTI_TENANT_REQUIREMENTS.md: the legacy
// instance created by migration 000455 has no admin at all (its
// created_by is NULL, since no single "creator" exists for an event that
// predates this concept), so someone needs a one-time ops action to become
// its first Host before the old GM_PASSCODE path is retired.
//
// Also useful any time you need to hand-grant access without going through
// the normal invite-by-email flow (e.g. local dev).
//
//	go run ./cmd/claim-owner --config-name local --instance-id <uuid> --email you@example.com
//	go run ./cmd/claim-owner --config-name local --instance-id <uuid> --phone +15555550123
package main

import (
	"context"
	"flag"
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/vampire-ascendancy/internal/config"
	"github.com/google/uuid"
)

func main() {
	instanceIDFlag := flag.String("instance-id", "", "The instance (Toast) to grant Host on. Required.")
	email := flag.String("email", "", "Look up the user by email.")
	phone := flag.String("phone", "", "Look up the user by phone number, e.g. +15555550123.")
	role := flag.String("role", "owner", "Role to grant: owner (Host) or admin (Co-Host).")

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if *instanceIDFlag == "" {
		log.Fatalf("--instance-id is required")
	}
	instanceID, err := uuid.Parse(*instanceIDFlag)
	if err != nil {
		log.Fatalf("invalid --instance-id: %v", err)
	}
	if (*email == "") == (*phone == "") {
		log.Fatalf("pass exactly one of --email or --phone")
	}
	if *role != models.InstanceAdminRoleOwner && *role != models.InstanceAdminRoleAdmin {
		log.Fatalf("--role must be %q or %q", models.InstanceAdminRoleOwner, models.InstanceAdminRoleAdmin)
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

	inst, err := dbClient.Vampire().GetInstanceByID(ctx, instanceID)
	if err != nil {
		log.Fatalf("failed to look up instance: %v", err)
	}
	if inst == nil {
		log.Fatalf("no instance %s found", instanceID)
	}

	admin, err := dbClient.Vampire().AddInstanceAdmin(ctx, instanceID, user.ID, *role)
	if err != nil {
		log.Fatalf("failed to grant access: %v", err)
	}

	log.Printf("%s (%s) is now %s on %q (%s)", user.Name, user.ID, admin.Role, inst.Name, inst.ID)
}
