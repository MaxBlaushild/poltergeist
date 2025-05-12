package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID  `db:"id" gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt         time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updatedAt"`
	Name              string     `json:"name"`
	PhoneNumber       string     `json:"phoneNumber" gorm:"unique"`
	Active            bool       `json:"active"`
	Profile           *SonarUser `json:"profile" gorm:"foreignKey:ID"`
	ProfilePictureUrl string     `json:"profilePictureUrl"`
	HasSeenTutorial   bool       `json:"hasSeenTutorial" gorm:"default:false"`
}
