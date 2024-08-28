package quartermaster

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

func (q *client) ApplyInventoryItemEffects(ctx context.Context, userID uuid.UUID, match *models.Match) error {
	for _, effect := range match.InventoryItemEffects {
		if effect.ExpiresAt.Before(time.Now()) {
			continue
		}

		if err := q.applyInventoryItemEffect(ctx, userID, match, &effect); err != nil {
			return err
		}
	}
	return nil
}

func (q *client) applyInventoryItemEffect(ctx context.Context, userID uuid.UUID, match *models.Match, effect *models.MatchInventoryItemEffect) error {
	switch effect.InventoryItemID {
	case 1:
		// scramble them words
		usersTeam := models.Team{}
		for _, team := range match.Teams {
			for _, user := range team.Users {
				if user.ID == userID {
					usersTeam = team
					break
				}
			}
		}

		if usersTeam.ID != effect.TeamID {
			for i, pointOfInterest := range match.PointsOfInterest {
				match.PointsOfInterest[i].Clue = scrambleAndObscureWords(pointOfInterest.Clue, usersTeam.ID)

				for j, challenge := range pointOfInterest.PointOfInterestChallenges {
					match.PointsOfInterest[i].PointOfInterestChallenges[j].Question = scrambleAndObscureWords(challenge.Question, usersTeam.ID)
				}
			}

		}
	default:
		return fmt.Errorf("unknown inventory item ID: %d", effect.InventoryItemID)
	}
	return nil
}

func scrambleAndObscureWords(input string, seed uuid.UUID) string {
	hasher := md5.New()
	hasher.Write(seed[:])
	seedBytes := hasher.Sum(nil)
	seedInt := binary.BigEndian.Uint64(seedBytes[:8]) // Use the first 8 bytes to create a uint64 seed

	rand.Seed(uint64(seedInt))
	words := strings.Fields(input)
	for i, word := range words {
		letters := []rune(word)
		// Only shuffle alphabetic characters
		alphabeticLetters := make([]rune, 0)
		nonAlphabeticMapping := make(map[int]rune) // Store non-alphabetic characters and their positions

		for idx, letter := range letters {
			if unicode.IsLetter(letter) {
				alphabeticLetters = append(alphabeticLetters, letter)
			} else {
				nonAlphabeticMapping[idx] = letter
			}
		}

		rand.Shuffle(len(alphabeticLetters), func(i, j int) {
			alphabeticLetters[i], alphabeticLetters[j] = alphabeticLetters[j], alphabeticLetters[i]
		})

		// Reconstruct the word with scrambled alphabetic characters and original non-alphabetic characters
		j := 0
		for k := range letters {
			if _, exists := nonAlphabeticMapping[k]; exists {
				letters[k] = nonAlphabeticMapping[k]
			} else {
				letters[k] = alphabeticLetters[j]
				j++
			}
		}

		words[i] = string(letters)
	}
	return strings.Join(words, " ")
}
