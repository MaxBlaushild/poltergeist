package models

import (
	"time"

	"github.com/google/uuid"
)

type ResourceType struct {
	ID                 uuid.UUID                   `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt          time.Time                   `json:"createdAt"`
	UpdatedAt          time.Time                   `json:"updatedAt"`
	Name               string                      `json:"name"`
	Slug               string                      `json:"slug"`
	Description        string                      `json:"description"`
	MapIconURL         string                      `json:"mapIconUrl" gorm:"column:map_icon_url"`
	MapIconPrompt      string                      `json:"mapIconPrompt" gorm:"column:map_icon_prompt"`
	GatherRequirements []ResourceGatherRequirement `json:"gatherRequirements,omitempty" gorm:"foreignKey:ResourceTypeID"`
}

func (ResourceType) TableName() string {
	return "resource_types"
}
