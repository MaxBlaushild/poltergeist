package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"golang.org/x/exp/rand"
	"gorm.io/gorm"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/google/uuid"
)

type LocationArchetypeChallenge struct {
	Question       string                  `json:"question"`
	SubmissionType QuestNodeSubmissionType `json:"submissionType"`
	Proficiency    *string                 `json:"proficiency"`
	Difficulty     int                     `json:"difficulty"`
}

type LocationArchetypeChallenges []LocationArchetypeChallenge

func (c LocationArchetypeChallenges) Value() (driver.Value, error) {
	if c == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(c)
}

func (c *LocationArchetypeChallenges) Scan(value interface{}) error {
	if value == nil {
		*c = LocationArchetypeChallenges{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return json.Unmarshal([]byte("[]"), c)
	}

	return json.Unmarshal(bytes, c)
}

type LocationArchetype struct {
	ID             uuid.UUID                   `json:"id"`
	Name           string                      `json:"name"`
	CreatedAt      time.Time                   `json:"createdAt"`
	UpdatedAt      time.Time                   `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt              `json:"deletedAt"`
	IncludedTypes  googlemaps.PlaceTypeSlice   `json:"includedTypes" gorm:"type:text[]"`
	ExcludedTypes  googlemaps.PlaceTypeSlice   `json:"excludedTypes" gorm:"type:text[]"`
	Challenges     LocationArchetypeChallenges `json:"challenges" gorm:"type:jsonb"`
	UsedChallenges []string                    `gorm:"-" json:"usedChallenges"`
}

func normalizeLocationArchetypeChallenge(challenge LocationArchetypeChallenge) LocationArchetypeChallenge {
	if strings.TrimSpace(string(challenge.SubmissionType)) == "" {
		challenge.SubmissionType = DefaultQuestNodeSubmissionType()
	}
	if challenge.Proficiency != nil {
		trimmed := strings.TrimSpace(*challenge.Proficiency)
		if trimmed == "" {
			challenge.Proficiency = nil
		} else {
			challenge.Proficiency = &trimmed
		}
	}
	if challenge.Difficulty < 0 {
		challenge.Difficulty = 0
	}
	return challenge
}

func (l *LocationArchetype) GetRandomChallenge() (LocationArchetypeChallenge, error) {
	if len(l.Challenges) == 0 {
		return LocationArchetypeChallenge{}, errors.New("no challenges found")
	}

	// Create map of used challenges for O(1) lookup
	usedMap := make(map[string]bool)
	for _, used := range l.UsedChallenges {
		usedMap[used] = true
	}

	// Get available challenges by filtering out used ones
	availableChallenges := make([]LocationArchetypeChallenge, 0)
	for _, challenge := range l.Challenges {
		question := strings.TrimSpace(challenge.Question)
		if question == "" {
			continue
		}
		if !usedMap[question] {
			challenge.Question = question
			availableChallenges = append(availableChallenges, normalizeLocationArchetypeChallenge(challenge))
		}
	}

	if len(availableChallenges) == 0 {
		return LocationArchetypeChallenge{}, errors.New("all challenges have been used")
	}

	// Pick random challenge from available ones
	challenge := availableChallenges[rand.Intn(len(availableChallenges))]
	l.UsedChallenges = append(l.UsedChallenges, challenge.Question)
	return challenge, nil
}

func (l *LocationArchetype) GetRandomChallengeByDifficulty(targetDifficulty int) (LocationArchetypeChallenge, error) {
	if len(l.Challenges) == 0 {
		return LocationArchetypeChallenge{}, errors.New("no challenges found")
	}

	usedMap := make(map[string]bool)
	for _, used := range l.UsedChallenges {
		usedMap[used] = true
	}

	availableChallenges := make([]LocationArchetypeChallenge, 0)
	for _, challenge := range l.Challenges {
		question := strings.TrimSpace(challenge.Question)
		if question == "" {
			continue
		}
		if !usedMap[question] {
			challenge.Question = question
			availableChallenges = append(availableChallenges, normalizeLocationArchetypeChallenge(challenge))
		}
	}

	if len(availableChallenges) == 0 {
		return LocationArchetypeChallenge{}, errors.New("all challenges have been used")
	}

	if targetDifficulty < 0 {
		targetDifficulty = 0
	}

	threshold := targetDifficulty
	candidates := make([]LocationArchetypeChallenge, 0)
	for {
		candidates = candidates[:0]
		for _, challenge := range availableChallenges {
			diff := challenge.Difficulty - threshold
			if diff < 0 {
				diff = -diff
			}
			if diff <= 5 {
				candidates = append(candidates, challenge)
			}
		}
		if len(candidates) > 0 {
			break
		}
		if threshold <= 0 {
			minDifficulty := availableChallenges[0].Difficulty
			for _, challenge := range availableChallenges[1:] {
				if challenge.Difficulty < minDifficulty {
					minDifficulty = challenge.Difficulty
				}
			}
			for _, challenge := range availableChallenges {
				if challenge.Difficulty == minDifficulty {
					candidates = append(candidates, challenge)
				}
			}
			if len(candidates) == 0 {
				candidates = availableChallenges
			}
			break
		}
		threshold -= 5
	}

	challenge := candidates[rand.Intn(len(candidates))]
	l.UsedChallenges = append(l.UsedChallenges, challenge.Question)
	return challenge, nil
}
