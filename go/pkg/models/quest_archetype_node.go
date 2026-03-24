package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestArchetypeNodeType string

const (
	QuestArchetypeNodeTypeLocation         QuestArchetypeNodeType = "location"
	QuestArchetypeNodeTypeMonsterEncounter QuestArchetypeNodeType = "monster_encounter"
	QuestArchetypeNodeTypeScenario         QuestArchetypeNodeType = "scenario"
)

func NormalizeQuestArchetypeNodeType(raw string) QuestArchetypeNodeType {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(QuestArchetypeNodeTypeScenario):
		return QuestArchetypeNodeTypeScenario
	case string(QuestArchetypeNodeTypeMonsterEncounter):
		return QuestArchetypeNodeTypeMonsterEncounter
	default:
		return QuestArchetypeNodeTypeLocation
	}
}

type QuestArchetypeNode struct {
	ID                        uuid.UUID                   `json:"id"`
	CreatedAt                 time.Time                   `json:"createdAt"`
	UpdatedAt                 time.Time                   `json:"updatedAt"`
	DeletedAt                 gorm.DeletedAt              `json:"deletedAt"`
	NodeType                  QuestArchetypeNodeType      `json:"nodeType" gorm:"column:node_type"`
	LocationArchetype         *LocationArchetype          `json:"locationArchetype,omitempty"`
	LocationArchetypeID       *uuid.UUID                  `json:"locationArchetypeId,omitempty"`
	ScenarioTemplate          *ScenarioTemplate           `json:"scenarioTemplate,omitempty"`
	ScenarioTemplateID        *uuid.UUID                  `json:"scenarioTemplateId,omitempty"`
	MonsterTemplateIDs        StringArray                 `json:"monsterTemplateIds" gorm:"column:monster_template_ids;type:jsonb"`
	TargetLevel               int                         `json:"targetLevel" gorm:"column:target_level;default:1"`
	EncounterRewardMode       RewardMode                  `json:"encounterRewardMode" gorm:"column:encounter_reward_mode"`
	EncounterRandomRewardSize RandomRewardSize            `json:"encounterRandomRewardSize" gorm:"column:encounter_random_reward_size"`
	EncounterRewardExperience int                         `json:"encounterRewardExperience" gorm:"column:encounter_reward_experience"`
	EncounterRewardGold       int                         `json:"encounterRewardGold" gorm:"column:encounter_reward_gold"`
	EncounterMaterialRewards  BaseMaterialRewards         `json:"encounterMaterialRewards" gorm:"column:encounter_material_rewards_json;type:jsonb;default:'[]'"`
	EncounterItemRewards      MonsterEncounterRewardItems `json:"encounterItemRewards" gorm:"column:encounter_item_rewards_json;type:jsonb;default:'[]'"`
	EncounterProximityMeters  int                         `json:"encounterProximityMeters" gorm:"column:encounter_proximity_meters;default:100"`
	Challenges                []QuestArchetypeChallenge   `json:"challenges" gorm:"many2many:quest_archetype_node_challenges;"`
	Difficulty                int                         `json:"difficulty" gorm:"default:0"`
}

func (q *QuestArchetypeNode) GetRandomChallenge() (LocationArchetypeChallenge, error) {
	if q == nil {
		return LocationArchetypeChallenge{}, fmt.Errorf("quest archetype node is required")
	}
	if q.NodeType == QuestArchetypeNodeTypeMonsterEncounter || q.NodeType == QuestArchetypeNodeTypeScenario {
		return LocationArchetypeChallenge{}, fmt.Errorf("%s nodes do not generate location challenges", q.NodeType)
	}
	if q.LocationArchetype == nil {
		return LocationArchetypeChallenge{}, fmt.Errorf("location archetype is required")
	}
	return q.LocationArchetype.GetRandomChallengeByDifficulty(q.Difficulty)
}

func (q *QuestArchetypeNode) BeforeSave(tx *gorm.DB) error {
	q.NodeType = NormalizeQuestArchetypeNodeType(string(q.NodeType))
	q.EncounterRewardMode = NormalizeRewardMode(string(q.EncounterRewardMode))
	q.EncounterRandomRewardSize = NormalizeRandomRewardSize(string(q.EncounterRandomRewardSize))
	if q.TargetLevel < 1 {
		q.TargetLevel = 1
	}
	if q.EncounterProximityMeters < 0 {
		q.EncounterProximityMeters = 0
	}
	return nil
}
