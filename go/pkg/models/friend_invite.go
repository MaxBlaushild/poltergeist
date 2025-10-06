package models

import (
	"time"

	"github.com/google/uuid"
)

type FriendInvite struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	InviterID uuid.UUID
	InviteeID uuid.UUID
	Invitee   User
	Inviter   User
}
