package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	DefaultStatValue     = 10
	StatPointsPerLevel   = 5
	MaxStatValue         = 20
	MinStatValue         = 8
)

type UserStats struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
	UserID              uuid.UUID `json:"userId"`
	Strength            int       `json:"strength" gorm:"default:10"`
	Dexterity           int       `json:"dexterity" gorm:"default:10"`
	Constitution        int       `json:"constitution" gorm:"default:10"`
	Intelligence        int       `json:"intelligence" gorm:"default:10"`
	Wisdom              int       `json:"wisdom" gorm:"default:10"`
	Charisma            int       `json:"charisma" gorm:"default:10"`
	AvailableStatPoints int       `json:"availableStatPoints" gorm:"default:0"`
}

func (u *UserStats) TableName() string {
	return "user_stats"
}

func (u *UserStats) AfterFind(tx *gorm.DB) (err error) {
	return
}

// AddStatPoints adds stat points when a user levels up
func (u *UserStats) AddStatPoints(points int) {
	u.AvailableStatPoints += points
}

// AllocateStatPoint allocates a single stat point to a specific stat
func (u *UserStats) AllocateStatPoint(statName string) error {
	if u.AvailableStatPoints <= 0 {
		return NewValidationError("No available stat points to allocate")
	}

	switch statName {
	case "strength":
		if u.Strength >= MaxStatValue {
			return NewValidationError("Strength is already at maximum value")
		}
		u.Strength++
	case "dexterity":
		if u.Dexterity >= MaxStatValue {
			return NewValidationError("Dexterity is already at maximum value")
		}
		u.Dexterity++
	case "constitution":
		if u.Constitution >= MaxStatValue {
			return NewValidationError("Constitution is already at maximum value")
		}
		u.Constitution++
	case "intelligence":
		if u.Intelligence >= MaxStatValue {
			return NewValidationError("Intelligence is already at maximum value")
		}
		u.Intelligence++
	case "wisdom":
		if u.Wisdom >= MaxStatValue {
			return NewValidationError("Wisdom is already at maximum value")
		}
		u.Wisdom++
	case "charisma":
		if u.Charisma >= MaxStatValue {
			return NewValidationError("Charisma is already at maximum value")
		}
		u.Charisma++
	default:
		return NewValidationError("Invalid stat name")
	}

	u.AvailableStatPoints--
	return nil
}

// GetStatValue returns the value of a specific stat
func (u *UserStats) GetStatValue(statName string) (int, error) {
	switch statName {
	case "strength":
		return u.Strength, nil
	case "dexterity":
		return u.Dexterity, nil
	case "constitution":
		return u.Constitution, nil
	case "intelligence":
		return u.Intelligence, nil
	case "wisdom":
		return u.Wisdom, nil
	case "charisma":
		return u.Charisma, nil
	default:
		return 0, NewValidationError("Invalid stat name")
	}
}

// GetStatModifier returns the D&D stat modifier for a given stat value
func (u *UserStats) GetStatModifier(statName string) (int, error) {
	value, err := u.GetStatValue(statName)
	if err != nil {
		return 0, err
	}
	return (value - 10) / 2, nil
}

// GetAllStats returns a map of all stats with their values
func (u *UserStats) GetAllStats() map[string]int {
	return map[string]int{
		"strength":     u.Strength,
		"dexterity":    u.Dexterity,
		"constitution": u.Constitution,
		"intelligence": u.Intelligence,
		"wisdom":       u.Wisdom,
		"charisma":     u.Charisma,
	}
}

// GetAllStatModifiers returns a map of all stat modifiers
func (u *UserStats) GetAllStatModifiers() map[string]int {
	modifiers := make(map[string]int)
	for stat := range u.GetAllStats() {
		modifier, _ := u.GetStatModifier(stat)
		modifiers[stat] = modifier
	}
	return modifiers
}