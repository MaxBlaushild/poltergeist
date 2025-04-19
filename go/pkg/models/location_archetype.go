package models

import (
	"errors"
	"time"

	"golang.org/x/exp/rand"
	"gorm.io/gorm"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/google/uuid"
)

type LocationArchetype struct {
	ID             uuid.UUID              `json:"id"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt         `json:"deletedAt"`
	IncludedTypes  []googlemaps.PlaceType `json:"includedTypes"`
	ExcludedTypes  []googlemaps.PlaceType `json:"excludedTypes"`
	Challenges     []string               `json:"challenges"`
	UsedChallenges []string               `gorm:"-" json:"usedChallenges"`
}

func (l *LocationArchetype) GetRandomChallenge() (string, error) {
	if len(l.Challenges) == 0 {
		return "", errors.New("no challenges found")
	}

	// Create map of used challenges for O(1) lookup
	usedMap := make(map[string]bool)
	for _, used := range l.UsedChallenges {
		usedMap[used] = true
	}

	// Get available challenges by filtering out used ones
	availableChallenges := make([]string, 0)
	for _, challenge := range l.Challenges {
		if !usedMap[challenge] {
			availableChallenges = append(availableChallenges, challenge)
		}
	}

	if len(availableChallenges) == 0 {
		return "", errors.New("all challenges have been used")
	}

	// Pick random challenge from available ones
	challenge := availableChallenges[rand.Intn(len(availableChallenges))]
	l.UsedChallenges = append(l.UsedChallenges, challenge)
	return challenge, nil
}
