package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultZoneGenreNameFantasy       = "Fantasy"
	defaultFantasyZoneGenrePromptSeed = "Keep the genre framing grounded in classic fantasy action RPG adventure: mythic beasts, arcane magic, dungeon ecology, swords-and-sorcery threats, and medieval-adjacent weapons, armor, and factions."
)

func DefaultFantasyZoneGenrePromptSeed() string {
	return defaultFantasyZoneGenrePromptSeed
}

func IsFantasyZoneGenreName(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), DefaultZoneGenreNameFantasy)
}

type ZoneGenre struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Name       string    `json:"name"`
	SortOrder  int       `json:"sortOrder" gorm:"column:sort_order"`
	Active     bool      `json:"active"`
	PromptSeed string    `json:"promptSeed" gorm:"column:prompt_seed"`
}

func (ZoneGenre) TableName() string {
	return "zone_genres"
}

type ZoneGenreScore struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	ZoneID    uuid.UUID  `json:"zoneId" gorm:"column:zone_id;type:uuid"`
	GenreID   uuid.UUID  `json:"genreId" gorm:"column:genre_id;type:uuid"`
	Score     int        `json:"score"`
	Genre     *ZoneGenre `json:"genre,omitempty" gorm:"foreignKey:GenreID"`
}

func (ZoneGenreScore) TableName() string {
	return "zone_genre_scores"
}
