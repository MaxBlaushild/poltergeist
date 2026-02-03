package models

import (
  "time"

  "github.com/google/uuid"
)

type CharacterLocation struct {
  ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
  CreatedAt   time.Time `json:"createdAt"`
  UpdatedAt   time.Time `json:"updatedAt"`
  CharacterID uuid.UUID `json:"characterId" gorm:"type:uuid"`
  Latitude    float64   `json:"latitude"`
  Longitude   float64   `json:"longitude"`
}

func (c *CharacterLocation) TableName() string {
  return "character_locations"
}
