package models

import (
	"time"

	"github.com/google/uuid"
)

type ZoneDiscovery struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uuid.UUID `json:"userId"`
	User      User      `json:"user"`
	ZoneID    uuid.UUID `json:"zoneId"`
	Zone      Zone      `json:"zone"`
}

func (ZoneDiscovery) TableName() string {
	return "zone_discoveries"
}
