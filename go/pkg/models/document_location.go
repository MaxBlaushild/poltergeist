package models

import (
	"time"

	"github.com/google/uuid"
)

type LocationType string

const (
	LocationTypeCity      LocationType = "city"
	LocationTypeCountry   LocationType = "country"
	LocationTypeContinent LocationType = "continent"
)

type DocumentLocation struct {
	ID               uuid.UUID    `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
	DocumentID       uuid.UUID    `gorm:"type:uuid;not null" json:"documentId"`
	Document         Document     `json:"document" gorm:"foreignKey:DocumentID"`
	PlaceID          string       `gorm:"type:varchar;not null" json:"placeId"`
	Name             string       `gorm:"type:varchar;not null" json:"name"`
	FormattedAddress string       `gorm:"type:text;not null" json:"formattedAddress"`
	Latitude         float64      `gorm:"type:double precision;not null" json:"latitude"`
	Longitude        float64      `gorm:"type:double precision;not null" json:"longitude"`
	LocationType     LocationType `gorm:"type:varchar;not null" json:"locationType"`
}

func (DocumentLocation) TableName() string {
	return "document_locations"
}
