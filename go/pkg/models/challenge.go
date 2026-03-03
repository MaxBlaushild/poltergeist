package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Challenge struct {
	ID              uuid.UUID               `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt       time.Time               `json:"createdAt"`
	UpdatedAt       time.Time               `json:"updatedAt"`
	ZoneID          uuid.UUID               `json:"zoneId" gorm:"column:zone_id"`
	Zone            Zone                    `json:"zone"`
	Latitude        float64                 `json:"latitude"`
	Longitude       float64                 `json:"longitude"`
	Geometry        string                  `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Question        string                  `json:"question"`
	ImageURL        string                  `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL    string                  `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	Reward          int                     `json:"reward"`
	InventoryItemID *int                    `json:"inventoryItemId" gorm:"column:inventory_item_id"`
	SubmissionType  QuestNodeSubmissionType `json:"submissionType" gorm:"type:text;default:photo"`
	Difficulty      int                     `json:"difficulty" gorm:"default:0"`
	StatTags        StringArray             `json:"statTags,omitempty" gorm:"type:jsonb"`
	Proficiency     *string                 `json:"proficiency,omitempty"`
}

func (c *Challenge) TableName() string {
	return "challenges"
}

func (c *Challenge) BeforeSave(tx *gorm.DB) error {
	if err := c.SetGeometry(c.Latitude, c.Longitude); err != nil {
		return err
	}
	return nil
}

func (c *Challenge) SetGeometry(latitude, longitude float64) error {
	c.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}
