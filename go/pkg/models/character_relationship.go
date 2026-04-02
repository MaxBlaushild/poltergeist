package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type CharacterRelationshipState struct {
	Trust   int `json:"trust"`
	Respect int `json:"respect"`
	Fear    int `json:"fear"`
	Debt    int `json:"debt"`
}

type CharacterRelationshipDelta = CharacterRelationshipState

func (s CharacterRelationshipState) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *CharacterRelationshipState) Scan(value interface{}) error {
	if value == nil {
		*s = CharacterRelationshipState{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan CharacterRelationshipState: value is not []byte")
	}
	var decoded CharacterRelationshipState
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*s = decoded
	return nil
}

func (s CharacterRelationshipState) IsZero() bool {
	return s.Trust == 0 && s.Respect == 0 && s.Fear == 0 && s.Debt == 0
}

type UserCharacterRelationship struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	UserID      uuid.UUID `json:"userId" gorm:"column:user_id;type:uuid;index"`
	CharacterID uuid.UUID `json:"characterId" gorm:"column:character_id;type:uuid;index"`
	Trust       int       `json:"trust" gorm:"column:trust;default:0"`
	Respect     int       `json:"respect" gorm:"column:respect;default:0"`
	Fear        int       `json:"fear" gorm:"column:fear;default:0"`
	Debt        int       `json:"debt" gorm:"column:debt;default:0"`
}

func (UserCharacterRelationship) TableName() string {
	return "user_character_relationships"
}

func (u UserCharacterRelationship) State() CharacterRelationshipState {
	return CharacterRelationshipState{
		Trust:   u.Trust,
		Respect: u.Respect,
		Fear:    u.Fear,
		Debt:    u.Debt,
	}
}
