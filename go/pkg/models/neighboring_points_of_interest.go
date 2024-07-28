package models

import (
	"time"

	"github.com/google/uuid"
)

type NeighboringPointsOfInterest struct {
	ID                   uuid.UUID       `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt            time.Time       `db:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at"`
	PointOfInterestOneID uuid.UUID       `json:"pointOfInterestOneId" binding:"required"`
	PointOfInterestOne   PointOfInterest `json:"pointOfInterestOne"`
	PointOfInterestTwoID uuid.UUID       `json:"pointOfInterestTwoId" binding:"required"`
	PointOfInterestTwo   PointOfInterest `json:"pointOfInterestTwo"`
}
