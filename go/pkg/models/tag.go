package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tag struct {
	ID         uuid.UUID      `gorm:"primary_key" json:"id"`
	TagGroupID uuid.UUID      `gorm:"not null" json:"tagGroupID"`
	Value      string         `gorm:"not null" json:"name"`
	CreatedAt  time.Time      `gorm:"not null" json:"createdAt"`
	UpdatedAt  time.Time      `gorm:"not null" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

func (Tag) TableName() string {
	return "tags"
}

func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}

func (t *Tag) BeforeUpdate(tx *gorm.DB) (err error) {
	t.UpdatedAt = time.Now()
	return
}

func (t *Tag) BeforeDelete(tx *gorm.DB) (err error) {
	t.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	return
}
