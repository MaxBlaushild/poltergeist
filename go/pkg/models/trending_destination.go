package models

import (
	"time"

	"github.com/google/uuid"
)

type TrendingDestination struct {
	ID               uuid.UUID    `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
	LocationType     LocationType `gorm:"type:varchar;not null" json:"locationType"`
	PlaceID          string       `gorm:"type:varchar;not null" json:"placeId"`
	Name             string       `gorm:"type:varchar;not null" json:"name"`
	FormattedAddress string       `gorm:"type:text;not null" json:"formattedAddress"`
	DocumentCount    int          `gorm:"type:integer;not null" json:"documentCount"`
	Rank             int          `gorm:"type:integer;not null" json:"rank"`
	Latitude         float64      `gorm:"type:double precision;not null" json:"latitude"`
	Longitude        float64      `gorm:"type:double precision;not null" json:"longitude"`
}

func (TrendingDestination) TableName() string {
	return "trending_destinations"
}
