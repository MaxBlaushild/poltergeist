package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestArchetypeNodeType string

const (
	QuestArchetypeNodeTypeChallenge        QuestArchetypeNodeType = "challenge"
	QuestArchetypeNodeTypeLocation         QuestArchetypeNodeType = "location"
	QuestArchetypeNodeTypeMonsterEncounter QuestArchetypeNodeType = "monster_encounter"
	QuestArchetypeNodeTypeScenario         QuestArchetypeNodeType = "scenario"
	QuestArchetypeNodeTypeExposition       QuestArchetypeNodeType = "exposition"
	QuestArchetypeNodeTypeFetchQuest       QuestArchetypeNodeType = "fetch_quest"
	QuestArchetypeNodeTypeStoryFlag        QuestArchetypeNodeType = "story_flag"
)

func NormalizeQuestArchetypeNodeType(raw string) QuestArchetypeNodeType {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(QuestArchetypeNodeTypeChallenge), string(QuestArchetypeNodeTypeLocation):
		return QuestArchetypeNodeTypeChallenge
	case string(QuestArchetypeNodeTypeExposition):
		return QuestArchetypeNodeTypeExposition
	case string(QuestArchetypeNodeTypeFetchQuest):
		return QuestArchetypeNodeTypeFetchQuest
	case string(QuestArchetypeNodeTypeStoryFlag):
		return QuestArchetypeNodeTypeStoryFlag
	case string(QuestArchetypeNodeTypeScenario):
		return QuestArchetypeNodeTypeScenario
	case string(QuestArchetypeNodeTypeMonsterEncounter):
		return QuestArchetypeNodeTypeMonsterEncounter
	default:
		return QuestArchetypeNodeTypeChallenge
	}
}

type QuestArchetypeNodeLocationSelectionMode string

const (
	QuestArchetypeNodeLocationSelectionModeRandom         QuestArchetypeNodeLocationSelectionMode = "random"
	QuestArchetypeNodeLocationSelectionModeClosest        QuestArchetypeNodeLocationSelectionMode = "closest"
	QuestArchetypeNodeLocationSelectionModeSameAsPrevious QuestArchetypeNodeLocationSelectionMode = "same_as_previous"
)

func NormalizeQuestArchetypeNodeLocationSelectionMode(
	raw string,
) QuestArchetypeNodeLocationSelectionMode {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(QuestArchetypeNodeLocationSelectionModeClosest):
		return QuestArchetypeNodeLocationSelectionModeClosest
	case string(QuestArchetypeNodeLocationSelectionModeSameAsPrevious):
		return QuestArchetypeNodeLocationSelectionModeSameAsPrevious
	default:
		return QuestArchetypeNodeLocationSelectionModeRandom
	}
}

type QuestArchetypeExpositionItemReward struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type QuestArchetypeExpositionItemRewards []QuestArchetypeExpositionItemReward

func (r QuestArchetypeExpositionItemRewards) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]QuestArchetypeExpositionItemReward{})
	}
	return json.Marshal(r)
}

func (r *QuestArchetypeExpositionItemRewards) Scan(value interface{}) error {
	if value == nil {
		*r = QuestArchetypeExpositionItemRewards{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan QuestArchetypeExpositionItemRewards: value is not []byte")
	}
	var decoded []QuestArchetypeExpositionItemReward
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*r = decoded
	return nil
}

type QuestArchetypeExpositionSpellReward struct {
	SpellID uuid.UUID `json:"spellId"`
}

type QuestArchetypeExpositionSpellRewards []QuestArchetypeExpositionSpellReward

func (r QuestArchetypeExpositionSpellRewards) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]QuestArchetypeExpositionSpellReward{})
	}
	return json.Marshal(r)
}

