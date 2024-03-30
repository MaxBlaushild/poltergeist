package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarCategory struct {
	ID              uuid.UUID       `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt       time.Time       `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updatedAt"`
	Title           string          `gorm:"unique" json:"title"`
	SonarActivities []SonarActivity `gorm:"foreignKey:SonarCategoryID" json:"activities"`
}
