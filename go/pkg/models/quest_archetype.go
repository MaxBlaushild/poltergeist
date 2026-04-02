package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestArchetype struct {
	ID                            uuid.UUID                   `json:"id"`
	Name                          string                      `json:"name"`
	Description                   string                      `json:"description"`
	Category                      string                      `json:"category" gorm:"column:category;default:'side'"`
	QuestGiverCharacterID         *uuid.UUID                  `json:"questGiverCharacterId,omitempty" gorm:"column:quest_giver_character_id;type:uuid"`
	QuestGiverCharacter           *Character                  `json:"questGiverCharacter,omitempty" gorm:"foreignKey:QuestGiverCharacterID"`
	RequiredStoryFlags            StringArray                 `json:"requiredStoryFlags" gorm:"column:required_story_flags;type:jsonb;default:'[]'"`
	SetStoryFlags                 StringArray                 `json:"setStoryFlags" gorm:"column:set_story_flags;type:jsonb;default:'[]'"`
	ClearStoryFlags               StringArray                 `json:"clearStoryFlags" gorm:"column:clear_story_flags;type:jsonb;default:'[]'"`
	QuestGiverRelationshipEffects CharacterRelationshipState  `json:"questGiverRelationshipEffects" gorm:"column:quest_giver_relationship_effects;type:jsonb;default:'{}'"`
	AcceptanceDialogue            StringArray                 `json:"acceptanceDialogue,omitempty" gorm:"type:jsonb"`
	ImageURL                      string                      `json:"imageUrl"`
	DifficultyMode                QuestDifficultyMode         `json:"difficultyMode" gorm:"column:difficulty_mode"`
	Difficulty                    int                         `json:"difficulty" gorm:"default:1"`
	MonsterEncounterTargetLevel   int                         `json:"monsterEncounterTargetLevel" gorm:"column:monster_encounter_target_level;default:1"`
	DefaultGold                   int                         `json:"defaultGold"`
	RewardMode                    RewardMode                  `json:"rewardMode" gorm:"column:reward_mode"`
	RandomRewardSize              RandomRewardSize            `json:"randomRewardSize" gorm:"column:random_reward_size"`
	RewardExperience              int                         `json:"rewardExperience" gorm:"column:reward_experience"`
	RecurrenceFrequency           *string                     `json:"recurrenceFrequency,omitempty"`
	MaterialRewards               BaseMaterialRewards         `json:"materialRewards" gorm:"column:material_rewards_json;type:jsonb;default:'[]'"`
	CharacterTags                 StringArray                 `json:"characterTags" gorm:"column:character_tags;type:jsonb"`
	InternalTags                  StringArray                 `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	CreatedAt                     time.Time                   `json:"createdAt"`
	UpdatedAt                     time.Time                   `json:"updatedAt"`
	DeletedAt                     gorm.DeletedAt              `json:"deletedAt"`
	Root                          QuestArchetypeNode          `json:"root"`
	RootID                        uuid.UUID                   `json:"rootId"`
	ItemRewards                   []QuestArchetypeItemReward  `json:"itemRewards" gorm:"foreignKey:QuestArchetypeID"`
	SpellRewards                  []QuestArchetypeSpellReward `json:"spellRewards" gorm:"foreignKey:QuestArchetypeID"`
}
