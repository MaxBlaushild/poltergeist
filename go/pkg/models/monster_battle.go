package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MonsterBattleState string

const (
	MonsterBattleStatePendingPartyResponses MonsterBattleState = "pending_party_responses"
	MonsterBattleStateActive                MonsterBattleState = "active"
)

type MonsterBattleAbilityCooldowns map[string]time.Time

func (c MonsterBattleAbilityCooldowns) Value() (driver.Value, error) {
	if c == nil {
		return json.Marshal(map[string]time.Time{})
	}
	return json.Marshal(c)
}

func (c *MonsterBattleAbilityCooldowns) Scan(value interface{}) error {
	if value == nil {
		*c = MonsterBattleAbilityCooldowns{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*c = MonsterBattleAbilityCooldowns{}
		return nil
	}

	return json.Unmarshal(bytes, c)
}

type MonsterBattleLastAction struct {
	ActionType        string     `json:"actionType,omitempty"`
	ActorType         string     `json:"actorType,omitempty"`
	ActorUserID       *uuid.UUID `json:"actorUserId,omitempty"`
	ActorMonsterID    *uuid.UUID `json:"actorMonsterId,omitempty"`
	ActorName         string     `json:"actorName,omitempty"`
	AbilityID         *uuid.UUID `json:"abilityId,omitempty"`
	AbilityName       string     `json:"abilityName,omitempty"`
	AbilityType       string     `json:"abilityType,omitempty"`
	TargetUserID      *uuid.UUID `json:"targetUserId,omitempty"`
	TargetMonsterID   *uuid.UUID `json:"targetMonsterId,omitempty"`
	TargetName        string     `json:"targetName,omitempty"`
	TargetsAllEnemies bool       `json:"targetsAllEnemies,omitempty"`
	Damage            int        `json:"damage,omitempty"`
	Heal              int        `json:"heal,omitempty"`
	StatusesApplied   int        `json:"statusesApplied,omitempty"`
	StatusesRemoved   int        `json:"statusesRemoved,omitempty"`
}

func (a MonsterBattleLastAction) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *MonsterBattleLastAction) Scan(value interface{}) error {
	if value == nil {
		*a = MonsterBattleLastAction{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*a = MonsterBattleLastAction{}
		return nil
	}
	if len(bytes) == 0 {
		*a = MonsterBattleLastAction{}
		return nil
	}

	return json.Unmarshal(bytes, a)
}

type MonsterBattle struct {
	ID                      uuid.UUID                     `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt               time.Time                     `json:"createdAt"`
	UpdatedAt               time.Time                     `json:"updatedAt"`
	UserID                  uuid.UUID                     `json:"userId" gorm:"column:user_id"`
	MonsterID               uuid.UUID                     `json:"monsterId" gorm:"column:monster_id"`
	MonsterEncounterID      *uuid.UUID                    `json:"monsterEncounterId,omitempty" gorm:"column:monster_encounter_id"`
	State                   string                        `json:"state" gorm:"column:state"`
	TurnIndex               int                           `json:"turnIndex" gorm:"column:turn_index"`
	StartedAt               time.Time                     `json:"startedAt" gorm:"column:started_at"`
	LastActivityAt          time.Time                     `json:"lastActivityAt" gorm:"column:last_activity_at"`
	MonsterHealthDeficit    int                           `json:"monsterHealthDeficit" gorm:"column:monster_health_deficit;default:0"`
	MonsterManaDeficit      int                           `json:"monsterManaDeficit" gorm:"column:monster_mana_deficit;default:0"`
	MonsterAbilityCooldowns MonsterBattleAbilityCooldowns `json:"monsterAbilityCooldowns" gorm:"column:monster_ability_cooldowns;type:jsonb;default:'{}'"`
	LastActionSequence      int                           `json:"lastActionSequence" gorm:"column:last_action_sequence;default:0"`
	LastAction              MonsterBattleLastAction       `json:"lastAction" gorm:"column:last_action;type:jsonb;default:'{}'"`
	EndedAt                 *time.Time                    `json:"endedAt,omitempty" gorm:"column:ended_at"`
}

func (m *MonsterBattle) TableName() string {
	return "monster_battles"
}
