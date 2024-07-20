package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarActivity struct {
	ID              uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt       time.Time     `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time     `db:"updated_at" json:"updatedAt"`
	Title           string        `gorm:"unique" json:"title"`
	SonarCategoryID uuid.UUID     `gorm:"type:uuid" json:"categoryId"`
	SonarCategory   SonarCategory `gorm:"foreignKey:SonarCategoryID" json:"category"`
	UserID          *uuid.UUID    `json:"userId"`
	User            *User         `gorm:"foreignKey:UserID" json:"user"`
}
