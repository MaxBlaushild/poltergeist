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
	"encoding/base64"
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

// genToken returns a fresh opaque login token for a player slot, matching the
// format the GM "create player" endpoint uses.
func genToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

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
		Tagline   string `json:"tagline"`
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
	// ImageURL is the character's portrait (empty until image files are supplied).
	ImageURL string `json:"image_url"`
	// PlayerName is the real-world player assigned to this character (from the
	// roster spreadsheet). It seeds the GM Players tab's "Guest name" box and is
	// GM-only — it never reaches the player app. Optional.
	PlayerName string `json:"player_name"`
	// Active seeds whether this character's player slot starts active. Defaults
	// to true when omitted; set false for a character no one is playing yet.
	Active *bool `json:"active"`
}

type seedMission struct {
	Ordinal      int    `json:"ordinal"`
	Tier         string `json:"tier"`
	RewardBT     int    `json:"reward_bt"`
	Prompt       string `json:"prompt"`
	AnswerFormat string `json:"answer_format"`
	// Optional sabotage: verifying this mission deducts SabotageHF House Favor
	// from the named SabotageHouse.
	SabotageHouse string `json:"sabotage_house"`
	SabotageHF    int    `json:"sabotage_hf"`
}

type seedQuiz struct {
	Part1 struct {
		Prompt string `json:"prompt"`
		Rubric string `json:"rubric"`
		MaxBT  int    `json:"maxBt"`
	} `json:"part1"`
	Part2 []struct {
		Ordinal       int      `json:"ordinal"`
		Prompt        string   `json:"prompt"`
		Tier          string   `json:"tier"`
		HFValue       float64  `json:"hfValue"`
		Options       []string `json:"options"`
		CorrectAnswer string   `json:"correctAnswer"`
		Type          string   `json:"type"` // multiple_choice (default) | number
	} `json:"part2"`
}

