package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type DndClass struct {
	ID                        uuid.UUID      `db:"id" gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt                 time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt                 time.Time      `db:"updated_at" json:"updatedAt"`
	Name                      string         `json:"name" gorm:"unique;not null"`
	Description               string         `json:"description"`
	HitDie                    int            `json:"hitDie" gorm:"default:8;not null"`
	PrimaryAbility            string         `json:"primaryAbility"`
	SavingThrowProficiencies  pq.StringArray `json:"savingThrowProficiencies" gorm:"type:text[]"`
	SkillOptions             pq.StringArray `json:"skillOptions" gorm:"type:text[]"`
	EquipmentProficiencies   pq.StringArray `json:"equipmentProficiencies" gorm:"type:text[]"`
	SpellCastingAbility      *string        `json:"spellCastingAbility"`
	IsSpellcaster            bool           `json:"isSpellcaster" gorm:"default:false"`
	Active                   bool           `json:"active" gorm:"default:true"`
}