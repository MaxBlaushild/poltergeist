package models

import (
	"time"

	"github.com/google/uuid"
)

type ZoneImport struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	MetroName    string    `json:"metroName"`
	Status       string    `json:"status"`
	ErrorMessage *string   `json:"errorMessage"`
	ZoneCount    int       `json:"zoneCount"`
}

func (z *ZoneImport) TableName() string {
	return "zone_imports"
}