func main() {
	// Register our own flags before config.ParseFlagsAndGetConfig calls flag.Parse.
	filePath := flag.String("file", "seed/characters.json", "Path to the characters.json seed file.")
	quizFile := flag.String("quiz-file", "seed/quiz.json", "Path to the quiz.json seed file (optional).")
	fresh := flag.Bool("fresh", false, "Wipe all existing characters and roster before seeding (from-scratch re-upload). Score ledgers are archived first.")

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

	// --fresh: clear the old roster + character content so this seed rebuilds from
	// scratch. Necessary when characters were renamed or removed, since the upsert
	// below is keyed by name and would otherwise leave orphaned characters behind.
	if *fresh {
		if err := v.WipeCharactersAndRoster(ctx); err != nil {
			log.Fatalf("failed to wipe characters and roster: %v", err)
		}
		log.Printf("--fresh: wiped existing characters, secrets, missions, and roster")
	}

	// Houses first; build a name -> id map for character assignment.
	houseIDs := map[string]uuid.UUID{}
	for _, h := range seed.Houses {
		house, err := v.UpsertHouse(ctx, h.Name, h.SortOrder, h.Tagline)
		if err != nil {
			log.Fatalf("failed to upsert house %q: %v", h.Name, err)
		}
		houseIDs[h.Name] = house.ID
	}
	log.Printf("upserted %d houses", len(seed.Houses))

	// Existing player slots keyed by character, so slot seeding is idempotent and
	// never duplicates a slot a GM has since marked inactive (a lookup filtered on
	// active=true would miss those and re-create them).
	existingPlayers, err := v.ListPlayers(ctx)
	if err != nil {
		log.Fatalf("failed to list players: %v", err)
	}
	slotByChar := map[uuid.UUID]models.VampirePlayer{}
	for _, p := range existingPlayers {
		if p.CharacterID != nil {
			slotByChar[*p.CharacterID] = p
		}
	}
	slotsCreated, slotsFilled := 0, 0

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
			ImageURL:        c.ImageURL,
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
			mission := models.VampireMission{
				Ordinal:      m.Ordinal,
				Tier:         m.Tier,
				RewardBT:     m.RewardBT,
				Prompt:       m.Prompt,
				AnswerFormat: m.AnswerFormat,
				SabotageHF:   m.SabotageHF,
			}
			if m.SabotageHouse != "" {
				if id, ok := houseIDs[m.SabotageHouse]; ok {
					hid := id
					mission.SabotageHouseID = &hid
				}
			}
			missions = append(missions, mission)
		}
		if err := v.ReplaceMissions(ctx, character.ID, missions); err != nil {
			log.Fatalf("failed to replace missions for %q: %v", c.Name, err)
		}

		// Roster: seed the GM Players tab from the spreadsheet assignment. Create
		// one login slot per character that has a player name, so GMs open the tab
		// pre-populated. The name lands in the GM-only guest-label field — it is
		// never sent to the player app.
		if c.PlayerName != "" {
			active := true
			if c.Active != nil {
				active = *c.Active
			}
			if slot, ok := slotByChar[character.ID]; ok {
				// Only fill an empty label, so we never clobber a day-of GM edit.
				if slot.GuestLabel == "" {
					if err := v.UpdatePlayerAssignment(ctx, slot.ID, slot.CharacterID, c.PlayerName, slot.Active); err != nil {
						log.Fatalf("failed to update player slot for %q: %v", c.Name, err)
					}
					slotsFilled++
				}
			} else {
				token, err := genToken()
				if err != nil {
					log.Fatalf("failed to generate token for %q: %v", c.Name, err)
				}
				cid := character.ID
				if err := v.CreatePlayer(ctx, &models.VampirePlayer{
					Token:       token,
					CharacterID: &cid,
					GuestLabel:  c.PlayerName,
					Active:      active,
				}); err != nil {
					log.Fatalf("failed to create player slot for %q: %v", c.Name, err)
				}
				slotsCreated++
			}
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
		var quizSeed seedQuiz
		if err := json.Unmarshal(quizRaw, &quizSeed); err != nil {
			log.Fatalf("failed to parse quiz file: %v", err)
		}
		questions := make([]models.VampireQuizQuestion, 0, len(quizSeed.Part2)+1)

		// Part 1: the single open-end prompt (AI-graded → BT).
		if quizSeed.Part1.Prompt != "" {
			maxBT := quizSeed.Part1.MaxBT
			if maxBT <= 0 {
				maxBT = 6
			}
			questions = append(questions, models.VampireQuizQuestion{
				Part:         1,
				Ordinal:      0,
				Prompt:       quizSeed.Part1.Prompt,
				QuestionType: "open",
				Rubric:       quizSeed.Part1.Rubric,
				MaxBT:        maxBT,
				Active:       true,
			})
		}

		// Part 2: the multiple-choice questions (normalized → HF).
		for _, q := range quizSeed.Part2 {
			opts, _ := json.Marshal(q.Options)
			if len(q.Options) == 0 {
				opts = []byte("[]")
			}
			qtype := q.Type
			if qtype == "" {
				qtype = "multiple_choice"
			}
			questions = append(questions, models.VampireQuizQuestion{
				Part:          2,
				Ordinal:       q.Ordinal,
				Prompt:        q.Prompt,
				QuestionType:  qtype,
				Options:       opts,
				CorrectAnswer: q.CorrectAnswer,
				HFValue:       q.HFValue,
				Tier:          q.Tier,
				Active:        true,
			})
		}

		if err := v.ReplaceQuizQuestions(ctx, questions); err != nil {
			log.Fatalf("failed to load quiz questions: %v", err)
		}
		quizCount = len(questions)
	}

	fmt.Printf("seeded %d characters across %d houses (assigned %d new sigils, %d quiz questions, %d player slots created, %d labels filled)\n",
		len(seed.Characters), len(seed.Houses), assigned, quizCount, slotsCreated, slotsFilled)
}
