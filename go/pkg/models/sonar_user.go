package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarUser struct {
	ID                uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt         time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time `db:"updated_at" json:"updatedAt"`
	ProfilePictureUrl string    `db:"profile_picture_url" json:"profilePictureUrl"`
}
