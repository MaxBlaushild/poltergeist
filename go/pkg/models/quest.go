package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	QuestRecurrenceDaily   = "daily"
	QuestRecurrenceWeekly  = "weekly"
	QuestRecurrenceMonthly = "monthly"

	QuestCategorySide      = "side"
	QuestCategoryMainStory = "main_story"
)

func NormalizeQuestRecurrenceFrequency(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func IsValidQuestRecurrenceFrequency(value string) bool {
	switch NormalizeQuestRecurrenceFrequency(value) {
	case QuestRecurrenceDaily, QuestRecurrenceWeekly, QuestRecurrenceMonthly:
		return true
	default:
		return false
	}
}

func NormalizeQuestCategory(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case QuestCategoryMainStory:
		return QuestCategoryMainStory
	default:
		return QuestCategorySide
	}
}

func IsMainStoryQuestCategory(value string) bool {
	return NormalizeQuestCategory(value) == QuestCategoryMainStory
}

func NextQuestRecurrenceAt(base time.Time, frequency string) (time.Time, bool) {
	switch NormalizeQuestRecurrenceFrequency(frequency) {
	case QuestRecurrenceDaily:
		return base.AddDate(0, 0, 1), true
	case QuestRecurrenceWeekly:
		return base.AddDate(0, 0, 7), true
	case QuestRecurrenceMonthly:
		return base.AddDate(0, 1, 0), true
	default:
		return time.Time{}, false
	}
}

type Quest struct {
	ID                             uuid.UUID                  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                      time.Time                  `json:"createdAt"`
	UpdatedAt                      time.Time                  `json:"updatedAt"`
	NodeCount                      int                        `json:"nodeCount,omitempty" gorm:"column:node_count;->"`
	Name                           string                     `json:"name"`
	Description                    string                     `json:"description"`
	Category                       string                     `json:"category" gorm:"column:category;default:'side'"`
	RequiredStoryFlags             StringArray                `json:"requiredStoryFlags" gorm:"column:required_story_flags;type:jsonb;default:'[]'"`
	SetStoryFlags                  StringArray                `json:"setStoryFlags" gorm:"column:set_story_flags;type:jsonb;default:'[]'"`
	ClearStoryFlags                StringArray                `json:"clearStoryFlags" gorm:"column:clear_story_flags;type:jsonb;default:'[]'"`
	QuestGiverRelationshipEffects  CharacterRelationshipState `json:"questGiverRelationshipEffects" gorm:"column:quest_giver_relationship_effects;type:jsonb;default:'{}'"`
	ClosurePolicy                  QuestClosurePolicy         `json:"closurePolicy" gorm:"column:closure_policy;default:'remote'"`
	DebriefPolicy                  QuestDebriefPolicy         `json:"debriefPolicy" gorm:"column:debrief_policy;default:'optional'"`
	ReturnBonusGold                int                        `json:"returnBonusGold" gorm:"column:return_bonus_gold;default:0"`
	ReturnBonusExperience          int                        `json:"returnBonusExperience" gorm:"column:return_bonus_experience;default:0"`
	ReturnBonusRelationshipEffects CharacterRelationshipState `json:"returnBonusRelationshipEffects" gorm:"column:return_bonus_relationship_effects;type:jsonb;default:'{}'"`
	AcceptanceDialogue             DialogueSequence           `json:"acceptanceDialogue,omitempty" gorm:"type:jsonb"`
	ImageURL                       string                     `json:"imageUrl"`
	OwnerUserID                    *uuid.UUID                 `json:"ownerUserId,omitempty" gorm:"column:owner_user_id;type:uuid"`
	Ephemeral                      bool                       `json:"ephemeral" gorm:"column:ephemeral"`
	ZoneID                         *uuid.UUID                 `json:"zoneId" gorm:"type:uuid"`
	QuestArchetypeID               *uuid.UUID                 `json:"questArchetypeId" gorm:"type:uuid"`
	QuestGiverCharacterID          *uuid.UUID                 `json:"questGiverCharacterId" gorm:"type:uuid"`
	MainStoryPreviousQuestID       *uuid.UUID                 `json:"mainStoryPreviousQuestId,omitempty" gorm:"column:main_story_previous_quest_id;type:uuid"`
	MainStoryNextQuestID           *uuid.UUID                 `json:"mainStoryNextQuestId,omitempty" gorm:"column:main_story_next_quest_id;type:uuid"`
	RecurringQuestID               *uuid.UUID                 `json:"recurringQuestId,omitempty" gorm:"type:uuid"`
	RecurrenceFrequency            *string                    `json:"recurrenceFrequency,omitempty"`
	NextRecurrenceAt               *time.Time                 `json:"nextRecurrenceAt,omitempty"`
	DifficultyMode                 QuestDifficultyMode        `json:"difficultyMode" gorm:"column:difficulty_mode"`
	Difficulty                     int                        `json:"difficulty" gorm:"default:1"`
	MonsterEncounterTargetLevel    int                        `json:"monsterEncounterTargetLevel" gorm:"column:monster_encounter_target_level;default:1"`
	RewardMode                     RewardMode                 `json:"rewardMode" gorm:"column:reward_mode"`
	RandomRewardSize               RandomRewardSize           `json:"randomRewardSize" gorm:"column:random_reward_size"`
	RewardExperience               int                        `json:"rewardExperience" gorm:"column:reward_experience"`
	Gold                           int                        `json:"gold"`
	MaterialRewards                BaseMaterialRewards        `json:"materialRewards" gorm:"column:material_rewards_json;type:jsonb;default:'[]'"`
	ItemRewards                    []QuestItemReward          `json:"itemRewards" gorm:"foreignKey:QuestID"`
	SpellRewards                   []QuestSpellReward         `json:"spellRewards" gorm:"foreignKey:QuestID"`
	Nodes                          []QuestNode                `json:"nodes" gorm:"foreignKey:QuestID"`
}

func (q *Quest) TableName() string {
	return "quests"
}
