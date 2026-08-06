// Command provision creates player slots with unique opaque tokens — one per
// playable character in one instance ("Toast") that does not already have a
// player, plus a fresh sigil for each. Each token is the per-character
// link/QR a guest uses to authenticate. Re-running only fills in characters
// that are still missing a player, so it is safe to run repeatedly as the
// roster firms up.
//
// This is the ops-only CLI equivalent of the in-app "Provision seats" button
// (POST /gm/players/provision) — self-serve Hosts without shell/DB access
// use that instead. See MULTI_TENANT_REQUIREMENTS.md.
//
//	go run ./cmd/provision --config-name local --instance-id <uuid> --base-url https://vampire-ascendancy.blaubertech.com
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"math/big"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/vampire-ascendancy/internal/config"
	"github.com/google/uuid"
)

func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// uniqueSigil returns a 4-digit PIN not already in used, marking it used.
func uniqueSigil(used map[string]bool) string {
	for {
		n, _ := rand.Int(rand.Reader, big.NewInt(10000))
		pin := fmt.Sprintf("%04d", n.Int64())
		if !used[pin] {
			used[pin] = true
			return pin
		}
	}
}

func main() {
	instanceIDFlag := flag.String("instance-id", "", "The instance (Toast) to provision player slots for. Required.")
	baseURL := flag.String("base-url", "https://vampire-ascendancy.blaubertech.com", "Base URL used to print shareable player links.")
	includeOptional := flag.Bool("include-optional", false, "Also provision players for optional (✦) characters.")

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

	// Characters that already have a player slot in this instance.
	existingPlayers, err := v.ListPlayers(ctx, instanceID)
	if err != nil {
		log.Fatalf("failed to list players: %v", err)
	}
	taken := map[string]bool{}
	for _, p := range existingPlayers {
		if p.CharacterID != nil {
			taken[p.CharacterID.String()] = true
		}
	}

	characters, err := v.ListIncludedCharacters(ctx, instanceID)
	if err != nil {
		log.Fatalf("failed to list this instance's characters: %v", err)
	}
	libraryChars, err := v.ListLibraryCharacters(ctx, instanceID)
	if err != nil {
		log.Fatalf("failed to list library characters: %v", err)
	}
	usedSigils := map[string]bool{}
	for _, c := range libraryChars {
		if c.Sigil != "" {
			usedSigils[c.Sigil] = true
		}
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
			InstanceID:  instanceID,
			Token:       token,
			CharacterID: &characterID,
			GuestLabel:  "",
			Active:      true,
		}
		if err := v.CreatePlayer(ctx, player); err != nil {
			log.Fatalf("failed to create player for %q: %v", c.Name, err)
		}

		sigil := uniqueSigil(usedSigils)
		if err := v.SetInstanceCharacterSigil(ctx, instanceID, c.ID, sigil); err != nil {
			log.Fatalf("failed to set sigil for %q: %v", c.Name, err)
		}
		created++
		// The link carries the character id (a pre-selector, not a secret); the
		// guest also needs the character's sigil to actually log in.
		fmt.Printf("%-28s sigil %-5s %s/e/%s/c/%s\n", c.Name, sigil, *baseURL, instanceID, c.ID)
	}

	fmt.Printf("\nprovisioned %d new player(s)\n", created)
}
