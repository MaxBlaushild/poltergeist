// Command provision creates player slots with unique opaque tokens — one per
// playable character that does not already have a player. Each token is the
// per-character link/QR a guest uses to authenticate. Re-running only fills in
// characters that are still missing a player, so it is safe to run repeatedly
// as the roster firms up.
//
//	go run ./cmd/provision --config-name local --base-url https://vampire-ascendancy.blaubertech.com
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/vampire-ascendancy/internal/config"
)

func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func main() {
	baseURL := flag.String("base-url", "https://vampire-ascendancy.blaubertech.com", "Base URL used to print shareable player links.")
	includeOptional := flag.Bool("include-optional", false, "Also provision players for optional (✦) characters.")

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
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
	v := dbClient.Vampire()

	// Characters that already have a player slot.
	existingPlayers, err := v.ListPlayers(ctx)
	if err != nil {
		log.Fatalf("failed to list players: %v", err)
	}
	taken := map[string]bool{}
	for _, p := range existingPlayers {
		if p.CharacterID != nil {
			taken[p.CharacterID.String()] = true
		}
	}

	characters, err := v.ListCharacters(ctx)
	if err != nil {
		log.Fatalf("failed to list characters: %v", err)
	}

	created := 0
	for _, c := range characters {
		if c.RoleType != "player" {
			continue // GM/NPC packets are not assigned to standard players
		}
		if c.IsOptional && !*includeOptional {
			continue
		}
		if taken[c.ID.String()] {
			continue
		}

		token, err := newToken()
		if err != nil {
			log.Fatalf("failed to generate token: %v", err)
		}
		characterID := c.ID
		player := &models.VampirePlayer{
			Token:       token,
			CharacterID: &characterID,
			GuestLabel:  "",
			Active:      true,
		}
		if err := v.CreatePlayer(ctx, player); err != nil {
			log.Fatalf("failed to create player for %q: %v", c.Name, err)
		}
		created++
		// The link carries the character id (a pre-selector, not a secret); the
		// guest also needs the character's sigil to actually log in.
		fmt.Printf("%-28s sigil %-5s %s/c/%s\n", c.Name, c.Password, *baseURL, c.ID)
	}

	fmt.Printf("\nprovisioned %d new player(s)\n", created)
}
