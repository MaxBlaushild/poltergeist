package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutfitGenerationStatusQueued     = "queued"
	OutfitGenerationStatusInProgress = "in_progress"
	OutfitGenerationStatusComplete   = "complete"
	OutfitGenerationStatusFailed     = "failed"
)

type OutfitProfileGeneration struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
	UserID                uuid.UUID `json:"userId"`
	OwnedInventoryItemID  uuid.UUID `json:"ownedInventoryItemId"`
	InventoryItemID       int       `json:"inventoryItemId"`
	OutfitName            string    `json:"outfitName"`
	SelfieUrl             string    `json:"selfieUrl"`
	Status                string    `json:"status"`
	ErrorMessage          *string   `json:"error,omitempty" gorm:"column:error_message"`
	ProfilePictureUrl     *string   `json:"profilePictureUrl,omitempty" gorm:"column:profile_picture_url"`
}

func (OutfitProfileGeneration) TableName() string {
	return "outfit_profile_generations"
}
