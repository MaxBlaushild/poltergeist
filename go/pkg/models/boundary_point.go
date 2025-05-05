package models

import (
	"time"

	"github.com/google/uuid"
)

type BoundaryPoint struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ZoneID    uuid.UUID `json:"zoneId"`
	PointID   uuid.UUID `json:"pointId"`
}
