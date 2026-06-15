// Command seed imports the Vampire Ascendancy character packets from
// seed/characters.json into the database. It is idempotent: houses and
// characters are upserted by name, and each character's secrets/missions are
// replaced wholesale, so re-running after editing the JSON re-imports cleanly
// without disturbing live players, submissions, or game state.
//
//	go run ./cmd/seed --config-name local --file seed/characters.json
package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/vampire-ascendancy/internal/config"
	"github.com/google/uuid"
)

// uniquePIN returns a 4-digit PIN not already in used, marking it used.
func uniquePIN(used map[string]bool) string {
	for {
		n, _ := rand.Int(rand.Reader, big.NewInt(10000))
		pin := fmt.Sprintf("%04d", n.Int64())
		if !used[pin] {
			used[pin] = true
			return pin
		}
	}
}

type seedFile struct {
	Houses []struct {
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
	} `json:"houses"`
	Characters []seedCharacter `json:"characters"`
}

type seedCharacter struct {
	Name            string        `json:"name"`
	Title           string        `json:"title"`
	House           string        `json:"house"`
	RoleType        string        `json:"role_type"`
	IsOptional      bool          `json:"is_optional"`
	PreEventInfo    string        `json:"pre_event_info"`
	PostAct1Context string        `json:"post_act1_context"`
	Secrets         []string      `json:"secrets"`
	Missions        []seedMission `json:"missions"`
}

type seedMission struct {
	Ordinal      int    `json:"ordinal"`
	Tier         string `json:"tier"`
	RewardBT     int    `json:"reward_bt"`
	Prompt       string `json:"prompt"`
	AnswerFormat string `json:"answer_format"`
}

type seedQuizQuestion struct {
	Ordinal       int            `json:"ordinal"`
	Prompt        string         `json:"prompt"`
	QuestionType  string         `json:"questionType"`
	Options       []string       `json:"options"`
	CorrectAnswer string         `json:"correctAnswer"`
	HFEffect      map[string]int `json:"hfEffect"`
}

func main() {
	// Register our own flags before config.ParseFlagsAndGetConfig calls flag.Parse.
	filePath := flag.String("file", "seed/characters.json", "Path to the characters.json seed file.")
	quizFile := flag.String("quiz-file", "seed/quiz.json", "Path to the quiz.json seed file (optional).")

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	raw, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("failed to read seed file %s: %v", *filePath, err)
	}
	var seed seedFile
	if err := json.Unmarshal(raw, &seed); err != nil {
		log.Fatalf("failed to parse seed file: %v", err)
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

	// Houses first; build a name -> id map for character assignment.
	houseIDs := map[string]uuid.UUID{}
	for _, h := range seed.Houses {
		house, err := v.UpsertHouse(ctx, h.Name, h.SortOrder)
		if err != nil {
			log.Fatalf("failed to upsert house %q: %v", h.Name, err)
		}
		houseIDs[h.Name] = house.ID
	}
	log.Printf("upserted %d houses", len(seed.Houses))

	for _, c := range seed.Characters {
		var houseID *uuid.UUID
		if id, ok := houseIDs[c.House]; ok {
			houseID = &id
		}

		character, err := v.UpsertCharacter(ctx, &models.VampireCharacter{
			Name:            c.Name,
			Title:           c.Title,
			HouseID:         houseID,
			RoleType:        c.RoleType,
			IsOptional:      c.IsOptional,
			PreEventInfo:    c.PreEventInfo,
			PostAct1Context: c.PostAct1Context,
		})
		if err != nil {
			log.Fatalf("failed to upsert character %q: %v", c.Name, err)
		}

		secrets := make([]models.VampireSecret, 0, len(c.Secrets))
		for i, body := range c.Secrets {
			secrets = append(secrets, models.VampireSecret{Ordinal: i + 1, Body: body})
		}
		if err := v.ReplaceSecrets(ctx, character.ID, secrets); err != nil {
			log.Fatalf("failed to replace secrets for %q: %v", c.Name, err)
		}

		missions := make([]models.VampireMission, 0, len(c.Missions))
		for _, m := range c.Missions {
			missions = append(missions, models.VampireMission{
				Ordinal:      m.Ordinal,
				Tier:         m.Tier,
				RewardBT:     m.RewardBT,
				Prompt:       m.Prompt,
				AnswerFormat: m.AnswerFormat,
			})
		}
		if err := v.ReplaceMissions(ctx, character.ID, missions); err != nil {
			log.Fatalf("failed to replace missions for %q: %v", c.Name, err)
		}
	}

	// Assign a unique sigil (4-digit PIN) to any character that lacks one.
	// Existing sigils are preserved so re-running the seed never changes a PIN
	// the GMs have already handed out.
	allChars, err := v.ListCharacters(ctx)
	if err != nil {
		log.Fatalf("failed to list characters: %v", err)
	}
	used := map[string]bool{}
	for _, c := range allChars {
		if c.Password != "" {
			used[c.Password] = true
		}
	}
	assigned := 0
	for _, c := range allChars {
		if c.Password != "" {
			continue
		}
		if err := v.SetCharacterPassword(ctx, c.ID, uniquePIN(used)); err != nil {
			log.Fatalf("failed to set sigil for %q: %v", c.Name, err)
		}
		assigned++
	}

	// Quiz questions (optional). Authored as a set, so this wholesale-replaces
	// the quiz — edit seed/quiz.json and re-run to update it.
	quizCount := 0
	if quizRaw, qerr := os.ReadFile(*quizFile); qerr == nil {
		var quizSeed []seedQuizQuestion
		if err := json.Unmarshal(quizRaw, &quizSeed); err != nil {
			log.Fatalf("failed to parse quiz file: %v", err)
		}
		questions := make([]models.VampireQuizQuestion, 0, len(quizSeed))
		for _, q := range quizSeed {
			opts, _ := json.Marshal(q.Options)
			if len(q.Options) == 0 {
				opts = []byte("[]")
			}
			eff, _ := json.Marshal(q.HFEffect)
			if len(q.HFEffect) == 0 {
				eff = []byte("{}")
			}
			qt := q.QuestionType
			if qt == "" {
				qt = "open"
			}
			questions = append(questions, models.VampireQuizQuestion{
				Ordinal:       q.Ordinal,
				Prompt:        q.Prompt,
				QuestionType:  qt,
				Options:       opts,
				CorrectAnswer: q.CorrectAnswer,
				HFEffect:      eff,
				Active:        true,
			})
		}
		if err := v.ReplaceQuizQuestions(ctx, questions); err != nil {
			log.Fatalf("failed to load quiz questions: %v", err)
		}
		quizCount = len(questions)
	}

	fmt.Printf("seeded %d characters across %d houses (assigned %d new sigils, %d quiz questions)\n",
		len(seed.Characters), len(seed.Houses), assigned, quizCount)
}
