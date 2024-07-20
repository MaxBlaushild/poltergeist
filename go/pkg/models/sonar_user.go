package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarUser struct {
	ID                uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt         time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time `db:"updated_at" json:"updatedAt"`
	ViewerID          uuid.UUID `db:"viewer_id" json:"viewerId"`
	Viewer            User      `gorm:"foreignKey:ViewerID" json:"viewer"`
	VieweeID          uuid.UUID `gorm:"type:uuid;foreignKey:ID;references:ID" db:"viewee_id" json:"vieweeId"`
	Viewee            User      `gorm:"foreignKey:VieweeID" json:"viewee"`
	ProfilePictureUrl string    `db:"profile_picture_url" json:"profilePictureUrl"`
}
