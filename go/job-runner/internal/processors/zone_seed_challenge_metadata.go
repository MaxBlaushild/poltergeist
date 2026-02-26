package processors

import (
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

type zoneSeedChallengeMetadata struct {
	Question   string
	Difficulty int
}

func buildZoneSeedChallengeMetadata(
	place googlemaps.Place,
	proposedQuestion string,
	proposedDifficulty *int,
) zoneSeedChallengeMetadata {
	question := strings.TrimSpace(proposedQuestion)
	if question == "" {
		question = fallbackSeedQuestChallengeQuestion(place)
	}
	question = normalizeSeedChallengeQuestion(question, place)

	difficulty := 0
	if proposedDifficulty != nil {
		difficulty = *proposedDifficulty
	}
	if difficulty <= 0 {
		difficulty = randomQuestDifficulty()
	}
	difficulty = clampQuestDifficulty(difficulty)

	return zoneSeedChallengeMetadata{
		Question:   question,
		Difficulty: difficulty,
	}
}

func regenerateZoneSeedChallengeMetadata(place googlemaps.Place) zoneSeedChallengeMetadata {
	return buildZoneSeedChallengeMetadata(place, "", nil)
}

func zoneSeedDraftPOIToGooglePlace(poi models.ZoneSeedPointOfInterestDraft) googlemaps.Place {
	place := googlemaps.Place{
		ID:    strings.TrimSpace(poi.PlaceID),
		Types: append([]string(nil), poi.Types...),
	}
	place.DisplayName.Text = strings.TrimSpace(poi.Name)
	place.PrimaryType = strings.TrimSpace(firstNonEmptyString(poi.Types...))
	place.Location.Latitude = poi.Latitude
	place.Location.Longitude = poi.Longitude
	place.Rating = poi.Rating
	if poi.UserRatingCount > 0 {
		count := poi.UserRatingCount
		place.UserRatingCount = &count
	}
	place.FormattedAddress = strings.TrimSpace(poi.Address)
	place.EditorialSummary.Text = strings.TrimSpace(poi.EditorialSummary)
	return place
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
