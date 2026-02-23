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
	Question       string                 `json:"question"`
	SubmissionType QuestNodeSubmissionType `json:"submissionType"`
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
	ID             uuid.UUID                 `json:"id"`
	Name           string                    `json:"name"`
	CreatedAt      time.Time                 `json:"createdAt"`
	UpdatedAt      time.Time                 `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt            `json:"deletedAt"`
	IncludedTypes  googlemaps.PlaceTypeSlice `json:"includedTypes" gorm:"type:text[]"`
	ExcludedTypes  googlemaps.PlaceTypeSlice `json:"excludedTypes" gorm:"type:text[]"`
	Challenges     LocationArchetypeChallenges `json:"challenges" gorm:"type:jsonb"`
	UsedChallenges []string                  `gorm:"-" json:"usedChallenges"`
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
		question := challenge.Question
		if question == "" {
			continue
		}
		if !usedMap[question] {
			availableChallenges = append(availableChallenges, challenge)
		}
	}

	if len(availableChallenges) == 0 {
		return LocationArchetypeChallenge{}, errors.New("all challenges have been used")
	}

	// Pick random challenge from available ones
	challenge := availableChallenges[rand.Intn(len(availableChallenges))]
	if strings.TrimSpace(string(challenge.SubmissionType)) == "" {
		challenge.SubmissionType = DefaultQuestNodeSubmissionType()
	}
	l.UsedChallenges = append(l.UsedChallenges, challenge.Question)
	return challenge, nil
}
