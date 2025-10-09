package models

import (
	"time"

	"github.com/google/uuid"
)

type PartyInvite struct {
	ID        uuid.UUID `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	InviterID uuid.UUID `json:"inviterId"`
	InviteeID uuid.UUID `json:"inviteeId"`
	Invitee   User      `json:"invitee"`
	Inviter   User      `json:"inviter"`
}
