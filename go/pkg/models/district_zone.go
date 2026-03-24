package models

import (
	"time"

	"github.com/google/uuid"
)

type DistrictZone struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	DistrictID uuid.UUID `json:"districtId" gorm:"type:uuid;not null;index"`
	ZoneID     uuid.UUID `json:"zoneId" gorm:"type:uuid;not null;index"`
	District   District  `json:"district"`
	Zone       Zone      `json:"zone"`
}

func (DistrictZone) TableName() string {
	return "district_zones"
}
