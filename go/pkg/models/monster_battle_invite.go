package models

import (
	"time"

	"github.com/google/uuid"
)

type MonsterBattleInviteStatus string

const (
	MonsterBattleInviteStatusPending      MonsterBattleInviteStatus = "pending"
	MonsterBattleInviteStatusAccepted     MonsterBattleInviteStatus = "accepted"
	MonsterBattleInviteStatusDeclined     MonsterBattleInviteStatus = "declined"
	MonsterBattleInviteStatusAutoDeclined MonsterBattleInviteStatus = "auto_declined"
)

type MonsterBattleInvite struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	BattleID      uuid.UUID  `json:"battleId" gorm:"column:battle_id"`
	InviterUserID uuid.UUID  `json:"inviterUserId" gorm:"column:inviter_user_id"`
	InviteeUserID uuid.UUID  `json:"inviteeUserId" gorm:"column:invitee_user_id"`
	MonsterID     uuid.UUID  `json:"monsterId" gorm:"column:monster_id"`
	Status        string     `json:"status" gorm:"column:status"`
	ExpiresAt     time.Time  `json:"expiresAt" gorm:"column:expires_at"`
	RespondedAt   *time.Time `json:"respondedAt,omitempty" gorm:"column:responded_at"`

	Inviter User `json:"inviter,omitempty" gorm:"foreignKey:InviterUserID"`
	Invitee User `json:"invitee,omitempty" gorm:"foreignKey:InviteeUserID"`
}

func (m *MonsterBattleInvite) TableName() string {
	return "monster_battle_invites"
}
