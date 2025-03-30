package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TagGroup struct {
	ID        uuid.UUID      `gorm:"primary_key" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	CreatedAt time.Time      `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"not null" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Tags      []Tag          `gorm:"foreignKey:TagGroupID" json:"tags"`
}

func (TagGroup) TableName() string {
	return "tag_groups"
}

func (t *TagGroup) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}
