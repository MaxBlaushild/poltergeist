package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type District struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Zones       []Zone         `json:"zones" gorm:"many2many:district_zones;"`
}

func (District) TableName() string {
	return "districts"
}
