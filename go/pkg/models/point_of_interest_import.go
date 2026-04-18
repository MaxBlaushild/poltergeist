package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestImport struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	PlaceID           string     `json:"placeId"`
	ZoneID            uuid.UUID  `json:"zoneId" gorm:"type:uuid"`
	GenreID           uuid.UUID  `json:"genreId" gorm:"column:genre_id;type:uuid"`
	Genre             *ZoneGenre `json:"genre,omitempty" gorm:"foreignKey:GenreID"`
	Status            string     `json:"status"`
	ErrorMessage      *string    `json:"errorMessage"`
	PointOfInterestID *uuid.UUID `json:"pointOfInterestId" gorm:"type:uuid"`
}

func (p *PointOfInterestImport) TableName() string {
	return "point_of_interest_imports"
}