func (r *QuestArchetypeExpositionSpellRewards) Scan(value interface{}) error {
	if value == nil {
		*r = QuestArchetypeExpositionSpellRewards{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan QuestArchetypeExpositionSpellRewards: value is not []byte")
	}
	var decoded []QuestArchetypeExpositionSpellReward
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*r = decoded
	return nil
}

type QuestArchetypeNode struct {
	ID                         uuid.UUID                               `json:"id"`
	CreatedAt                  time.Time                               `json:"createdAt"`
	UpdatedAt                  time.Time                               `json:"updatedAt"`
	DeletedAt                  gorm.DeletedAt                          `json:"deletedAt"`
	NodeType                   QuestArchetypeNodeType                  `json:"nodeType" gorm:"column:node_type"`
	LocationArchetype          *LocationArchetype                      `json:"locationArchetype,omitempty"`
	LocationArchetypeID        *uuid.UUID                              `json:"locationArchetypeId,omitempty"`
	LocationSelectionMode      QuestArchetypeNodeLocationSelectionMode `json:"locationSelectionMode" gorm:"column:location_selection_mode"`
	ChallengeTemplate          *ChallengeTemplate                      `json:"challengeTemplate,omitempty"`
	ChallengeTemplateID        *uuid.UUID                              `json:"challengeTemplateId,omitempty" gorm:"column:challenge_template_id;type:uuid"`
	ScenarioTemplate           *ScenarioTemplate                       `json:"scenarioTemplate,omitempty"`
	ScenarioTemplateID         *uuid.UUID                              `json:"scenarioTemplateId,omitempty"`
	FetchCharacter             *Character                              `json:"fetchCharacter,omitempty" gorm:"foreignKey:FetchCharacterID"`
	FetchCharacterID           *uuid.UUID                              `json:"fetchCharacterId,omitempty" gorm:"column:fetch_character_id;type:uuid"`
	FetchCharacterTemplate     *CharacterTemplate                      `json:"fetchCharacterTemplate,omitempty" gorm:"foreignKey:FetchCharacterTemplateID"`
	FetchCharacterTemplateID   *uuid.UUID                              `json:"fetchCharacterTemplateId,omitempty" gorm:"column:fetch_character_template_id;type:uuid"`
	FetchRequirements          FetchQuestRequirements                  `json:"fetchRequirements" gorm:"column:fetch_requirements_json;type:jsonb;default:'[]'"`
	ObjectiveDescription       string                                  `json:"objectiveDescription,omitempty" gorm:"column:objective_description"`
	StoryFlagKey               string                                  `json:"storyFlagKey,omitempty" gorm:"column:story_flag_key"`
	MonsterTemplateIDs         StringArray                             `json:"monsterTemplateIds" gorm:"column:monster_template_ids;type:jsonb"`
	TargetLevel                int                                     `json:"targetLevel" gorm:"column:target_level;default:1"`
	EncounterRewardMode        RewardMode                              `json:"encounterRewardMode" gorm:"column:encounter_reward_mode"`
	EncounterRandomRewardSize  RandomRewardSize                        `json:"encounterRandomRewardSize" gorm:"column:encounter_random_reward_size"`
	EncounterRewardExperience  int                                     `json:"encounterRewardExperience" gorm:"column:encounter_reward_experience"`
	EncounterRewardGold        int                                     `json:"encounterRewardGold" gorm:"column:encounter_reward_gold"`
	EncounterMaterialRewards   BaseMaterialRewards                     `json:"encounterMaterialRewards" gorm:"column:encounter_material_rewards_json;type:jsonb;default:'[]'"`
	EncounterItemRewards       MonsterEncounterRewardItems             `json:"encounterItemRewards" gorm:"column:encounter_item_rewards_json;type:jsonb;default:'[]'"`
	EncounterProximityMeters   int                                     `json:"encounterProximityMeters" gorm:"column:encounter_proximity_meters;default:100"`
	ExpositionTemplate         *ExpositionTemplate                     `json:"expositionTemplate,omitempty" gorm:"foreignKey:ExpositionTemplateID"`
	ExpositionTemplateID       *uuid.UUID                              `json:"expositionTemplateId,omitempty" gorm:"column:exposition_template_id;type:uuid"`
	ExpositionTitle            string                                  `json:"expositionTitle" gorm:"column:exposition_title"`
	ExpositionDescription      string                                  `json:"expositionDescription" gorm:"column:exposition_description"`
	ExpositionDialogue         DialogueSequence                        `json:"expositionDialogue" gorm:"column:exposition_dialogue;type:jsonb;default:'[]'"`
	ExpositionRewardMode       RewardMode                              `json:"expositionRewardMode" gorm:"column:exposition_reward_mode"`
	ExpositionRandomRewardSize RandomRewardSize                        `json:"expositionRandomRewardSize" gorm:"column:exposition_random_reward_size"`
	ExpositionRewardExperience int                                     `json:"expositionRewardExperience" gorm:"column:exposition_reward_experience"`
	ExpositionRewardGold       int                                     `json:"expositionRewardGold" gorm:"column:exposition_reward_gold"`
	ExpositionMaterialRewards  BaseMaterialRewards                     `json:"expositionMaterialRewards" gorm:"column:exposition_material_rewards_json;type:jsonb;default:'[]'"`
	ExpositionItemRewards      QuestArchetypeExpositionItemRewards     `json:"expositionItemRewards" gorm:"column:exposition_item_rewards_json;type:jsonb;default:'[]'"`
	ExpositionSpellRewards     QuestArchetypeExpositionSpellRewards    `json:"expositionSpellRewards" gorm:"column:exposition_spell_rewards_json;type:jsonb;default:'[]'"`
	Challenges                 []QuestArchetypeChallenge               `json:"challenges" gorm:"many2many:quest_archetype_node_challenges;"`
	Difficulty                 int                                     `json:"difficulty" gorm:"default:0"`
}

func (q *QuestArchetypeNode) GetRandomChallenge() (LocationArchetypeChallenge, error) {
	if q == nil {
		return LocationArchetypeChallenge{}, fmt.Errorf("quest archetype node is required")
	}
	if NormalizeQuestArchetypeNodeType(string(q.NodeType)) == QuestArchetypeNodeTypeMonsterEncounter ||
		NormalizeQuestArchetypeNodeType(string(q.NodeType)) == QuestArchetypeNodeTypeScenario ||
		NormalizeQuestArchetypeNodeType(string(q.NodeType)) == QuestArchetypeNodeTypeExposition ||
		NormalizeQuestArchetypeNodeType(string(q.NodeType)) == QuestArchetypeNodeTypeFetchQuest ||
		NormalizeQuestArchetypeNodeType(string(q.NodeType)) == QuestArchetypeNodeTypeStoryFlag {
		return LocationArchetypeChallenge{}, fmt.Errorf("%s nodes do not generate location challenges", q.NodeType)
	}
	if q.LocationArchetype == nil {
		return LocationArchetypeChallenge{}, fmt.Errorf("location archetype is required")
	}
	return q.LocationArchetype.GetRandomChallengeByDifficulty(q.Difficulty)
}

func (q *QuestArchetypeNode) BeforeSave(tx *gorm.DB) error {
	q.NodeType = NormalizeQuestArchetypeNodeType(string(q.NodeType))
	q.ObjectiveDescription = strings.TrimSpace(q.ObjectiveDescription)
	q.StoryFlagKey = NormalizeStoryFlagKey(q.StoryFlagKey)
	q.LocationSelectionMode = NormalizeQuestArchetypeNodeLocationSelectionMode(
		string(q.LocationSelectionMode),
	)
	q.EncounterRewardMode = NormalizeRewardMode(string(q.EncounterRewardMode))
	q.EncounterRandomRewardSize = NormalizeRandomRewardSize(string(q.EncounterRandomRewardSize))
	q.ExpositionRewardMode = NormalizeRewardMode(string(q.ExpositionRewardMode))
	q.ExpositionRandomRewardSize = NormalizeRandomRewardSize(string(q.ExpositionRandomRewardSize))
	if q.TargetLevel < 1 {
		q.TargetLevel = 1
	}
	if q.EncounterProximityMeters < 0 {
		q.EncounterProximityMeters = 0
	}
	if q.ExpositionDialogue == nil {
		q.ExpositionDialogue = DialogueSequence{}
	}
	q.FetchRequirements = NormalizeFetchQuestRequirements(q.FetchRequirements)
	if q.ExpositionMaterialRewards == nil {
		q.ExpositionMaterialRewards = BaseMaterialRewards{}
	}
	if q.ExpositionItemRewards == nil {
		q.ExpositionItemRewards = QuestArchetypeExpositionItemRewards{}
	}
	if q.ExpositionSpellRewards == nil {
		q.ExpositionSpellRewards = QuestArchetypeExpositionSpellRewards{}
	}
	if q.ExpositionRewardExperience < 0 {
		q.ExpositionRewardExperience = 0
	}
	if q.ExpositionRewardGold < 0 {
		q.ExpositionRewardGold = 0
	}
	return nil
}
