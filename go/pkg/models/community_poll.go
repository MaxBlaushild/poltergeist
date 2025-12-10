package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// StringArray is a custom type for handling JSONB string arrays in PostgreSQL
type StringArray []string

// Value implements the driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return json.Unmarshal([]byte{}, a)
	}

	return json.Unmarshal(bytes, a)
}

type CommunityPoll struct {
	ID        uuid.UUID   `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time   `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time   `gorm:"column:updated_at" json:"updatedAt"`
	UserID    uuid.UUID   `gorm:"type:uuid;column:user_id;not null" json:"userId"`
	User      User        `json:"user" gorm:"foreignKey:UserID"`
	Question  string      `gorm:"type:text;not null" json:"question"`
	Options   StringArray `gorm:"type:jsonb;not null" json:"options"`
}

func (CommunityPoll) TableName() string {
	return "community_polls"
}
