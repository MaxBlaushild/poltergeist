package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type monsterTemplateUpsertRequest struct {
	Archived                      *bool    `json:"archived"`
	MonsterType                   string   `json:"monsterType"`
	GenreID                       string   `json:"genreId"`
	Name                          string   `json:"name"`
	Description                   string   `json:"description"`
	ImageURL                      string   `json:"imageUrl"`
	ThumbnailURL                  string   `json:"thumbnailUrl"`
	BaseStrength                  int      `json:"baseStrength"`
	BaseDexterity                 int      `json:"baseDexterity"`
	BaseConstitution              int      `json:"baseConstitution"`
	BaseIntelligence              int      `json:"baseIntelligence"`
	BaseWisdom                    int      `json:"baseWisdom"`
	BaseCharisma                  int      `json:"baseCharisma"`
	PhysicalDamageBonusPercent    int      `json:"physicalDamageBonusPercent"`
	PiercingDamageBonusPercent    int      `json:"piercingDamageBonusPercent"`
	SlashingDamageBonusPercent    int      `json:"slashingDamageBonusPercent"`
	BludgeoningDamageBonusPercent int      `json:"bludgeoningDamageBonusPercent"`
	FireDamageBonusPercent        int      `json:"fireDamageBonusPercent"`
	IceDamageBonusPercent         int      `json:"iceDamageBonusPercent"`
	LightningDamageBonusPercent   int      `json:"lightningDamageBonusPercent"`
	PoisonDamageBonusPercent      int      `json:"poisonDamageBonusPercent"`
	ArcaneDamageBonusPercent      int      `json:"arcaneDamageBonusPercent"`
	HolyDamageBonusPercent        int      `json:"holyDamageBonusPercent"`
	ShadowDamageBonusPercent      int      `json:"shadowDamageBonusPercent"`
	PhysicalResistancePercent     int      `json:"physicalResistancePercent"`
	PiercingResistancePercent     int      `json:"piercingResistancePercent"`
	SlashingResistancePercent     int      `json:"slashingResistancePercent"`
	BludgeoningResistancePercent  int      `json:"bludgeoningResistancePercent"`
	FireResistancePercent         int      `json:"fireResistancePercent"`
	IceResistancePercent          int      `json:"iceResistancePercent"`
	LightningResistancePercent    int      `json:"lightningResistancePercent"`
	PoisonResistancePercent       int      `json:"poisonResistancePercent"`
	ArcaneResistancePercent       int      `json:"arcaneResistancePercent"`
	HolyResistancePercent         int      `json:"holyResistancePercent"`
	ShadowResistancePercent       int      `json:"shadowResistancePercent"`
	SpellIDs                      []string `json:"spellIds"`
	ProgressionIDs                []string `json:"progressionIds"`
}

type bulkGenerateMonsterTemplatesRequest struct {
	Count       int    `json:"count"`
	MonsterType string `json:"monsterType"`
	GenreID     string `json:"genreId"`
}

type dndMonsterTemplateSeed struct {
	Name             string
	Description      string
	BaseStrength     int
	BaseDexterity    int
	BaseConstitution int
	BaseIntelligence int
	BaseWisdom       int
	BaseCharisma     int
}

type generatedMonsterTemplatePayload struct {
	Templates []jobs.MonsterTemplateCreationSpec `json:"templates"`
}

const generateMonsterTemplatesPromptTemplate = `
You are designing %d original %s monster templates for a fantasy action RPG.

Template role guidance:
%s

Avoid these existing monster template names:
%s

Hard constraints:
- Output exactly %d templates.
- Use unique names (2-4 words) that are NOT in the existing names list.
- Keep descriptions concise and practical (8-18 words), focused on monster behavior/combat role.
- Do not reference DnD, tabletop, or copyrighted franchises.
- All base stats must be integers from 1 to 20.
- Return JSON only.

Respond as:
{
  "templates": [
    {
      "name": "string",
      "description": "string",
      "baseStrength": 10,
      "baseDexterity": 10,
      "baseConstitution": 10,
      "baseIntelligence": 10,
      "baseWisdom": 10,
      "baseCharisma": 10
    }
  ]
}
`

const generateMonsterTemplatesWithGenrePromptTemplate = `
You are designing %d original %s monster templates for a %s action RPG.

Genre direction:
%s

Template role guidance:
%s

Avoid these existing monster template names:
%s

Hard constraints:
- Output exactly %d templates.
- Use unique names (2-4 words) that are NOT in the existing names list.
- Keep descriptions concise and practical (8-18 words), focused on monster behavior/combat role.
- Do not reference DnD, tabletop, or copyrighted franchises.
- All base stats must be integers from 1 to 20.
- Return JSON only.

Respond as:
{
  "templates": [
    {
      "name": "string",
      "description": "string",
      "baseStrength": 10,
      "baseDexterity": 10,
      "baseConstitution": 10,
      "baseIntelligence": 10,
      "baseWisdom": 10,
      "baseCharisma": 10
    }
  ]
}
`

var dndMonsterTemplateSeeds = []dndMonsterTemplateSeed{
	{
		Name:             "Goblin Skirmisher",
		Description:      "A nimble ambusher that favors hit-and-run tactics.",
		BaseStrength:     8,
		BaseDexterity:    14,
		BaseConstitution: 10,
		BaseIntelligence: 10,
		BaseWisdom:       8,
		BaseCharisma:     8,
	},
	{
		Name:             "Orc Berserker",
		Description:      "A brutal frontline raider built for relentless melee pressure.",
		BaseStrength:     16,
		BaseDexterity:    10,
		BaseConstitution: 14,
		BaseIntelligence: 8,
		BaseWisdom:       10,
		BaseCharisma:     10,
	},
	{
		Name:             "Kobold Trapmaster",
		Description:      "A crafty tunnel fighter that relies on tricks and terrain control.",
		BaseStrength:     7,
		BaseDexterity:    15,
		BaseConstitution: 10,
		BaseIntelligence: 11,
		BaseWisdom:       10,
		BaseCharisma:     8,
	},
	{
		Name:             "Bugbear Enforcer",
		Description:      "A heavy ambush predator that combines reach and sudden violence.",
		BaseStrength:     15,
		BaseDexterity:    12,
		BaseConstitution: 13,
		BaseIntelligence: 8,
		BaseWisdom:       11,
		BaseCharisma:     9,
	},
	{
		Name:             "Hobgoblin Captain",
		Description:      "A disciplined war leader who excels in organized combat.",
		BaseStrength:     14,
		BaseDexterity:    12,
		BaseConstitution: 13,
		BaseIntelligence: 12,
		BaseWisdom:       11,
		BaseCharisma:     12,
	},
	{
		Name:             "Gnoll Fang",
		Description:      "A savage pack hunter driven by bloodlust and momentum.",
		BaseStrength:     14,
		BaseDexterity:    12,
		BaseConstitution: 12,
		BaseIntelligence: 8,
		BaseWisdom:       10,
		BaseCharisma:     8,
	},
	{
		Name:             "Skeleton Legionary",
		Description:      "An undead soldier, tireless and unnervingly precise.",
		BaseStrength:     10,
		BaseDexterity:    12,
		BaseConstitution: 12,
		BaseIntelligence: 6,
		BaseWisdom:       8,
		BaseCharisma:     5,
	},
	{
		Name:             "Zombie Brute",
		Description:      "A shambling terror, slow but difficult to put down.",
		BaseStrength:     14,
		BaseDexterity:    6,
		BaseConstitution: 16,
		BaseIntelligence: 3,
		BaseWisdom:       6,
		BaseCharisma:     5,
	},
	{
		Name:             "Owlbear Ravager",
		Description:      "A feral apex beast that overwhelms prey with sheer force.",
		BaseStrength:     18,
		BaseDexterity:    12,
		BaseConstitution: 16,
		BaseIntelligence: 3,
		BaseWisdom:       12,
		BaseCharisma:     7,
	},
	{
		Name:             "Displacer Beast Stalker",
		Description:      "A predatory illusion-weaver that is hard to pin down.",
		BaseStrength:     13,
		BaseDexterity:    15,
		BaseConstitution: 13,
		BaseIntelligence: 6,
		BaseWisdom:       12,
		BaseCharisma:     8,
	},
	{
		Name:             "Mimic Lurker",
		Description:      "A shape-shifting ambusher that hides in plain sight.",
		BaseStrength:     15,
		BaseDexterity:    12,
		BaseConstitution: 14,
		BaseIntelligence: 5,
		BaseWisdom:       10,
		BaseCharisma:     8,
	},
	{
		Name:             "Gelatinous Cube",
		Description:      "A dungeon ooze, corrosive and inexorable.",
		BaseStrength:     14,
		BaseDexterity:    4,
		BaseConstitution: 16,
		BaseIntelligence: 1,
		BaseWisdom:       6,
		BaseCharisma:     1,
	},
	{
		Name:             "Mind Flayer Arcanist",
		Description:      "A psionic manipulator, dangerous at range and in control.",
		BaseStrength:     11,
		BaseDexterity:    12,
		BaseConstitution: 12,
		BaseIntelligence: 17,
		BaseWisdom:       15,
		BaseCharisma:     16,
	},
	{
		Name:             "Beholder Tyrant",
		Description:      "An aberrant overseer that dominates space with magical pressure.",
		BaseStrength:     10,
		BaseDexterity:    14,
		BaseConstitution: 16,
		BaseIntelligence: 17,
		BaseWisdom:       15,
		BaseCharisma:     17,
	},
	{
		Name:             "Young Red Dragon",
		Description:      "A proud draconic terror that blends mobility and overwhelming damage.",
		BaseStrength:     19,
		BaseDexterity:    12,
		BaseConstitution: 17,
		BaseIntelligence: 14,
		BaseWisdom:       11,
		BaseCharisma:     15,
	},
}

var genericMonsterTemplateRoleSeeds = []dndMonsterTemplateSeed{
	{
		Name:             "Skirmisher",
		Description:      "A nimble attacker built for burst pressure and quick repositioning.",
		BaseStrength:     10,
		BaseDexterity:    15,
		BaseConstitution: 11,
		BaseIntelligence: 10,
		BaseWisdom:       9,
		BaseCharisma:     8,
	},
	{
		Name:             "Bruiser",
		Description:      "A heavy frontline threat that wins fights through toughness and force.",
		BaseStrength:     16,
		BaseDexterity:    9,
		BaseConstitution: 15,
		BaseIntelligence: 8,
		BaseWisdom:       10,
		BaseCharisma:     9,
	},
	{
		Name:             "Controller",
		Description:      "A tactical menace that disrupts movement and dictates battlefield tempo.",
		BaseStrength:     8,
		BaseDexterity:    11,
		BaseConstitution: 11,
		BaseIntelligence: 16,
		BaseWisdom:       14,
		BaseCharisma:     10,
	},
	{
		Name:             "Stalker",
		Description:      "A patient hunter that excels at ambushes, pursuit, and isolated kills.",
		BaseStrength:     12,
		BaseDexterity:    15,
		BaseConstitution: 11,
		BaseIntelligence: 9,
		BaseWisdom:       13,
		BaseCharisma:     8,
	},
	{
		Name:             "Mystic",
		Description:      "A supernatural combatant that pairs ranged pressure with strange utility.",
		BaseStrength:     8,
		BaseDexterity:    10,
		BaseConstitution: 11,
		BaseIntelligence: 16,
		BaseWisdom:       14,
		BaseCharisma:     13,
	},
	{
		Name:             "Sentinel",
		Description:      "A disciplined defender that holds ground and punishes reckless advances.",
		BaseStrength:     14,
		BaseDexterity:    10,
		BaseConstitution: 15,
		BaseIntelligence: 10,
		BaseWisdom:       12,
		BaseCharisma:     10,
	},
}

type monsterRewardItemPayload struct {
	InventoryItemID int                   `json:"inventoryItemId"`
	Quantity        int                   `json:"quantity"`
	InventoryItem   *models.InventoryItem `json:"inventoryItem,omitempty"`
}

type monsterUpsertRequest struct {
	Name                        string                      `json:"name"`
	Description                 string                      `json:"description"`
	ImageURL                    string                      `json:"imageUrl"`
	ThumbnailURL                string                      `json:"thumbnailUrl"`
	ZoneID                      string                      `json:"zoneId"`
	Latitude                    float64                     `json:"latitude"`
	Longitude                   float64                     `json:"longitude"`
	TemplateID                  string                      `json:"templateId"`
	DominantHandInventoryItemID *int                        `json:"dominantHandInventoryItemId"`
	OffHandInventoryItemID      *int                        `json:"offHandInventoryItemId"`
	WeaponInventoryItemID       *int                        `json:"weaponInventoryItemId"`
	Level                       int                         `json:"level"`
	RewardMode                  string                      `json:"rewardMode"`
	RandomRewardSize            string                      `json:"randomRewardSize"`
	RewardExperience            int                         `json:"rewardExperience"`
	RewardGold                  int                         `json:"rewardGold"`
	MaterialRewards             []baseMaterialRewardPayload `json:"materialRewards"`
	ItemRewards                 []monsterRewardItemPayload  `json:"itemRewards"`
}

type monsterEncounterUpsertRequest struct {
	Name                string                      `json:"name"`
	Description         string                      `json:"description"`
	ImageURL            string                      `json:"imageUrl"`
	ThumbnailURL        string                      `json:"thumbnailUrl"`
	EncounterType       string                      `json:"encounterType"`
	RewardMode          string                      `json:"rewardMode"`
	RandomRewardSize    string                      `json:"randomRewardSize"`
	RewardExperience    int                         `json:"rewardExperience"`
	RewardGold          int                         `json:"rewardGold"`
	MaterialRewards     []baseMaterialRewardPayload `json:"materialRewards"`
	ItemRewards         []monsterRewardItemPayload  `json:"itemRewards"`
	ScaleWithUserLevel  bool                        `json:"scaleWithUserLevel"`
	RecurrenceFrequency *string                     `json:"recurrenceFrequency"`
	ZoneID              string                      `json:"zoneId"`
	PointOfInterestID   string                      `json:"pointOfInterestId"`
	Latitude            float64                     `json:"latitude"`
	Longitude           float64                     `json:"longitude"`
	MonsterIDs          []string                    `json:"monsterIds"`
}

type monsterBattleActionRequest struct {
	ActionType        string  `json:"actionType"`
	AbilityID         *string `json:"abilityId"`
	AbilityName       string  `json:"abilityName"`
	AbilityType       string  `json:"abilityType"`
	TargetsAllEnemies bool    `json:"targetsAllEnemies"`
	Heal              int     `json:"heal"`
}

type monsterTemplateResponse struct {
	ID                    uuid.UUID                            `json:"id"`
	CreatedAt             time.Time                            `json:"createdAt"`
	UpdatedAt             time.Time                            `json:"updatedAt"`
	Archived              bool                                 `json:"archived"`
	MonsterType           string                               `json:"monsterType"`
	GenreID               uuid.UUID                            `json:"genreId"`
	Genre                 *models.ZoneGenre                    `json:"genre,omitempty"`
	Name                  string                               `json:"name"`
	Description           string                               `json:"description"`
	ImageURL              string                               `json:"imageUrl"`
	ThumbnailURL          string                               `json:"thumbnailUrl"`
	BaseStrength          int                                  `json:"baseStrength"`
	BaseDexterity         int                                  `json:"baseDexterity"`
	BaseConstitution      int                                  `json:"baseConstitution"`
	BaseIntelligence      int                                  `json:"baseIntelligence"`
	BaseWisdom            int                                  `json:"baseWisdom"`
	BaseCharisma          int                                  `json:"baseCharisma"`
	AffinityDamageBonuses map[string]int                       `json:"affinityDamageBonuses"`
	AffinityResistances   map[string]int                       `json:"affinityResistances"`
	Progressions          []monsterTemplateProgressionResponse `json:"progressions"`
	Spells                []models.Spell                       `json:"spells"`
	ImageGenerationStatus string                               `json:"imageGenerationStatus"`
	ImageGenerationError  *string                              `json:"imageGenerationError,omitempty"`
}

type monsterTemplateProgressionResponse struct {
	ID          uuid.UUID               `json:"id"`
	Name        string                  `json:"name"`
	AbilityType models.SpellAbilityType `json:"abilityType"`
}

type monsterResponse struct {
	ID                          uuid.UUID                  `json:"id"`
	CreatedAt                   time.Time                  `json:"createdAt"`
	UpdatedAt                   time.Time                  `json:"updatedAt"`
	Name                        string                     `json:"name"`
	Description                 string                     `json:"description"`
	ImageURL                    string                     `json:"imageUrl"`
	ThumbnailURL                string                     `json:"thumbnailUrl"`
	ZoneID                      uuid.UUID                  `json:"zoneId"`
	Zone                        models.Zone                `json:"zone"`
	GenreID                     uuid.UUID                  `json:"genreId"`
	Genre                       *models.ZoneGenre          `json:"genre,omitempty"`
	Latitude                    float64                    `json:"latitude"`
	Longitude                   float64                    `json:"longitude"`
	TemplateID                  *uuid.UUID                 `json:"templateId,omitempty"`
	Template                    *monsterTemplateResponse   `json:"template,omitempty"`
	DominantHandInventoryItemID *int                       `json:"dominantHandInventoryItemId,omitempty"`
	DominantHandInventoryItem   *models.InventoryItem      `json:"dominantHandInventoryItem,omitempty"`
	OffHandInventoryItemID      *int                       `json:"offHandInventoryItemId,omitempty"`
	OffHandInventoryItem        *models.InventoryItem      `json:"offHandInventoryItem,omitempty"`
	WeaponInventoryItemID       *int                       `json:"weaponInventoryItemId,omitempty"`
	WeaponInventoryItem         *models.InventoryItem      `json:"weaponInventoryItem,omitempty"`
	Level                       int                        `json:"level"`
	Strength                    int                        `json:"strength"`
	Dexterity                   int                        `json:"dexterity"`
	Constitution                int                        `json:"constitution"`
	Intelligence                int                        `json:"intelligence"`
	Wisdom                      int                        `json:"wisdom"`
	Charisma                    int                        `json:"charisma"`
	Health                      int                        `json:"health"`
	MaxHealth                   int                        `json:"maxHealth"`
	Mana                        int                        `json:"mana"`
	MaxMana                     int                        `json:"maxMana"`
	AttackDamageMin             int                        `json:"attackDamageMin"`
	AttackDamageMax             int                        `json:"attackDamageMax"`
	AttackSwipesPerAttack       int                        `json:"attackSwipesPerAttack"`
	AffinityDamageBonuses       map[string]int             `json:"affinityDamageBonuses"`
	AffinityResistances         map[string]int             `json:"affinityResistances"`
	Spells                      []models.Spell             `json:"spells"`
	Statuses                    []models.MonsterStatus     `json:"statuses"`
	ActiveBattleID              *uuid.UUID                 `json:"activeBattleId,omitempty"`
	RewardMode                  models.RewardMode          `json:"rewardMode"`
	RandomRewardSize            models.RandomRewardSize    `json:"randomRewardSize"`
	RewardExperience            int                        `json:"rewardExperience"`
	RewardGold                  int                        `json:"rewardGold"`
	MaterialRewards             models.BaseMaterialRewards `json:"materialRewards"`
	ItemRewards                 []models.MonsterItemReward `json:"itemRewards"`
	ImageGenerationStatus       string                     `json:"imageGenerationStatus"`
	ImageGenerationError        *string                    `json:"imageGenerationError,omitempty"`
}

type monsterBattleResponse struct {
	ID                   uuid.UUID                      `json:"id"`
	UserID               uuid.UUID                      `json:"userId"`
	MonsterID            uuid.UUID                      `json:"monsterId"`
	MonsterEncounterID   *uuid.UUID                     `json:"monsterEncounterId,omitempty"`
	State                string                         `json:"state"`
	TurnIndex            int                            `json:"turnIndex"`
	StartedAt            time.Time                      `json:"startedAt"`
	LastActivityAt       time.Time                      `json:"lastActivityAt"`
	MonsterHealthDeficit int                            `json:"monsterHealthDeficit"`
	MonsterManaDeficit   int                            `json:"monsterManaDeficit"`
	LastActionSequence   int                            `json:"lastActionSequence"`
	LastAction           models.MonsterBattleLastAction `json:"lastAction"`
	EndedAt              *time.Time                     `json:"endedAt,omitempty"`
}

type monsterEncounterMemberResponse struct {
	Slot    int             `json:"slot"`
	Monster monsterResponse `json:"monster"`
}

type monsterEncounterResponse struct {
	ID                          uuid.UUID                        `json:"id"`
	CreatedAt                   time.Time                        `json:"createdAt"`
	UpdatedAt                   time.Time                        `json:"updatedAt"`
	Name                        string                           `json:"name"`
	Description                 string                           `json:"description"`
	ImageURL                    string                           `json:"imageUrl"`
	ThumbnailURL                string                           `json:"thumbnailUrl"`
	EncounterType               models.MonsterEncounterType      `json:"encounterType"`
	RewardMode                  models.RewardMode                `json:"rewardMode"`
	RandomRewardSize            models.RandomRewardSize          `json:"randomRewardSize"`
	RewardExperience            int                              `json:"rewardExperience"`
	RewardGold                  int                              `json:"rewardGold"`
	MaterialRewards             models.BaseMaterialRewards       `json:"materialRewards"`
	ItemRewards                 []monsterRewardItemPayload       `json:"itemRewards"`
	ScaleWithUserLevel          bool                             `json:"scaleWithUserLevel"`
	RecurringMonsterEncounterID *uuid.UUID                       `json:"recurringMonsterEncounterId,omitempty"`
	RecurrenceFrequency         *string                          `json:"recurrenceFrequency,omitempty"`
	NextRecurrenceAt            *time.Time                       `json:"nextRecurrenceAt,omitempty"`
	ZoneID                      uuid.UUID                        `json:"zoneId"`
	Zone                        models.Zone                      `json:"zone"`
	PointOfInterestID           *uuid.UUID                       `json:"pointOfInterestId,omitempty"`
	Latitude                    float64                          `json:"latitude"`
	Longitude                   float64                          `json:"longitude"`
	MonsterCount                int                              `json:"monsterCount"`
	Members                     []monsterEncounterMemberResponse `json:"members"`
	Monsters                    []monsterResponse                `json:"monsters"`
}

type paginatedMonsterTemplateResponse struct {
	Items         []monsterTemplateResponse `json:"items"`
	Total         int64                     `json:"total"`
	Page          int                       `json:"page"`
	PageSize      int                       `json:"pageSize"`
	ActiveCount   int64                     `json:"activeCount"`
	ArchivedCount int64                     `json:"archivedCount"`
}

type paginatedMonsterResponse struct {
	Items    []monsterResponse `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

type paginatedMonsterEncounterResponse struct {
	Items    []monsterEncounterResponse `json:"items"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"pageSize"`
}

const (
	adminMonsterListDefaultPageSize = 25
	adminMonsterListMaxPageSize     = 100
)

func clampAdminMonsterListPage(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

func clampAdminMonsterListPageSize(value int) int {
	if value <= 0 {
		return adminMonsterListDefaultPageSize
	}
	if value > adminMonsterListMaxPageSize {
		return adminMonsterListMaxPageSize
	}
	return value
}

func parseAdminMonsterListPage(ctx *gin.Context) int {
	page, err := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page", "1")))
	if err != nil {
		return 1
	}
	return clampAdminMonsterListPage(page)
}

func parseAdminMonsterListPageSize(ctx *gin.Context) int {
	pageSize, err := strconv.Atoi(strings.TrimSpace(
		ctx.DefaultQuery("pageSize", strconv.Itoa(adminMonsterListDefaultPageSize)),
	))
	if err != nil {
		return adminMonsterListDefaultPageSize
	}
	return clampAdminMonsterListPageSize(pageSize)
}

func parseOptionalArchivedFilter(raw string) (*bool, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "":
		return nil, nil
	case "true":
		value := true
		return &value, nil
	case "false":
		value := false
		return &value, nil
	default:
		return nil, errors.New("archived must be true or false")
	}
}

func parseOptionalGenreIDFilter(raw string) (*uuid.UUID, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", "all":
		return nil, nil
	}

	genreID, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || genreID == uuid.Nil {
		return nil, errors.New("genreId must be a valid UUID")
	}
	return &genreID, nil
}

func monsterBattleResponseFrom(battle *models.MonsterBattle) *monsterBattleResponse {
	if battle == nil {
		return nil
	}
	return &monsterBattleResponse{
		ID:                   battle.ID,
		UserID:               battle.UserID,
		MonsterID:            battle.MonsterID,
		MonsterEncounterID:   battle.MonsterEncounterID,
		State:                battle.State,
		TurnIndex:            battle.TurnIndex,
		StartedAt:            battle.StartedAt,
		LastActivityAt:       battle.LastActivityAt,
		MonsterHealthDeficit: battle.MonsterHealthDeficit,
		MonsterManaDeficit:   battle.MonsterManaDeficit,
		LastActionSequence:   battle.LastActionSequence,
		LastAction:           battle.LastAction,
		EndedAt:              battle.EndedAt,
	}
}

func parseMonsterBattleActionAbilityID(raw *string) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid abilityId")
	}
	return &parsed, nil
}

func monsterBattleLastActionFromRequest(
	user *models.User,
	monster *models.Monster,
	request monsterBattleActionRequest,
	abilityID *uuid.UUID,
	appliedDamage int,
) models.MonsterBattleLastAction {
	targetName := "the enemy"
	var targetMonsterID *uuid.UUID
	if monster != nil {
		targetName = strings.TrimSpace(monster.Name)
		if targetName == "" {
			targetName = "the enemy"
		}
		targetMonsterID = &monster.ID
	}
	return models.MonsterBattleLastAction{
		ActionType:        normalizeMonsterBattleActionType(request.ActionType),
		ActorType:         "user",
		ActorUserID:       &user.ID,
		ActorName:         monsterBattleUserDisplayName(user),
		AbilityID:         abilityID,
		AbilityName:       strings.TrimSpace(request.AbilityName),
		AbilityType:       strings.TrimSpace(request.AbilityType),
		TargetMonsterID:   targetMonsterID,
		TargetName:        targetName,
		TargetsAllEnemies: request.TargetsAllEnemies,
		Damage:            max(0, appliedDamage),
		Heal:              max(0, request.Heal),
	}
}

func (s *server) buildMonsterEncounterRewardPayloads(
	ctx context.Context,
	rewards []models.MonsterEncounterRewardItem,
) ([]monsterRewardItemPayload, []scenarioRewardItem, error) {
	itemPayloads := make([]monsterRewardItemPayload, 0, len(rewards))
	rewardItems := make([]scenarioRewardItem, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, reward.InventoryItemID)
		if err != nil {
			return nil, nil, err
		}
		itemPayloads = append(itemPayloads, monsterRewardItemPayload{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
			InventoryItem:   item,
		})
		rewardItems = append(rewardItems, scenarioRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return itemPayloads, rewardItems, nil
}

func monsterEncounterRewardItemsToScenarioRewards(
	rewards []models.MonsterEncounterRewardItem,
) []scenarioRewardItem {
	out := make([]scenarioRewardItem, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		out = append(out, scenarioRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func (s *server) resolveMonsterEncounterRewardsForUser(
	ctx context.Context,
	userID uuid.UUID,
	encounter *models.MonsterEncounter,
) (models.RewardMode, models.RandomRewardSize, int, int, []monsterRewardItemPayload, []scenarioRewardItem, error) {
	if encounter == nil {
		return models.RewardModeExplicit, models.RandomRewardSizeSmall, 0, 0, nil, nil, nil
	}
	rewardMode := models.NormalizeRewardMode(string(encounter.RewardMode))
	rewardSize := models.NormalizeRandomRewardSize(string(encounter.RandomRewardSize))
	if rewardMode == models.RewardModeRandom {
		rewardSeed := fmt.Sprintf("monster-encounter:%s:user:%s", encounter.ID, userID)
		plan, itemByID, userLevel, err := s.randomRewardPlanForUser(
			ctx,
			userID,
			rewardSize,
			rewardSeed,
		)
		if err != nil {
			return rewardMode, rewardSize, 0, 0, nil, nil, err
		}
		plan = ensureMonsterEncounterRandomRewardHasItem(plan, itemByID, userLevel, rewardSeed)
		guaranteedItems := monsterEncounterRewardItemsToScenarioRewards(encounter.ItemRewards)
		for _, reward := range encounter.ItemRewards {
			if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
				continue
			}
			if _, ok := itemByID[reward.InventoryItemID]; ok {
				continue
			}
			item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, reward.InventoryItemID)
			if err != nil {
				return rewardMode, rewardSize, 0, 0, nil, nil, err
			}
			if item != nil {
				itemByID[reward.InventoryItemID] = *item
			}
		}
		itemGrants := models.MergeRandomRewardItemGrants(
			plan.ItemGrants,
			scenarioRewardItemsToRandomRewardItemGrants(guaranteedItems),
		)
		payloads := make([]monsterRewardItemPayload, 0, len(itemGrants))
		rewardItems := make([]scenarioRewardItem, 0, len(itemGrants))
		for _, grant := range itemGrants {
			if grant.InventoryItemID <= 0 || grant.Quantity <= 0 {
				continue
			}
			payload := monsterRewardItemPayload{
				InventoryItemID: grant.InventoryItemID,
				Quantity:        grant.Quantity,
			}
			if item, ok := itemByID[grant.InventoryItemID]; ok {
				itemCopy := item
				payload.InventoryItem = &itemCopy
			}
			payloads = append(payloads, payload)
			rewardItems = append(rewardItems, scenarioRewardItem{
				InventoryItemID: grant.InventoryItemID,
				Quantity:        grant.Quantity,
			})
		}
		return rewardMode, rewardSize, plan.Experience, plan.Gold, payloads, rewardItems, nil
	}

	itemPayloads, rewardItems, err := s.buildMonsterEncounterRewardPayloads(ctx, encounter.ItemRewards)
	if err != nil {
		return rewardMode, rewardSize, 0, 0, nil, nil, err
	}
	rewardExperience := max(0, encounter.RewardExperience)
	rewardGold := max(0, encounter.RewardGold)
	return rewardMode, rewardSize, rewardExperience, rewardGold, itemPayloads, rewardItems, nil
}

func (s *server) monsterEncounterResponseFrom(
	ctx context.Context,
	userID uuid.UUID,
	encounter *models.MonsterEncounter,
	userLevel int,
	applyLevelScaling bool,
) (monsterEncounterResponse, error) {
	rewardMode, rewardSize, rewardExperience, rewardGold, itemRewards, _, err := s.resolveMonsterEncounterRewardsForUser(
		ctx,
		userID,
		encounter,
	)
	if err != nil {
		return monsterEncounterResponse{}, err
	}
	members := make([]monsterEncounterMemberResponse, 0, len(encounter.Members))
	monsters := make([]monsterResponse, 0, len(encounter.Members))
	for i := range encounter.Members {
		member := encounter.Members[i]
		monster := member.Monster
		if applyLevelScaling && encounter.ScaleWithUserLevel {
			monster.Level = scaledEncounterMonsterLevelForUserLevelAndType(
				userLevel,
				len(encounter.Members),
				encounter.EncounterType,
			)
		}
		resolvedUserLevel := 0
		if applyLevelScaling {
			resolvedUserLevel = userLevel
		}
		entry, err := s.buildMonsterResponseForUserLevel(
			ctx,
			userID,
			&monster,
			resolvedUserLevel,
		)
		if err != nil {
			return monsterEncounterResponse{}, err
		}
		members = append(members, monsterEncounterMemberResponse{
			Slot:    member.Slot,
			Monster: entry,
		})
		monsters = append(monsters, entry)
	}

	imageURL := strings.TrimSpace(encounter.ImageURL)
	thumbnailURL := strings.TrimSpace(encounter.ThumbnailURL)
	if thumbnailURL == "" {
		thumbnailURL = imageURL
	}
	if imageURL == "" && len(monsters) > 0 {
		imageURL = strings.TrimSpace(monsters[0].ImageURL)
	}
	if thumbnailURL == "" && len(monsters) > 0 {
		thumbnailURL = strings.TrimSpace(monsters[0].ThumbnailURL)
	}

	return monsterEncounterResponse{
		ID:                          encounter.ID,
		CreatedAt:                   encounter.CreatedAt,
		UpdatedAt:                   encounter.UpdatedAt,
		Name:                        encounter.Name,
		Description:                 encounter.Description,
		ImageURL:                    imageURL,
		ThumbnailURL:                thumbnailURL,
		EncounterType:               models.NormalizeMonsterEncounterType(string(encounter.EncounterType)),
		RewardMode:                  rewardMode,
		RandomRewardSize:            rewardSize,
		RewardExperience:            rewardExperience,
		RewardGold:                  rewardGold,
		MaterialRewards:             encounter.MaterialRewards,
		ItemRewards:                 itemRewards,
		ScaleWithUserLevel:          encounter.ScaleWithUserLevel,
		RecurringMonsterEncounterID: encounter.RecurringMonsterEncounterID,
		RecurrenceFrequency:         encounter.RecurrenceFrequency,
		NextRecurrenceAt:            encounter.NextRecurrenceAt,
		ZoneID:                      encounter.ZoneID,
		Zone:                        encounter.Zone,
		PointOfInterestID:           encounter.PointOfInterestID,
		Latitude:                    encounter.Latitude,
		Longitude:                   encounter.Longitude,
		MonsterCount:                len(monsters),
		Members:                     members,
		Monsters:                    monsters,
	}, nil
}

func monsterTemplateResponseFrom(template *models.MonsterTemplate) *monsterTemplateResponse {
	if template == nil {
		return nil
	}
	affinityBonuses := template.AffinityBonuses()
	progressions := make([]monsterTemplateProgressionResponse, 0, len(template.Progressions))
	seenProgressionIDs := make(map[uuid.UUID]struct{}, len(template.Progressions))
	for _, link := range template.Progressions {
		progression := link.Progression
		if progression.ID == uuid.Nil {
			continue
		}
		if _, exists := seenProgressionIDs[progression.ID]; exists {
			continue
		}
		seenProgressionIDs[progression.ID] = struct{}{}
		progressions = append(progressions, monsterTemplateProgressionResponse{
			ID:          progression.ID,
			Name:        progression.Name,
			AbilityType: progression.AbilityType,
		})
	}
	spells := make([]models.Spell, 0, len(template.Spells)+len(template.Progressions))
	for _, spell := range monsterTemplateResolvedAbilitiesForLevel(template, 1) {
		spells = append(spells, spell)
	}
	if len(spells) == 0 {
		spells = make([]models.Spell, 0, len(template.Spells))
	}
	for _, templateSpell := range template.Spells {
		if templateSpell.Spell.ID == uuid.Nil {
			continue
		}
		if len(template.Progressions) == 0 {
			spells = append(spells, templateSpell.Spell)
		}
	}
	return &monsterTemplateResponse{
		ID:                    template.ID,
		CreatedAt:             template.CreatedAt,
		UpdatedAt:             template.UpdatedAt,
		Archived:              template.Archived,
		MonsterType:           string(models.NormalizeMonsterTemplateType(string(template.MonsterType))),
		GenreID:               template.GenreID,
		Genre:                 template.Genre,
		Name:                  template.Name,
		Description:           template.Description,
		ImageURL:              template.ImageURL,
		ThumbnailURL:          template.ThumbnailURL,
		BaseStrength:          template.BaseStrength,
		BaseDexterity:         template.BaseDexterity,
		BaseConstitution:      template.BaseConstitution,
		BaseIntelligence:      template.BaseIntelligence,
		BaseWisdom:            template.BaseWisdom,
		BaseCharisma:          template.BaseCharisma,
		AffinityDamageBonuses: affinityBonuses.AffinityDamageBonusMap(),
		AffinityResistances:   affinityBonuses.AffinityResistanceMap(),
		Progressions:          progressions,
		Spells:                spells,
		ImageGenerationStatus: template.ImageGenerationStatus,
		ImageGenerationError:  template.ImageGenerationError,
	}
}

func monsterTemplateResolvedAbilitiesForLevel(
	template *models.MonsterTemplate,
	level int,
) []models.Spell {
	if template == nil {
		return []models.Spell{}
	}
	if level < 1 {
		level = 1
	}
	estimatedSize := len(template.Progressions)
	if estimatedSize == 0 {
		estimatedSize = len(template.Spells)
	}
	abilities := make([]models.Spell, 0, estimatedSize)
	seen := map[uuid.UUID]struct{}{}
	appendAbility := func(spell models.Spell) {
		if spell.ID == uuid.Nil {
			return
		}
		if _, exists := seen[spell.ID]; exists {
			return
		}
		seen[spell.ID] = struct{}{}
		abilities = append(abilities, spell)
	}

	for _, link := range template.Progressions {
		best := monsterTemplateProgressionSpellAtOrBelowLevel(link.Progression, level)
		if best != nil {
			appendAbility(*best)
		}
	}

	if len(abilities) > 0 {
		return abilities
	}

	for _, templateSpell := range template.Spells {
		if !spellUnlockedForMonsterAbilityLevel(templateSpell.Spell, level) {
			continue
		}
		appendAbility(templateSpell.Spell)
	}
	return abilities
}

func spellUnlockedForMonsterAbilityLevel(spell models.Spell, level int) bool {
	if spell.ID == uuid.Nil {
		return false
	}
	if level < 1 {
		level = 1
	}
	unlockLevel := spell.AbilityLevel
	if unlockLevel <= 0 {
		unlockLevel = 1
	}
	return unlockLevel <= level
}

func monsterTemplateProgressionSpellAtOrBelowLevel(
	progression models.SpellProgression,
	level int,
) *models.Spell {
	if level < 1 {
		level = 1
	}
	bestIndex := -1
	bestLevel := 0
	for index, member := range progression.Members {
		if member.Spell.ID == uuid.Nil {
			continue
		}
		memberLevel := member.Spell.AbilityLevel
		if memberLevel <= 0 {
			memberLevel = member.LevelBand
		}
		if memberLevel <= 0 {
			memberLevel = 1
		}
		if memberLevel > level {
			continue
		}
		if bestIndex == -1 || memberLevel > bestLevel {
			bestIndex = index
			bestLevel = memberLevel
		}
	}
	if bestIndex == -1 {
		return nil
	}
	spell := progression.Members[bestIndex].Spell
	return &spell
}

func monsterResponseFrom(
	monster *models.Monster,
	statusBonuses models.CharacterStatBonuses,
	activeStatuses []models.MonsterStatus,
	activeBattle *models.MonsterBattle,
	userLevel int,
) monsterResponse {
	genreID := monster.GenreID
	genre := monster.Genre
	if genreID == uuid.Nil && monster.Template != nil {
		genreID = monster.Template.GenreID
	}
	if genre == nil && monster.Template != nil {
		genre = monster.Template.Genre
	}
	stats := monster.EffectiveStatsWithBonuses(statusBonuses)
	maxHealth := monster.DerivedMaxHealthWithBonuses(statusBonuses)
	maxMana := monster.DerivedMaxManaWithBonuses(statusBonuses)
	damageMin, damageMax, swipes := monster.DerivedAttackProfileWithBonuses(statusBonuses)
	currentHealth := maxHealth
	currentMana := maxMana
	if activeBattle != nil {
		currentHealth = maxHealth - activeBattle.MonsterHealthDeficit
		if currentHealth < 0 {
			currentHealth = 0
		}
		currentMana = maxMana - activeBattle.MonsterManaDeficit
		if currentMana < 0 {
			currentMana = 0
		}
	}
	resolvedAbilityLevel := cappedMonsterAbilityLevelForUserLevel(
		monster.EffectiveLevel(),
		userLevel,
	)
	spells := monsterTemplateResolvedAbilitiesForLevel(
		monster.Template,
		resolvedAbilityLevel,
	)
	imageURL := monster.ImageURL
	if imageURL == "" && monster.Template != nil {
		imageURL = monster.Template.ImageURL
	}
	thumbnailURL := monster.ThumbnailURL
	if thumbnailURL == "" && monster.Template != nil {
		thumbnailURL = monster.Template.ThumbnailURL
	}
	if thumbnailURL == "" {
		thumbnailURL = imageURL
	}
	totalAffinityBonuses := statusBonuses
	if monster.Template != nil {
		totalAffinityBonuses = monster.Template.AffinityBonuses().Add(totalAffinityBonuses)
	}

	dominantItemID := monster.DominantHandInventoryItemID
	dominantItem := monster.DominantHandInventoryItem
	if dominantItemID == nil {
		dominantItemID = monster.WeaponInventoryItemID
	}
	if dominantItem == nil {
		dominantItem = monster.WeaponInventoryItem
	}

	return monsterResponse{
		ID:                          monster.ID,
		CreatedAt:                   monster.CreatedAt,
		UpdatedAt:                   monster.UpdatedAt,
		Name:                        monster.Name,
		Description:                 monster.Description,
		ImageURL:                    imageURL,
		ThumbnailURL:                thumbnailURL,
		ZoneID:                      monster.ZoneID,
		Zone:                        monster.Zone,
		GenreID:                     genreID,
		Genre:                       genre,
		Latitude:                    monster.Latitude,
		Longitude:                   monster.Longitude,
		TemplateID:                  monster.TemplateID,
		Template:                    monsterTemplateResponseFrom(monster.Template),
		DominantHandInventoryItemID: dominantItemID,
		DominantHandInventoryItem:   dominantItem,
		OffHandInventoryItemID:      monster.OffHandInventoryItemID,
		OffHandInventoryItem:        monster.OffHandInventoryItem,
		WeaponInventoryItemID:       dominantItemID,
		WeaponInventoryItem:         dominantItem,
		Level:                       monster.EffectiveLevel(),
		Strength:                    stats.Strength,
		Dexterity:                   stats.Dexterity,
		Constitution:                stats.Constitution,
		Intelligence:                stats.Intelligence,
		Wisdom:                      stats.Wisdom,
		Charisma:                    stats.Charisma,
		Health:                      currentHealth,
		MaxHealth:                   maxHealth,
		Mana:                        currentMana,
		MaxMana:                     maxMana,
		AttackDamageMin:             damageMin,
		AttackDamageMax:             damageMax,
		AttackSwipesPerAttack:       swipes,
		AffinityDamageBonuses:       totalAffinityBonuses.AffinityDamageBonusMap(),
		AffinityResistances:         totalAffinityBonuses.AffinityResistanceMap(),
		Spells:                      spells,
		Statuses:                    activeStatuses,
		ActiveBattleID: func() *uuid.UUID {
			if activeBattle == nil {
				return nil
			}
			return &activeBattle.ID
		}(),
		RewardMode:            monster.RewardMode,
		RandomRewardSize:      monster.RandomRewardSize,
		RewardExperience:      monster.RewardExperience,
		RewardGold:            monster.RewardGold,
		MaterialRewards:       monster.MaterialRewards,
		ItemRewards:           monster.ItemRewards,
		ImageGenerationStatus: monster.ImageGenerationStatus,
		ImageGenerationError:  monster.ImageGenerationError,
	}
}

func (s *server) getOrCreateActiveMonsterBattle(
	ctx context.Context,
	userID uuid.UUID,
	monsterID uuid.UUID,
) (*models.MonsterBattle, error) {
	activeBattle, err := s.dbClient.MonsterBattle().FindActiveByUserAndMonster(ctx, userID, monsterID)
	if err != nil {
		return nil, err
	}
	if activeBattle != nil {
		log.Printf(
			"[party-combat][start] reusing active battle as owner user=%s monster=%s battle=%s state=%s",
			userID,
			monsterID,
			activeBattle.ID,
			activeBattle.State,
		)
		return activeBattle, nil
	}
	activeBattle, err = s.dbClient.MonsterBattle().FindActiveByParticipantAndMonster(ctx, userID, monsterID)
	if err != nil {
		return nil, err
	}
	if activeBattle != nil {
		log.Printf(
			"[party-combat][start] reusing active battle as participant user=%s monster=%s battle=%s state=%s",
			userID,
			monsterID,
			activeBattle.ID,
			activeBattle.State,
		)
		return activeBattle, nil
	}

	now := time.Now()
	battle := &models.MonsterBattle{
		UserID:         userID,
		MonsterID:      monsterID,
		State:          string(models.MonsterBattleStateActive),
		TurnIndex:      0,
		StartedAt:      now,
		LastActivityAt: now,
	}
	if err := s.dbClient.MonsterBattle().Create(ctx, battle); err != nil {
		return nil, err
	}
	log.Printf(
		"[party-combat][start] created new battle user=%s monster=%s battle=%s",
		userID,
		monsterID,
		battle.ID,
	)
	if err := s.initializeMonsterBattlePartyState(ctx, battle); err != nil {
		return nil, err
	}
	return battle, nil
}

func (s *server) createFreshMonsterBattle(
	ctx context.Context,
	userID uuid.UUID,
	monsterID uuid.UUID,
	monsterEncounterID *uuid.UUID,
) (*models.MonsterBattle, error) {
	now := time.Now()

	ownerBattle, err := s.dbClient.MonsterBattle().FindActiveByUserAndMonster(ctx, userID, monsterID)
	if err != nil {
		return nil, err
	}
	if ownerBattle != nil {
		if err := s.dbClient.MonsterBattle().End(ctx, ownerBattle.ID, now); err != nil {
			return nil, err
		}
		log.Printf(
			"[party-combat][start] ended previous owner battle user=%s monster=%s battle=%s",
			userID,
			monsterID,
			ownerBattle.ID,
		)
	}

	participantBattle, err := s.dbClient.MonsterBattle().FindActiveByParticipantAndMonster(ctx, userID, monsterID)
	if err != nil {
		return nil, err
	}
	if participantBattle != nil {
		log.Printf(
			"[party-combat][start] user has active participant battle; creating fresh owner battle user=%s monster=%s participantBattle=%s",
			userID,
			monsterID,
			participantBattle.ID,
		)
	}

	battle := &models.MonsterBattle{
		UserID:             userID,
		MonsterID:          monsterID,
		MonsterEncounterID: monsterEncounterID,
		State:              string(models.MonsterBattleStateActive),
		TurnIndex:          0,
		StartedAt:          now,
		LastActivityAt:     now,
	}
	if err := s.dbClient.MonsterBattle().Create(ctx, battle); err != nil {
		return nil, err
	}
	log.Printf(
		"[party-combat][start] created fresh battle user=%s monster=%s battle=%s",
		userID,
		monsterID,
		battle.ID,
	)
	if err := s.initializeMonsterBattlePartyState(ctx, battle); err != nil {
		return nil, err
	}
	return battle, nil
}

func (s *server) monsterBattleScalingLevel(
	ctx context.Context,
	battle *models.MonsterBattle,
) (int, error) {
	if battle == nil {
		return 1, nil
	}

	maxLevel := 1
	foundParticipant := false
	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		return 0, err
	}
	for _, participant := range participants {
		level, err := s.currentUserLevel(ctx, participant.UserID)
		if err != nil {
			return 0, err
		}
		if !foundParticipant || level > maxLevel {
			maxLevel = level
		}
		foundParticipant = true
	}
	if foundParticipant {
		return maxLevel, nil
	}
	if battle.UserID == uuid.Nil {
		return maxLevel, nil
	}
	level, err := s.currentUserLevel(ctx, battle.UserID)
	if err != nil {
		return 0, err
	}
	if level > maxLevel {
		maxLevel = level
	}
	return maxLevel, nil
}

func (s *server) monsterScaledForBattle(
	ctx context.Context,
	battle *models.MonsterBattle,
	monster *models.Monster,
) (*models.Monster, error) {
	if battle == nil || monster == nil {
		return monster, nil
	}
	var (
		encounter *models.MonsterEncounter
		err       error
	)
	if battle.MonsterEncounterID != nil && *battle.MonsterEncounterID != uuid.Nil {
		encounter, err = s.dbClient.MonsterEncounter().FindByID(ctx, *battle.MonsterEncounterID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if encounter == nil {
		encounter, err = s.dbClient.MonsterEncounter().FindFirstByMonsterID(ctx, monster.ID)
		if err != nil {
			return nil, err
		}
	}
	if encounter == nil || !encounter.ScaleWithUserLevel {
		return monster, nil
	}
	userLevel, err := s.monsterBattleScalingLevel(ctx, battle)
	if err != nil {
		return nil, err
	}
	scaledMonster := *monster
	scaledMonster.Level = scaledEncounterMonsterLevelForUserLevelAndType(
		userLevel,
		len(encounter.Members),
		encounter.EncounterType,
	)
	return &scaledMonster, nil
}

func (s *server) buildMonsterResponse(
	ctx context.Context,
	userID uuid.UUID,
	monster *models.Monster,
) (monsterResponse, error) {
	return s.buildMonsterResponseForUserLevel(ctx, userID, monster, 0)
}

func (s *server) buildMonsterResponseForUserLevel(
	ctx context.Context,
	userID uuid.UUID,
	monster *models.Monster,
	userLevel int,
) (monsterResponse, error) {
	activeBattle, err := s.findActiveMonsterBattleForUser(ctx, userID, monster.ID)
	if err != nil {
		return monsterResponse{}, err
	}
	if activeBattle == nil {
		response := monsterResponseFrom(
			monster,
			models.CharacterStatBonuses{},
			[]models.MonsterStatus{},
			nil,
			userLevel,
		)
		if err := s.applyMonsterRewardsForUser(ctx, userID, monster, &response); err != nil {
			return monsterResponse{}, err
		}
		return response, nil
	}

	activeStatuses, err := s.dbClient.MonsterStatus().FindActiveByBattleID(ctx, activeBattle.ID)
	if err != nil {
		return monsterResponse{}, err
	}
	totalStatusBonuses := models.CharacterStatBonuses{}
	for _, status := range activeStatuses {
		totalStatusBonuses = totalStatusBonuses.Add(status.StatModifiers())
	}
	battleMonster, err := s.monsterScaledForBattle(ctx, activeBattle, monster)
	if err != nil {
		return monsterResponse{}, err
	}
	resolvedUserLevel := userLevel
	if resolvedUserLevel <= 0 {
		resolvedUserLevel, err = s.monsterBattleScalingLevel(ctx, activeBattle)
		if err != nil {
			return monsterResponse{}, err
		}
	}
	response := monsterResponseFrom(
		battleMonster,
		totalStatusBonuses,
		activeStatuses,
		activeBattle,
		resolvedUserLevel,
	)
	if err := s.applyMonsterRewardsForUser(ctx, userID, monster, &response); err != nil {
		return monsterResponse{}, err
	}
	return response, nil
}

func (s *server) parseMonsterTemplateUpsertRequest(
	ctx context.Context,
	body monsterTemplateUpsertRequest,
) (*models.MonsterTemplate, []models.MonsterTemplateProgression, []models.MonsterTemplateSpell, error) {
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return nil, nil, nil, fmt.Errorf("name is required")
	}
	if body.BaseStrength < 1 ||
		body.BaseDexterity < 1 ||
		body.BaseConstitution < 1 ||
		body.BaseIntelligence < 1 ||
		body.BaseWisdom < 1 ||
		body.BaseCharisma < 1 {
		return nil, nil, nil, fmt.Errorf("all base stats must be positive")
	}
	monsterType := models.NormalizeMonsterTemplateType(body.MonsterType)
	genre, err := s.resolveMonsterGenre(ctx, body.GenreID)
	if err != nil {
		return nil, nil, nil, err
	}

	progressions := []models.MonsterTemplateProgression{}
	seenProgressionIDs := map[uuid.UUID]bool{}
	for index, progressionIDString := range body.ProgressionIDs {
		progressionID, err := uuid.Parse(strings.TrimSpace(progressionIDString))
		if err != nil {
			return nil, nil, nil, fmt.Errorf("progressionIds[%d] must be a valid UUID", index)
		}
		if seenProgressionIDs[progressionID] {
			continue
		}
		seenProgressionIDs[progressionID] = true
		if _, err := s.dbClient.Spell().FindProgressionByID(ctx, progressionID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, nil, fmt.Errorf("progressionIds[%d] not found", index)
			}
			return nil, nil, nil, err
		}
		progressions = append(progressions, models.MonsterTemplateProgression{
			ProgressionID: progressionID,
		})
	}

	spells := []models.MonsterTemplateSpell{}
	seenSpellIDs := map[uuid.UUID]bool{}
	for index, spellIDString := range body.SpellIDs {
		spellID, err := uuid.Parse(strings.TrimSpace(spellIDString))
		if err != nil {
			return nil, nil, nil, fmt.Errorf("spellIds[%d] must be a valid UUID", index)
		}
		if seenSpellIDs[spellID] {
			continue
		}
		seenSpellIDs[spellID] = true
		if _, err := s.dbClient.Spell().FindByID(ctx, spellID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, nil, fmt.Errorf("spellIds[%d] not found", index)
			}
			return nil, nil, nil, err
		}
		spells = append(spells, models.MonsterTemplateSpell{SpellID: spellID})
	}

	template := &models.MonsterTemplate{
		Archived:                      body.Archived != nil && *body.Archived,
		MonsterType:                   monsterType,
		GenreID:                       genre.ID,
		Genre:                         genre,
		Name:                          name,
		Description:                   strings.TrimSpace(body.Description),
		ImageURL:                      strings.TrimSpace(body.ImageURL),
		ThumbnailURL:                  strings.TrimSpace(body.ThumbnailURL),
		BaseStrength:                  body.BaseStrength,
		BaseDexterity:                 body.BaseDexterity,
		BaseConstitution:              body.BaseConstitution,
		BaseIntelligence:              body.BaseIntelligence,
		BaseWisdom:                    body.BaseWisdom,
		BaseCharisma:                  body.BaseCharisma,
		PhysicalDamageBonusPercent:    body.PhysicalDamageBonusPercent,
		PiercingDamageBonusPercent:    body.PiercingDamageBonusPercent,
		SlashingDamageBonusPercent:    body.SlashingDamageBonusPercent,
		BludgeoningDamageBonusPercent: body.BludgeoningDamageBonusPercent,
		FireDamageBonusPercent:        body.FireDamageBonusPercent,
		IceDamageBonusPercent:         body.IceDamageBonusPercent,
		LightningDamageBonusPercent:   body.LightningDamageBonusPercent,
		PoisonDamageBonusPercent:      body.PoisonDamageBonusPercent,
		ArcaneDamageBonusPercent:      body.ArcaneDamageBonusPercent,
		HolyDamageBonusPercent:        body.HolyDamageBonusPercent,
		ShadowDamageBonusPercent:      body.ShadowDamageBonusPercent,
		PhysicalResistancePercent:     body.PhysicalResistancePercent,
		PiercingResistancePercent:     body.PiercingResistancePercent,
		SlashingResistancePercent:     body.SlashingResistancePercent,
		BludgeoningResistancePercent:  body.BludgeoningResistancePercent,
		FireResistancePercent:         body.FireResistancePercent,
		IceResistancePercent:          body.IceResistancePercent,
		LightningResistancePercent:    body.LightningResistancePercent,
		PoisonResistancePercent:       body.PoisonResistancePercent,
		ArcaneResistancePercent:       body.ArcaneResistancePercent,
		HolyResistancePercent:         body.HolyResistancePercent,
		ShadowResistancePercent:       body.ShadowResistancePercent,
	}
	if template.ThumbnailURL == "" && template.ImageURL != "" {
		template.ThumbnailURL = template.ImageURL
	}
	return template, progressions, spells, nil
}

func (s *server) parseMonsterUpsertRequest(
	ctx context.Context,
	body monsterUpsertRequest,
) (*models.Monster, []models.MonsterItemReward, error) {
	zoneID, err := uuid.Parse(strings.TrimSpace(body.ZoneID))
	if err != nil {
		return nil, nil, fmt.Errorf("zoneId must be a valid UUID")
	}
	if _, err := s.dbClient.Zone().FindByID(ctx, zoneID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("zone not found")
		}
		return nil, nil, err
	}

	templateID, err := uuid.Parse(strings.TrimSpace(body.TemplateID))
	if err != nil {
		return nil, nil, fmt.Errorf("templateId must be a valid UUID")
	}
	template, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("template not found")
		}
		return nil, nil, err
	}

	if body.Level < 1 {
		return nil, nil, fmt.Errorf("level must be positive")
	}
	if body.RewardExperience < 0 {
		return nil, nil, fmt.Errorf("rewardExperience must be zero or greater")
	}
	if body.RewardGold < 0 {
		return nil, nil, fmt.Errorf("rewardGold must be zero or greater")
	}
	materialRewards, err := parseBaseMaterialRewards(body.MaterialRewards, "materialRewards")
	if err != nil {
		return nil, nil, err
	}
	rewardMode := models.NormalizeRewardMode(body.RewardMode)
	if strings.TrimSpace(body.RewardMode) == "" {
		if body.RewardExperience > 0 || body.RewardGold > 0 || len(body.ItemRewards) > 0 || len(materialRewards) > 0 {
			rewardMode = models.RewardModeExplicit
		}
	}
	randomRewardSize := models.NormalizeRandomRewardSize(body.RandomRewardSize)

	dominantItemID := body.DominantHandInventoryItemID
	if dominantItemID == nil || (dominantItemID != nil && *dominantItemID <= 0) {
		if body.WeaponInventoryItemID != nil && *body.WeaponInventoryItemID > 0 {
			dominantItemID = body.WeaponInventoryItemID
		}
	}
	if dominantItemID == nil || *dominantItemID <= 0 {
		return nil, nil, fmt.Errorf("dominantHandInventoryItemId is required")
	}
	dominantItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, *dominantItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("dominantHandInventoryItemId not found")
		}
		return nil, nil, err
	}
	if !isEligibleMonsterDominantHandItem(dominantItem) {
		return nil, nil, fmt.Errorf("dominantHandInventoryItemId must reference an eligible dominant-hand weapon or staff")
	}

	var offHandItemID *int
	var offHandItem *models.InventoryItem
	if body.OffHandInventoryItemID != nil && *body.OffHandInventoryItemID <= 0 {
		return nil, nil, fmt.Errorf("offHandInventoryItemId must be positive when set")
	}
	if body.OffHandInventoryItemID != nil && *body.OffHandInventoryItemID > 0 {
		offHandItemID = body.OffHandInventoryItemID
		offHandItem, err = s.dbClient.InventoryItem().FindInventoryItemByID(ctx, *offHandItemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fmt.Errorf("offHandInventoryItemId not found")
			}
			return nil, nil, err
		}
		if !isEligibleMonsterOffHandItem(offHandItem) {
			return nil, nil, fmt.Errorf("offHandInventoryItemId must reference an eligible off-hand item or one-handed weapon")
		}
		if isTwoHandedDominantItem(dominantItem) {
			return nil, nil, fmt.Errorf("offHandInventoryItemId cannot be set when dominant hand item is two_handed")
		}
	}

	imageURL := strings.TrimSpace(body.ImageURL)
	thumbnailURL := strings.TrimSpace(body.ThumbnailURL)
	if imageURL == "" {
		imageURL = strings.TrimSpace(template.ImageURL)
	}
	if thumbnailURL == "" {
		thumbnailURL = strings.TrimSpace(template.ThumbnailURL)
	}
	if thumbnailURL == "" && imageURL != "" {
		thumbnailURL = imageURL
	}

	itemQtyByID := map[int]int{}
	for index, reward := range body.ItemRewards {
		if reward.InventoryItemID <= 0 {
			return nil, nil, fmt.Errorf("itemRewards[%d].inventoryItemId must be positive", index)
		}
		if reward.Quantity <= 0 {
			return nil, nil, fmt.Errorf("itemRewards[%d].quantity must be positive", index)
		}
		if _, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, reward.InventoryItemID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fmt.Errorf("itemRewards[%d].inventoryItemId not found", index)
			}
			return nil, nil, err
		}
		itemQtyByID[reward.InventoryItemID] += reward.Quantity
	}
	itemRewards := make([]models.MonsterItemReward, 0, len(itemQtyByID))
	for inventoryItemID, quantity := range itemQtyByID {
		itemRewards = append(itemRewards, models.MonsterItemReward{
			InventoryItemID: inventoryItemID,
			Quantity:        quantity,
		})
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = template.Name
	}
	description := strings.TrimSpace(body.Description)
	if description == "" {
		description = template.Description
	}

	monster := &models.Monster{
		Name:                        name,
		Description:                 description,
		ImageURL:                    imageURL,
		ThumbnailURL:                thumbnailURL,
		ZoneID:                      zoneID,
		GenreID:                     template.GenreID,
		Genre:                       template.Genre,
		Latitude:                    body.Latitude,
		Longitude:                   body.Longitude,
		TemplateID:                  &templateID,
		DominantHandInventoryItemID: dominantItemID,
		OffHandInventoryItemID:      offHandItemID,
		WeaponInventoryItemID:       dominantItemID,
		Level:                       body.Level,
		RewardMode:                  rewardMode,
		RandomRewardSize:            randomRewardSize,
		RewardExperience:            body.RewardExperience,
		RewardGold:                  body.RewardGold,
		MaterialRewards:             materialRewards,
		ImageGenerationStatus:       models.MonsterImageGenerationStatusNone,
	}
	return monster, itemRewards, nil
}

func (s *server) parseMonsterEncounterUpsertRequest(
	ctx context.Context,
	body monsterEncounterUpsertRequest,
) (*models.MonsterEncounter, []models.MonsterEncounterMember, error) {
	zoneID, err := uuid.Parse(strings.TrimSpace(body.ZoneID))
	if err != nil {
		return nil, nil, fmt.Errorf("zoneId must be a valid UUID")
	}
	if _, err := s.dbClient.Zone().FindByID(ctx, zoneID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("zone not found")
		}
		return nil, nil, err
	}
	pointOfInterestID, err := parseStandalonePointOfInterestID(body.PointOfInterestID)
	if err != nil {
		return nil, nil, err
	}
	resolvedPointOfInterestID, resolvedLatitude, resolvedLongitude, err := s.resolveStandaloneLocation(
		ctx,
		&zoneID,
		pointOfInterestID,
		body.Latitude,
		body.Longitude,
	)
	if err != nil {
		return nil, nil, err
	}

	if len(body.MonsterIDs) < 1 || len(body.MonsterIDs) > 9 {
		return nil, nil, fmt.Errorf("monsterIds must include between 1 and 9 monsters")
	}
	if body.RewardExperience < 0 {
		return nil, nil, fmt.Errorf("rewardExperience must be zero or greater")
	}
	if body.RewardGold < 0 {
		return nil, nil, fmt.Errorf("rewardGold must be zero or greater")
	}
	materialRewards, err := parseBaseMaterialRewards(body.MaterialRewards, "materialRewards")
	if err != nil {
		return nil, nil, err
	}

	rewardMode := models.NormalizeRewardMode(body.RewardMode)
	if strings.TrimSpace(body.RewardMode) == "" {
		if body.RewardExperience > 0 || body.RewardGold > 0 || len(body.ItemRewards) > 0 || len(materialRewards) > 0 {
			rewardMode = models.RewardModeExplicit
		}
	}
	randomRewardSize := models.NormalizeRandomRewardSize(body.RandomRewardSize)

	itemQtyByID := map[int]int{}
	for index, reward := range body.ItemRewards {
		if reward.InventoryItemID <= 0 {
			return nil, nil, fmt.Errorf("itemRewards[%d].inventoryItemId must be positive", index)
		}
		if reward.Quantity <= 0 {
			return nil, nil, fmt.Errorf("itemRewards[%d].quantity must be positive", index)
		}
		if _, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, reward.InventoryItemID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fmt.Errorf("itemRewards[%d].inventoryItemId not found", index)
			}
			return nil, nil, err
		}
		itemQtyByID[reward.InventoryItemID] += reward.Quantity
	}
	itemRewards := make(models.MonsterEncounterRewardItems, 0, len(itemQtyByID))
	for inventoryItemID, quantity := range itemQtyByID {
		itemRewards = append(itemRewards, models.MonsterEncounterRewardItem{
			InventoryItemID: inventoryItemID,
			Quantity:        quantity,
		})
	}

	seenMonsterIDs := map[uuid.UUID]struct{}{}
	members := make([]models.MonsterEncounterMember, 0, len(body.MonsterIDs))
	resolvedMonsters := make([]*models.Monster, 0, len(body.MonsterIDs))
	for index, raw := range body.MonsterIDs {
		monsterID, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return nil, nil, fmt.Errorf("monsterIds[%d] must be a valid UUID", index)
		}
		if _, exists := seenMonsterIDs[monsterID]; exists {
			continue
		}
		seenMonsterIDs[monsterID] = struct{}{}
		monster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fmt.Errorf("monsterIds[%d] not found", index)
			}
			return nil, nil, err
		}
		if monster.ZoneID != zoneID {
			return nil, nil, fmt.Errorf("monsterIds[%d] belongs to a different zone", index)
		}
		resolvedMonsters = append(resolvedMonsters, monster)
		members = append(members, models.MonsterEncounterMember{
			MonsterID: monsterID,
			Slot:      len(members) + 1,
		})
	}

	if len(members) < 1 || len(members) > 9 {
		return nil, nil, fmt.Errorf("monsterIds must include between 1 and 9 unique monsters")
	}

	encounterType := models.NormalizeMonsterEncounterType(body.EncounterType)
	name := strings.TrimSpace(body.Name)
	if name == "" {
		encounterLabel := "Encounter"
		switch encounterType {
		case models.MonsterEncounterTypeBoss:
			encounterLabel = "Boss Encounter"
		case models.MonsterEncounterTypeRaid:
			encounterLabel = "Raid Encounter"
		}
		if len(resolvedMonsters) == 1 {
			name = fmt.Sprintf("%s %s", strings.TrimSpace(resolvedMonsters[0].Name), encounterLabel)
		} else {
			name = fmt.Sprintf("%d-Monster %s", len(resolvedMonsters), encounterLabel)
		}
	}

	description := strings.TrimSpace(body.Description)
	if description == "" && len(resolvedMonsters) > 0 {
		description = strings.TrimSpace(resolvedMonsters[0].Description)
	}

	imageURL := strings.TrimSpace(body.ImageURL)
	thumbnailURL := strings.TrimSpace(body.ThumbnailURL)
	if imageURL == "" && len(resolvedMonsters) > 0 {
		imageURL = strings.TrimSpace(resolvedMonsters[0].ImageURL)
		if imageURL == "" {
			imageURL = strings.TrimSpace(resolvedMonsters[0].ThumbnailURL)
		}
	}
	if thumbnailURL == "" && len(resolvedMonsters) > 0 {
		thumbnailURL = strings.TrimSpace(resolvedMonsters[0].ThumbnailURL)
		if thumbnailURL == "" {
			thumbnailURL = strings.TrimSpace(resolvedMonsters[0].ImageURL)
		}
	}
	if thumbnailURL == "" && imageURL != "" {
		thumbnailURL = imageURL
	}

	encounter := &models.MonsterEncounter{
		Name:               name,
		Description:        description,
		ImageURL:           imageURL,
		ThumbnailURL:       thumbnailURL,
		EncounterType:      encounterType,
		RewardMode:         rewardMode,
		RandomRewardSize:   randomRewardSize,
		RewardExperience:   body.RewardExperience,
		RewardGold:         body.RewardGold,
		MaterialRewards:    materialRewards,
		ItemRewards:        itemRewards,
		ScaleWithUserLevel: body.ScaleWithUserLevel,
		ZoneID:             zoneID,
		PointOfInterestID:  resolvedPointOfInterestID,
		Latitude:           resolvedLatitude,
		Longitude:          resolvedLongitude,
	}
	return encounter, members, nil
}

func isEligibleMonsterDominantHandItem(item *models.InventoryItem) bool {
	if item == nil || item.EquipSlot == nil {
		return false
	}
	if strings.TrimSpace(*item.EquipSlot) != string(models.EquipmentSlotDominantHand) {
		return false
	}
	if item.HandItemCategory == nil {
		return false
	}
	category := strings.TrimSpace(*item.HandItemCategory)
	if category != string(models.HandItemCategoryWeapon) && category != string(models.HandItemCategoryStaff) {
		return false
	}
	return item.DamageMin != nil && item.DamageMax != nil && item.SwipesPerAttack != nil
}

func isEligibleMonsterOffHandItem(item *models.InventoryItem) bool {
	if item == nil || item.EquipSlot == nil {
		return false
	}
	equipSlot := strings.TrimSpace(*item.EquipSlot)
	category := ""
	if item.HandItemCategory != nil {
		category = strings.TrimSpace(*item.HandItemCategory)
	}
	handedness := ""
	if item.Handedness != nil {
		handedness = strings.TrimSpace(*item.Handedness)
	}
	if equipSlot == string(models.EquipmentSlotOffHand) &&
		handedness == string(models.HandednessOneHanded) &&
		(category == string(models.HandItemCategoryShield) || category == string(models.HandItemCategoryOrb)) {
		return true
	}
	if equipSlot == string(models.EquipmentSlotDominantHand) &&
		handedness == string(models.HandednessOneHanded) &&
		category == string(models.HandItemCategoryWeapon) &&
		item.DamageMin != nil &&
		item.DamageMax != nil &&
		item.SwipesPerAttack != nil {
		return true
	}
	return false
}

func (s *server) getMonsterTemplates(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	templates, err := s.dbClient.MonsterTemplate().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]monsterTemplateResponse, 0, len(templates))
	for i := range templates {
		response = append(response, *monsterTemplateResponseFrom(&templates[i]))
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getAdminMonsterTemplates(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	archived, err := parseOptionalArchivedFilter(ctx.Query("archived"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	genreID, err := parseOptionalGenreIDFilter(ctx.Query("genreId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page := parseAdminMonsterListPage(ctx)
	pageSize := parseAdminMonsterListPageSize(ctx)
	result, err := s.dbClient.MonsterTemplate().ListAdmin(ctx, db.MonsterTemplateAdminListParams{
		Page:        page,
		PageSize:    pageSize,
		Query:       ctx.Query("query"),
		ZoneQuery:   ctx.Query("zoneQuery"),
		GenreID:     genreID,
		Archived:    archived,
		MonsterType: ctx.Query("monsterType"),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]monsterTemplateResponse, 0, len(result.Templates))
	for i := range result.Templates {
		items = append(items, *monsterTemplateResponseFrom(&result.Templates[i]))
	}

	ctx.JSON(http.StatusOK, paginatedMonsterTemplateResponse{
		Items:         items,
		Total:         result.Total,
		Page:          page,
		PageSize:      pageSize,
		ActiveCount:   result.ActiveCount,
		ArchivedCount: result.ArchivedCount,
	})
}

func (s *server) getMonsterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	template, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, monsterTemplateResponseFrom(template))
}

func (s *server) createMonsterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody monsterTemplateUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, progressions, spells, err := s.parseMonsterTemplateUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if template.ImageURL != "" {
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusComplete
		emptyError := ""
		template.ImageGenerationError = &emptyError
	} else {
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusNone
		emptyError := ""
		template.ImageGenerationError = &emptyError
	}

	if err := s.dbClient.MonsterTemplate().Create(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MonsterTemplate().ReplaceProgressions(ctx, template.ID, progressions); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MonsterTemplate().ReplaceSpells(ctx, template.ID, spells); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.dbClient.MonsterTemplate().FindByID(ctx, template.ID)
	if err != nil {
		ctx.JSON(http.StatusCreated, monsterTemplateResponseFrom(template))
		return
	}
	ctx.JSON(http.StatusCreated, monsterTemplateResponseFrom(created))
}

func nextUniqueMonsterTemplateName(base string, used map[string]struct{}) string {
	trimmed := strings.TrimSpace(base)
	if trimmed == "" {
		trimmed = "Monster Template"
	}
	candidate := trimmed
	suffix := 2
	for {
		key := strings.ToLower(strings.TrimSpace(candidate))
		if _, exists := used[key]; !exists {
			used[key] = struct{}{}
			return candidate
		}
		candidate = fmt.Sprintf("%s %d", trimmed, suffix)
		suffix++
	}
}

func monsterTemplateTypePromptLabel(monsterType models.MonsterTemplateType) string {
	switch monsterType {
	case models.MonsterTemplateTypeBoss:
		return "boss"
	case models.MonsterTemplateTypeRaid:
		return "raid"
	default:
		return "standard"
	}
}

func monsterTemplateTypePromptGuidance(monsterType models.MonsterTemplateType) string {
	switch monsterType {
	case models.MonsterTemplateTypeBoss:
		return "- Boss templates should feel like elite solo threats suited for centerpiece fights.\n- Favor commanding names, climactic descriptions, and a strong single-foe identity."
	case models.MonsterTemplateTypeRaid:
		return "- Raid templates should feel like apex threats intended to pressure a full five-player party.\n- Favor large-scale menace, battlefield presence, and dramatic danger."
	default:
		return "- Standard templates should feel like everyday field monsters, ambushers, skirmishers, or common elites."
	}
}

func buildMonsterTemplateGenerationPrompt(
	count int,
	monsterType models.MonsterTemplateType,
	genre *models.ZoneGenre,
	denyList []string,
) string {
	if isBaselineFantasyMonsterGenre(genre) {
		return fmt.Sprintf(
			generateMonsterTemplatesPromptTemplate,
			count,
			monsterTemplateTypePromptLabel(monsterType),
			monsterTemplateTypePromptGuidance(monsterType),
			formatMonsterTemplateNamesForPrompt(denyList),
			count,
		)
	}

	genreDirection := monsterGenrePromptSeed(genre)
	if genreDirection == "" {
		genreDirection = fmt.Sprintf(
			"Lean into %s fiction, aesthetics, vocabulary, and creature logic while still fitting an action RPG combat loop.",
			monsterGenrePromptLabel(genre),
		)
	}

	return fmt.Sprintf(
		generateMonsterTemplatesWithGenrePromptTemplate,
		count,
		monsterTemplateTypePromptLabel(monsterType),
		monsterGenrePromptLabel(genre),
		genreDirection,
		monsterTemplateTypePromptGuidance(monsterType),
		formatMonsterTemplateNamesForPrompt(denyList),
		count,
	)
}

func buildBulkMonsterTemplateSpecsFromSeeds(
	count int,
	usedNames map[string]struct{},
	monsterType models.MonsterTemplateType,
	genre *models.ZoneGenre,
) []jobs.MonsterTemplateCreationSpec {
	specs := make([]jobs.MonsterTemplateCreationSpec, 0, count)
	if count <= 0 {
		return specs
	}
	seedPool := dndMonsterTemplateSeeds
	if !isBaselineFantasyMonsterGenre(genre) {
		seedPool = genericMonsterTemplateRoleSeeds
	}
	if len(seedPool) == 0 {
		return specs
	}
	genreID := ""
	if genre != nil {
		genreID = genre.ID.String()
	}
	for i := 0; i < count; i++ {
		seed := seedPool[i%len(seedPool)]
		name := strings.TrimSpace(seed.Name)
		description := strings.TrimSpace(seed.Description)
		if !isBaselineFantasyMonsterGenre(genre) {
			name = fmt.Sprintf("%s %s", strings.TrimSpace(genre.Name), name)
			description = fmt.Sprintf(
				"%s Designed for a %s action RPG setting.",
				description,
				monsterGenrePromptLabel(genre),
			)
		}
		specs = append(specs, jobs.MonsterTemplateCreationSpec{
			MonsterType:      string(monsterType),
			GenreID:          genreID,
			Name:             nextUniqueMonsterTemplateName(name, usedNames),
			Description:      description,
			BaseStrength:     seed.BaseStrength,
			BaseDexterity:    seed.BaseDexterity,
			BaseConstitution: seed.BaseConstitution,
			BaseIntelligence: seed.BaseIntelligence,
			BaseWisdom:       seed.BaseWisdom,
			BaseCharisma:     seed.BaseCharisma,
		})
	}
	return specs
}

func sanitizeMonsterTemplateSpec(spec jobs.MonsterTemplateCreationSpec) jobs.MonsterTemplateCreationSpec {
	spec.MonsterType = string(models.NormalizeMonsterTemplateType(spec.MonsterType))
	spec.Name = strings.TrimSpace(spec.Name)
	spec.Description = strings.TrimSpace(spec.Description)
	if spec.Description == "" {
		spec.Description = "A dangerous creature with a specialized combat role."
	}
	spec.BaseStrength = clampMonsterTemplateStat(spec.BaseStrength)
	spec.BaseDexterity = clampMonsterTemplateStat(spec.BaseDexterity)
	spec.BaseConstitution = clampMonsterTemplateStat(spec.BaseConstitution)
	spec.BaseIntelligence = clampMonsterTemplateStat(spec.BaseIntelligence)
	spec.BaseWisdom = clampMonsterTemplateStat(spec.BaseWisdom)
	spec.BaseCharisma = clampMonsterTemplateStat(spec.BaseCharisma)
	return spec
}

func clampMonsterTemplateStat(value int) int {
	if value < 1 {
		return 10
	}
	if value > 20 {
		return 20
	}
	return value
}

func formatMonsterTemplateNamesForPrompt(names []string) string {
	if len(names) == 0 {
		return "(none)"
	}

	sorted := make([]string, 0, len(names))
	seen := map[string]struct{}{}
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		normalized := strings.ToLower(trimmed)
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		sorted = append(sorted, trimmed)
	}
	sort.Strings(sorted)
	if len(sorted) == 0 {
		return "(none)"
	}

	const maxNames = 200
	limited := sorted
	remaining := 0
	if len(sorted) > maxNames {
		limited = sorted[:maxNames]
		remaining = len(sorted) - maxNames
	}

	var builder strings.Builder
	for _, name := range limited {
		builder.WriteString("- ")
		builder.WriteString(name)
		builder.WriteByte('\n')
	}
	if remaining > 0 {
		builder.WriteString(fmt.Sprintf("- ... and %d more\n", remaining))
	}
	return strings.TrimSpace(builder.String())
}

func extractJSONPayload(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}

	objectStart := strings.Index(trimmed, "{")
	arrayStart := strings.Index(trimmed, "[")
	start := -1
	end := -1

	if objectStart >= 0 && (arrayStart < 0 || objectStart < arrayStart) {
		start = objectStart
		end = strings.LastIndex(trimmed, "}")
	} else if arrayStart >= 0 {
		start = arrayStart
		end = strings.LastIndex(trimmed, "]")
	}

	if start >= 0 && end >= start {
		return strings.TrimSpace(trimmed[start : end+1])
	}
	return trimmed
}

func parseGeneratedMonsterTemplates(raw string) ([]jobs.MonsterTemplateCreationSpec, error) {
	payload := extractJSONPayload(raw)
	if payload == "" {
		return nil, fmt.Errorf("empty generation payload")
	}

	var wrapped generatedMonsterTemplatePayload
	if err := json.Unmarshal([]byte(payload), &wrapped); err == nil && len(wrapped.Templates) > 0 {
		return wrapped.Templates, nil
	}

	var list []jobs.MonsterTemplateCreationSpec
	if err := json.Unmarshal([]byte(payload), &list); err == nil && len(list) > 0 {
		return list, nil
	}

	return nil, fmt.Errorf("invalid monster template generation payload")
}

func (s *server) buildBulkMonsterTemplateSpecs(
	count int,
	usedNames map[string]struct{},
	existingNames []string,
	monsterType models.MonsterTemplateType,
	genre *models.ZoneGenre,
) ([]jobs.MonsterTemplateCreationSpec, string, error) {
	if count <= 0 {
		return []jobs.MonsterTemplateCreationSpec{}, "none", nil
	}

	specs := make([]jobs.MonsterTemplateCreationSpec, 0, count)
	source := "seed_generated"

	if s.deepPriest != nil {
		aiSpecs, err := s.generateMonsterTemplateSpecsWithLLM(
			count,
			usedNames,
			existingNames,
			monsterType,
			genre,
		)
		if err == nil && len(aiSpecs) > 0 {
			specs = append(specs, aiSpecs...)
			source = "ai_generated"
		}
	}

	if remaining := count - len(specs); remaining > 0 {
		fallback := buildBulkMonsterTemplateSpecsFromSeeds(
			remaining,
			usedNames,
			monsterType,
			genre,
		)
		specs = append(specs, fallback...)
		if source == "ai_generated" {
			source = "ai_generated_with_seed_fallback"
		}
	}

	if len(specs) == 0 {
		return nil, "none", fmt.Errorf("no templates prepared for generation")
	}

	if len(specs) > count {
		specs = specs[:count]
	}
	return specs, source, nil
}

func (s *server) generateMonsterTemplateSpecsWithLLM(
	count int,
	usedNames map[string]struct{},
	existingNames []string,
	monsterType models.MonsterTemplateType,
	genre *models.ZoneGenre,
) ([]jobs.MonsterTemplateCreationSpec, error) {
	specs := make([]jobs.MonsterTemplateCreationSpec, 0, count)
	if count <= 0 {
		return specs, nil
	}

	denyList := make([]string, 0, len(existingNames)+len(usedNames))
	denyList = append(denyList, existingNames...)
	for used := range usedNames {
		denyList = append(denyList, used)
	}

	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts && len(specs) < count; attempt++ {
		remaining := count - len(specs)
		prompt := buildMonsterTemplateGenerationPrompt(
			remaining,
			monsterType,
			genre,
			denyList,
		)
		answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{
			Question: prompt,
		})
		if err != nil {
			continue
		}

		candidates, err := parseGeneratedMonsterTemplates(answer.Answer)
		if err != nil {
			continue
		}

		for _, candidate := range candidates {
			if len(specs) >= count {
				break
			}
			candidate = sanitizeMonsterTemplateSpec(candidate)
			candidate.MonsterType = string(monsterType)
			if genre != nil {
				candidate.GenreID = genre.ID.String()
			}
			if candidate.Name == "" {
				continue
			}
			normalized := strings.ToLower(candidate.Name)
			if _, exists := usedNames[normalized]; exists {
				continue
			}
			usedNames[normalized] = struct{}{}
			denyList = append(denyList, candidate.Name)
			specs = append(specs, candidate)
		}
	}

	if len(specs) == 0 {
		return nil, fmt.Errorf("failed to generate monster templates with llm")
	}
	return specs, nil
}

func (s *server) setMonsterTemplateBulkStatus(ctx context.Context, status jobs.MonsterTemplateBulkStatus) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client unavailable")
	}
	payload, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return s.redisClient.Set(
		ctx,
		jobs.MonsterTemplateBulkStatusKey(status.JobID),
		payload,
		jobs.MonsterTemplateBulkStatusTTL,
	).Err()
}

func (s *server) getMonsterTemplateBulkStatus(ctx context.Context, jobID uuid.UUID) (*jobs.MonsterTemplateBulkStatus, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client unavailable")
	}
	value, err := s.redisClient.Get(ctx, jobs.MonsterTemplateBulkStatusKey(jobID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var status jobs.MonsterTemplateBulkStatus
	if err := json.Unmarshal([]byte(value), &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *server) setMonsterTemplateAffinityRefreshStatus(
	ctx context.Context,
	status jobs.MonsterTemplateAffinityRefreshStatus,
) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client unavailable")
	}
	payload, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return s.redisClient.Set(
		ctx,
		jobs.MonsterTemplateAffinityRefreshStatusKey(status.JobID),
		payload,
		jobs.MonsterTemplateAffinityRefreshStatusTTL,
	).Err()
}

func (s *server) getMonsterTemplateAffinityRefreshStatus(
	ctx context.Context,
	jobID uuid.UUID,
) (*jobs.MonsterTemplateAffinityRefreshStatus, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client unavailable")
	}
	value, err := s.redisClient.Get(ctx, jobs.MonsterTemplateAffinityRefreshStatusKey(jobID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var status jobs.MonsterTemplateAffinityRefreshStatus
	if err := json.Unmarshal([]byte(value), &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *server) setMonsterTemplateProgressionResetStatus(
	ctx context.Context,
	status jobs.MonsterTemplateProgressionResetStatus,
) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client unavailable")
	}
	payload, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return s.redisClient.Set(
		ctx,
		jobs.MonsterTemplateProgressionResetStatusKey(status.JobID),
		payload,
		jobs.MonsterTemplateProgressionResetStatusTTL,
	).Err()
}

func (s *server) getMonsterTemplateProgressionResetStatus(
	ctx context.Context,
	jobID uuid.UUID,
) (*jobs.MonsterTemplateProgressionResetStatus, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client unavailable")
	}
	value, err := s.redisClient.Get(ctx, jobs.MonsterTemplateProgressionResetStatusKey(jobID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var status jobs.MonsterTemplateProgressionResetStatus
	if err := json.Unmarshal([]byte(value), &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *server) bulkGenerateMonsterTemplates(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody bulkGenerateMonsterTemplatesRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.Count < 1 || requestBody.Count > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "count must be between 1 and 100"})
		return
	}
	monsterType := models.NormalizeMonsterTemplateType(requestBody.MonsterType)
	genre, err := s.resolveMonsterGenre(ctx, requestBody.GenreID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if s.asyncClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "async client unavailable"})
		return
	}
	if s.redisClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "redis client unavailable"})
		return
	}

	existingTemplates, err := s.dbClient.MonsterTemplate().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	usedNames := make(map[string]struct{}, len(existingTemplates)+requestBody.Count)
	for _, template := range existingTemplates {
		normalized := strings.ToLower(strings.TrimSpace(template.Name))
		if normalized == "" {
			continue
		}
		usedNames[normalized] = struct{}{}
	}

	existingNames := make([]string, 0, len(existingTemplates))
	for _, template := range existingTemplates {
		name := strings.TrimSpace(template.Name)
		if name == "" {
			continue
		}
		existingNames = append(existingNames, name)
	}

	templateSpecs, source, err := s.buildBulkMonsterTemplateSpecs(
		requestBody.Count,
		usedNames,
		existingNames,
		monsterType,
		genre,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New()
	queuedAt := time.Now().UTC()
	status := jobs.MonsterTemplateBulkStatus{
		JobID:        jobID,
		Status:       jobs.MonsterTemplateBulkStatusQueued,
		Source:       source,
		MonsterType:  string(monsterType),
		TotalCount:   len(templateSpecs),
		CreatedCount: 0,
		QueuedAt:     &queuedAt,
		UpdatedAt:    queuedAt,
	}
	if err := s.setMonsterTemplateBulkStatus(ctx.Request.Context(), status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload := jobs.GenerateMonsterTemplatesBulkTaskPayload{
		JobID:       jobID,
		Source:      source,
		MonsterType: string(monsterType),
		TotalCount:  len(templateSpecs),
		Templates:   templateSpecs,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateMonsterTemplatesBulkTaskType, payloadBytes)); err != nil {
		failedAt := time.Now().UTC()
		status.Status = jobs.MonsterTemplateBulkStatusFailed
		status.Error = err.Error()
		status.CompletedAt = &failedAt
		status.UpdatedAt = failedAt
		_ = s.setMonsterTemplateBulkStatus(ctx.Request.Context(), status)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"jobId":        status.JobID,
		"status":       status.Status,
		"source":       status.Source,
		"monsterType":  status.MonsterType,
		"totalCount":   status.TotalCount,
		"createdCount": status.CreatedCount,
		"queuedAt":     status.QueuedAt,
		"updatedAt":    status.UpdatedAt,
	})
}

func (s *server) getBulkGenerateMonsterTemplatesStatus(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	jobID, err := uuid.Parse(ctx.Param("jobId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	status, err := s.getMonsterTemplateBulkStatus(ctx.Request.Context(), jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if status == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "bulk generation job not found"})
		return
	}

	ctx.JSON(http.StatusOK, status)
}

func (s *server) refreshMonsterTemplateAffinities(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if s.asyncClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "async client unavailable"})
		return
	}
	if s.redisClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "redis client unavailable"})
		return
	}

	var requestBody struct {
		IDs []string `json:"ids"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil && !errors.Is(err, io.EOF) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	templateIDs := make([]uuid.UUID, 0, len(requestBody.IDs))
	seen := map[uuid.UUID]struct{}{}
	for _, rawID := range requestBody.IDs {
		templateID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid template ID: %s", rawID)})
			return
		}
		if _, exists := seen[templateID]; exists {
			continue
		}
		seen[templateID] = struct{}{}
		templateIDs = append(templateIDs, templateID)
	}

	totalCount := len(templateIDs)
	if totalCount == 0 {
		templates, err := s.dbClient.MonsterTemplate().FindAll(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		totalCount = len(templates)
	}

	jobID := uuid.New()
	queuedAt := time.Now().UTC()
	status := jobs.MonsterTemplateAffinityRefreshStatus{
		JobID:        jobID,
		Status:       jobs.MonsterTemplateAffinityRefreshStatusQueued,
		TotalCount:   totalCount,
		UpdatedCount: 0,
		TemplateIDs:  templateIDs,
		QueuedAt:     &queuedAt,
		UpdatedAt:    queuedAt,
	}
	if err := s.setMonsterTemplateAffinityRefreshStatus(ctx.Request.Context(), status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.RefreshMonsterTemplateAffinitiesTaskPayload{
		JobID:              jobID,
		MonsterTemplateIDs: templateIDs,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.RefreshMonsterTemplateAffinitiesTaskType, payloadBytes)); err != nil {
		failedAt := time.Now().UTC()
		status.Status = jobs.MonsterTemplateAffinityRefreshStatusFailed
		status.Error = err.Error()
		status.CompletedAt = &failedAt
		status.UpdatedAt = failedAt
		_ = s.setMonsterTemplateAffinityRefreshStatus(ctx.Request.Context(), status)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, status)
}

func (s *server) getRefreshMonsterTemplateAffinitiesStatus(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	jobID, err := uuid.Parse(ctx.Param("jobId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	status, err := s.getMonsterTemplateAffinityRefreshStatus(ctx.Request.Context(), jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if status == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "affinity refresh job not found"})
		return
	}

	ctx.JSON(http.StatusOK, status)
}

func (s *server) resetMonsterTemplateProgressions(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if s.asyncClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "async client unavailable"})
		return
	}
	if s.redisClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "redis client unavailable"})
		return
	}

	var requestBody struct {
		IDs []string `json:"ids"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil && !errors.Is(err, io.EOF) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	templateIDs := make([]uuid.UUID, 0, len(requestBody.IDs))
	seen := map[uuid.UUID]struct{}{}
	for _, rawID := range requestBody.IDs {
		templateID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid template ID: %s", rawID)})
			return
		}
		if _, exists := seen[templateID]; exists {
			continue
		}
		seen[templateID] = struct{}{}
		templateIDs = append(templateIDs, templateID)
	}

	totalCount := len(templateIDs)
	if totalCount == 0 {
		templates, err := s.dbClient.MonsterTemplate().FindAll(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		totalCount = len(templates)
	}

	jobID := uuid.New()
	queuedAt := time.Now().UTC()
	status := jobs.MonsterTemplateProgressionResetStatus{
		JobID:        jobID,
		Status:       jobs.MonsterTemplateProgressionResetStatusQueued,
		TotalCount:   totalCount,
		UpdatedCount: 0,
		TemplateIDs:  templateIDs,
		QueuedAt:     &queuedAt,
		UpdatedAt:    queuedAt,
	}
	if err := s.setMonsterTemplateProgressionResetStatus(ctx.Request.Context(), status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.ResetMonsterTemplateProgressionsTaskPayload{
		JobID:              jobID,
		MonsterTemplateIDs: templateIDs,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.ResetMonsterTemplateProgressionsTaskType, payloadBytes)); err != nil {
		failedAt := time.Now().UTC()
		status.Status = jobs.MonsterTemplateProgressionResetStatusFailed
		status.Error = err.Error()
		status.CompletedAt = &failedAt
		status.UpdatedAt = failedAt
		_ = s.setMonsterTemplateProgressionResetStatus(ctx.Request.Context(), status)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, status)
}

func (s *server) getResetMonsterTemplateProgressionsStatus(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	jobID, err := uuid.Parse(ctx.Param("jobId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	status, err := s.getMonsterTemplateProgressionResetStatus(ctx.Request.Context(), jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if status == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "progression reset job not found"})
		return
	}

	ctx.JSON(http.StatusOK, status)
}

func (s *server) updateMonsterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}
	existingTemplate, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody monsterTemplateUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, progressions, spells, err := s.parseMonsterTemplateUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template.Archived = existingTemplate.Archived
	if requestBody.Archived != nil {
		template.Archived = *requestBody.Archived
	}
	if template.ImageURL != "" {
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusComplete
		emptyError := ""
		template.ImageGenerationError = &emptyError
	} else if existingTemplate.ImageGenerationStatus == models.MonsterTemplateImageGenerationStatusQueued ||
		existingTemplate.ImageGenerationStatus == models.MonsterTemplateImageGenerationStatusInProgress {
		template.ImageGenerationStatus = existingTemplate.ImageGenerationStatus
		template.ImageGenerationError = existingTemplate.ImageGenerationError
	} else {
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusNone
		emptyError := ""
		template.ImageGenerationError = &emptyError
	}

	if err := s.dbClient.MonsterTemplate().Update(ctx, templateID, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MonsterTemplate().ReplaceProgressions(ctx, templateID, progressions); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MonsterTemplate().ReplaceSpells(ctx, templateID, spells); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"id": templateID})
		return
	}
	ctx.JSON(http.StatusOK, monsterTemplateResponseFrom(updated))
}

func (s *server) bulkArchiveMonsterTemplates(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		IDs      []string `json:"ids" binding:"required"`
		Archived *bool    `json:"archived"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ids array is required"})
		return
	}
	if len(requestBody.IDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ids array cannot be empty"})
		return
	}

	archived := true
	if requestBody.Archived != nil {
		archived = *requestBody.Archived
	}

	seen := map[uuid.UUID]struct{}{}
	updatedIDs := make([]uuid.UUID, 0, len(requestBody.IDs))
	for _, rawID := range requestBody.IDs {
		templateID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid template ID: %s", rawID)})
			return
		}
		if _, exists := seen[templateID]; exists {
			continue
		}
		seen[templateID] = struct{}{}

		template, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("template not found: %s", templateID)})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if template == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("template not found: %s", templateID)})
			return
		}
		if template.Archived != archived {
			template.Archived = archived
			if err := s.dbClient.MonsterTemplate().Update(ctx, templateID, template); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		updatedIDs = append(updatedIDs, templateID)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"count":    len(updatedIDs),
		"archived": archived,
		"ids":      updatedIDs,
	})
}

func (s *server) deleteMonsterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	if _, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	usageCount, err := s.dbClient.Monster().CountByTemplateID(ctx, templateID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if usageCount > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "template is in use by existing monsters"})
		return
	}

	if err := s.dbClient.MonsterTemplate().Delete(ctx, templateID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "template deleted successfully"})
}

func (s *server) generateMonsterTemplateImage(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	template, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusQueued
	emptyError := ""
	template.ImageGenerationError = &emptyError
	if err := s.dbClient.MonsterTemplate().Update(ctx, templateID, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue monster template image generation: " + err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.GenerateMonsterTemplateImageTaskPayload{MonsterTemplateID: templateID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build monster template image generation payload"})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateMonsterTemplateImageTaskType, payloadBytes)); err != nil {
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusFailed
		errMessage := err.Error()
		template.ImageGenerationError = &errMessage
		_ = s.dbClient.MonsterTemplate().Update(ctx, templateID, template)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue monster template image generation: " + err.Error()})
		return
	}

	updatedTemplate, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		ctx.JSON(http.StatusOK, monsterTemplateResponseFrom(template))
		return
	}
	ctx.JSON(http.StatusOK, monsterTemplateResponseFrom(updatedTemplate))
}

func (s *server) getMonsters(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsters, err := s.dbClient.Monster().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]monsterResponse, 0, len(monsters))
	for i := range monsters {
		if !monsterVisibleToUser(user.ID, &monsters[i]) {
			continue
		}
		entry, err := s.buildMonsterResponse(ctx, user.ID, &monsters[i])
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response = append(response, entry)
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getAdminMonsters(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	page := parseAdminMonsterListPage(ctx)
	pageSize := parseAdminMonsterListPageSize(ctx)
	genreID, err := parseOptionalGenreIDFilter(ctx.Query("genreId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := s.dbClient.Monster().ListAdmin(ctx, db.MonsterAdminListParams{
		Page:      page,
		PageSize:  pageSize,
		Query:     ctx.Query("query"),
		ZoneQuery: ctx.Query("zoneQuery"),
		GenreID:   genreID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]monsterResponse, 0, len(result.Monsters))
	for i := range result.Monsters {
		items = append(
			items,
			monsterResponseFrom(
				&result.Monsters[i],
				models.CharacterStatBonuses{},
				[]models.MonsterStatus{},
				nil,
				0,
			),
		)
	}

	ctx.JSON(http.StatusOK, paginatedMonsterResponse{
		Items:    items,
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *server) getMonsterEncounters(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	encounters, err := s.dbClient.MonsterEncounter().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	activeStoryFlags, err := s.loadUserStoryFlagMap(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]monsterEncounterResponse, 0, len(encounters))
	for i := range encounters {
		if !monsterEncounterAvailableForStoryFlags(&encounters[i], activeStoryFlags) {
			continue
		}
		entry, err := s.monsterEncounterResponseFrom(ctx, user.ID, &encounters[i], 1, false)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response = append(response, entry)
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getAdminMonsterEncounters(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	page := parseAdminMonsterListPage(ctx)
	pageSize := parseAdminMonsterListPageSize(ctx)
	genreID, err := parseOptionalGenreIDFilter(ctx.Query("genreId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := s.dbClient.MonsterEncounter().ListAdmin(ctx, db.MonsterEncounterAdminListParams{
		Page:      page,
		PageSize:  pageSize,
		Query:     ctx.Query("query"),
		ZoneQuery: ctx.Query("zoneQuery"),
		GenreID:   genreID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]monsterEncounterResponse, 0, len(result.Encounters))
	for i := range result.Encounters {
		entry, err := s.monsterEncounterResponseFrom(ctx, user.ID, &result.Encounters[i], 1, false)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = append(items, entry)
	}

	ctx.JSON(http.StatusOK, paginatedMonsterEncounterResponse{
		Items:    items,
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *server) getMonsterEncounter(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	encounterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster encounter ID"})
		return
	}

	encounter, err := s.dbClient.MonsterEncounter().FindByID(ctx, encounterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !monsterEncounterVisibleToUser(user.ID, encounter) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
		return
	}
	activeStoryFlags, err := s.loadUserStoryFlagMap(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !monsterEncounterAvailableForStoryFlags(encounter, activeStoryFlags) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
		return
	}
	userLevel, err := s.currentPartyMaxLevel(ctx, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response, err := s.monsterEncounterResponseFrom(ctx, user.ID, encounter, userLevel, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getMonsterEncountersForZone(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	encounters, err := s.dbClient.MonsterEncounter().FindByZoneIDExcludingQuestNodes(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	activeStoryFlags, err := s.loadUserStoryFlagMap(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defeatedEncounterIDs, err := s.dbClient.UserMonsterEncounterVictory().
		FindEncounterIDsByUserAndZone(ctx, user.ID, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defeatedSet := make(map[uuid.UUID]struct{}, len(defeatedEncounterIDs))
	for _, encounterID := range defeatedEncounterIDs {
		defeatedSet[encounterID] = struct{}{}
	}
	userLevel, err := s.currentPartyMaxLevel(ctx, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]monsterEncounterResponse, 0, len(encounters))
	for i := range encounters {
		if !monsterEncounterAvailableForStoryFlags(&encounters[i], activeStoryFlags) {
			continue
		}
		if _, defeated := defeatedSet[encounters[i].ID]; defeated {
			continue
		}
		if !monsterEncounterVisibleToUser(user.ID, &encounters[i]) {
			continue
		}
		entry, err := s.monsterEncounterResponseFrom(ctx, user.ID, &encounters[i], userLevel, true)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response = append(response, entry)
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) createMonsterEncounter(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody monsterEncounterUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encounter, members, err := s.parseMonsterEncounterUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := applyStandaloneRecurrenceForCreate(
		requestBody.RecurrenceFrequency,
		time.Now(),
		&encounter.RecurringMonsterEncounterID,
		&encounter.RecurrenceFrequency,
		&encounter.NextRecurrenceAt,
	); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.MonsterEncounter().Create(ctx, encounter); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MonsterEncounter().ReplaceMembers(ctx, encounter.ID, members); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.dbClient.MonsterEncounter().FindByID(ctx, encounter.ID)
	if err != nil {
		ctx.JSON(http.StatusCreated, encounter)
		return
	}
	response, err := s.monsterEncounterResponseFrom(ctx, user.ID, created, 1, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

func (s *server) updateMonsterEncounter(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	encounterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster encounter ID"})
		return
	}

	existing, err := s.dbClient.MonsterEncounter().FindByID(ctx, encounterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
		return
	}

	var requestBody monsterEncounterUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encounter, members, err := s.parseMonsterEncounterUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	encounter.RequiredStoryFlags = existing.RequiredStoryFlags
	if strings.TrimSpace(requestBody.PointOfInterestID) == "" &&
		existing.PointOfInterestID != nil &&
		requestBody.Latitude == existing.Latitude &&
		requestBody.Longitude == existing.Longitude {
		encounter.PointOfInterestID = existing.PointOfInterestID
	}
	encounter.RecurringMonsterEncounterID = existing.RecurringMonsterEncounterID
	encounter.RecurrenceFrequency = existing.RecurrenceFrequency
	encounter.NextRecurrenceAt = existing.NextRecurrenceAt
	if err := applyStandaloneRecurrenceForUpdate(
		requestBody.RecurrenceFrequency,
		time.Now(),
		&encounter.RecurringMonsterEncounterID,
		&encounter.RecurrenceFrequency,
		&encounter.NextRecurrenceAt,
	); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MonsterEncounter().Update(ctx, encounterID, encounter); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MonsterEncounter().ReplaceMembers(ctx, encounterID, members); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.MonsterEncounter().FindByID(ctx, encounterID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"id": encounterID})
		return
	}
	response, err := s.monsterEncounterResponseFrom(ctx, user.ID, updated, 1, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) deleteMonsterEncounter(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	encounterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster encounter ID"})
		return
	}

	if _, err := s.dbClient.MonsterEncounter().FindByID(ctx, encounterID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.MonsterEncounter().Delete(ctx, encounterID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "monster encounter deleted successfully"})
}

func (s *server) getMonster(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	monster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !monsterVisibleToUser(user.ID, monster) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
		return
	}
	response, err := s.buildMonsterResponse(ctx, user.ID, monster)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getMonstersForZone(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	monsters, err := s.dbClient.Monster().FindByZoneIDExcludingQuestNodes(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]monsterResponse, 0, len(monsters))
	for i := range monsters {
		entry, err := s.buildMonsterResponse(ctx, user.ID, &monsters[i])
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response = append(response, entry)
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) startMonsterBattle(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}
	monster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !monsterVisibleToUser(user.ID, monster) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
		return
	}

	var requestBody struct {
		MonsterEncounterID *string `json:"monsterEncounterId"`
	}
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&requestBody); err != nil && !errors.Is(err, io.EOF) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	var encounterID *uuid.UUID
	if requestBody.MonsterEncounterID != nil {
		trimmed := strings.TrimSpace(*requestBody.MonsterEncounterID)
		if trimmed != "" {
			parsedEncounterID, err := uuid.Parse(trimmed)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster encounter ID"})
				return
			}
			encounter, err := s.dbClient.MonsterEncounter().FindByID(ctx, parsedEncounterID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
					return
				}
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if !monsterEncounterVisibleToUser(user.ID, encounter) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
				return
			}
			activeStoryFlags, err := s.loadUserStoryFlagMap(ctx, user.ID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if !monsterEncounterAvailableForStoryFlags(encounter, activeStoryFlags) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
				return
			}
			memberMatch := false
			for _, member := range encounter.Members {
				if member.MonsterID == monsterID {
					memberMatch = true
					break
				}
			}
			if !memberMatch {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "monster is not part of the requested encounter"})
				return
			}
			encounterID = &parsedEncounterID
		}
	}

	battle, err := s.createFreshMonsterBattle(ctx, user.ID, monsterID, encounterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle, err = s.refreshMonsterBattleInviteState(ctx, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf(
		"[monster-rewards][start] user=%s monster=%s battle=%s encounter=%v",
		user.ID,
		monsterID,
		battle.ID,
		battle.MonsterEncounterID,
	)
	if battle.State == string(models.MonsterBattleStateActive) {
		now := time.Now()
		if err := s.dbClient.MonsterBattle().Touch(ctx, battle.ID, now); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		battle.LastActivityAt = now
	}

	response, err := s.monsterBattleDetailResponse(ctx, battle)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getMonsterBattleStatus(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	battle, err := s.findActiveMonsterBattleForUser(ctx, user.ID, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no active battle for this monster"})
		return
	}
	log.Printf(
		"[party-combat][status-by-monster] user=%s monster=%s battle=%s",
		user.ID,
		monsterID,
		battle.ID,
	)
	if battle, err = s.refreshMonsterBattleInviteState(ctx, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response, err := s.monsterBattleDetailResponse(ctx, battle)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getMonsterBattleStatusByID(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	battleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid battle ID"})
		return
	}

	battle, err := s.dbClient.MonsterBattle().FindByID(ctx, battleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "battle not found"})
		return
	}
	canAccess, err := s.userCanAccessMonsterBattle(ctx, user.ID, battle)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canAccess {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "not a battle participant"})
		return
	}
	log.Printf("[party-combat][status-by-id] user=%s battle=%s", user.ID, battle.ID)

	if battle, err = s.refreshMonsterBattleInviteState(ctx, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response, err := s.monsterBattleDetailResponse(ctx, battle)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) endMonsterBattle(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		Outcome string `json:"outcome"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil && !errors.Is(err, io.EOF) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	outcome := strings.TrimSpace(strings.ToLower(requestBody.Outcome))

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	battle, err := s.findActiveMonsterBattleForUser(ctx, user.ID, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no active battle for this monster"})
		return
	}

	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(participants) == 0 {
		participants = []models.MonsterBattleParticipant{{BattleID: battle.ID, UserID: user.ID}}
	}
	log.Printf(
		"[combat][defeat-recovery][end-battle] battle=%s monster=%s participants=%d outcome=%s",
		battle.ID,
		monsterID,
		len(participants),
		outcome,
	)
	for _, participant := range participants {
		recoveryErr := error(nil)
		if outcome == "defeat" {
			recoveryErr = s.applyMonsterBattleDefeatPenalty(ctx, participant.UserID)
		} else {
			recoveryErr = s.restoreUserToOneHealthIfDowned(ctx, participant.UserID)
		}
		if recoveryErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": recoveryErr.Error()})
			return
		}
	}

	if err := s.dbClient.MonsterStatus().DeleteAllForBattleID(ctx, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	endedAt := time.Now()
	if err := s.dbClient.MonsterBattle().End(ctx, battle.ID, endedAt); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battle.EndedAt = &endedAt
	battle.LastActivityAt = endedAt

	ctx.JSON(http.StatusOK, monsterBattleResponseFrom(battle))
}

func (s *server) escapeMonsterBattle(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	battle, err := s.findActiveMonsterBattleForUser(ctx, user.ID, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no active battle for this monster"})
		return
	}
	if battle, err = s.refreshMonsterBattleInviteState(ctx, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle == nil || battle.EndedAt != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "battle not found"})
		return
	}

	if err := s.dbClient.MonsterBattleParticipant().DeleteByBattleAndUser(ctx, battle.ID, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	endedBattle := false
	if len(participants) == 0 {
		if err := s.dbClient.MonsterStatus().DeleteAllForBattleID(ctx, battle.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		endedAt := time.Now()
		if err := s.dbClient.MonsterBattle().End(ctx, battle.ID, endedAt); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		battle.EndedAt = &endedAt
		battle.LastActivityAt = endedAt
		endedBattle = true
	} else {
		anyStandingParticipant := false
		for _, participant := range participants {
			_, _, _, currentHealth, _, err := s.getScenarioResourceState(ctx, participant.UserID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if currentHealth > 0 {
				anyStandingParticipant = true
				break
			}
		}
		if !anyStandingParticipant {
			if err := s.dbClient.MonsterStatus().DeleteAllForBattleID(ctx, battle.ID); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			endedAt := time.Now()
			if err := s.dbClient.MonsterBattle().End(ctx, battle.ID, endedAt); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			battle.EndedAt = &endedAt
			battle.LastActivityAt = endedAt
			endedBattle = true
		}
	}

	if endedBattle {
		log.Printf(
			"[party-combat][escape] battle ended user=%s monster=%s battle=%s",
			user.ID,
			monsterID,
			battle.ID,
		)
		ctx.JSON(http.StatusOK, gin.H{
			"battle":  monsterBattleResponseFrom(battle),
			"ended":   true,
			"message": "escaped battle",
		})
		return
	}

	refreshed, err := s.dbClient.MonsterBattle().FindByID(ctx, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if refreshed == nil || refreshed.EndedAt != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"battle":  monsterBattleResponseFrom(refreshed),
			"ended":   true,
			"message": "escaped battle",
		})
		return
	}
	detail, err := s.monsterBattleDetailResponse(ctx, refreshed)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf(
		"[party-combat][escape] user=%s monster=%s battle=%s remainingParticipants=%d",
		user.ID,
		monsterID,
		battle.ID,
		len(detail.Participants),
	)
	ctx.JSON(http.StatusOK, gin.H{
		"battle":       monsterBattleResponseFrom(refreshed),
		"battleDetail": detail,
		"ended":        false,
		"message":      "escaped battle",
	})
}

func (s *server) applyMonsterBattleDamage(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}
	if _, err := s.dbClient.Monster().FindByID(ctx, monsterID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		Damage         int                         `json:"damage"`
		DamageAffinity *string                     `json:"damageAffinity"`
		Action         *monsterBattleActionRequest `json:"action"`
	}
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.Damage <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "damage must be positive"})
		return
	}
	var damageAffinity *string
	if requestBody.DamageAffinity != nil {
		damageAffinity, err = parseOptionalDamageAffinity(*requestBody.DamageAffinity, "damageAffinity")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	actionRequest := monsterBattleActionRequest{ActionType: "attack"}
	if requestBody.Action != nil {
		actionRequest = *requestBody.Action
	}
	abilityID, err := parseMonsterBattleActionAbilityID(actionRequest.AbilityID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	battle, err := s.findActiveMonsterBattleForUser(ctx, user.ID, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no active battle for this monster"})
		return
	}
	log.Printf(
		"[party-combat][damage-by-monster] user=%s monster=%s battle=%s damage=%d",
		user.ID,
		monsterID,
		battle.ID,
		requestBody.Damage,
	)
	if battle, err = s.refreshMonsterBattleInviteState(ctx, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle.State != string(models.MonsterBattleStateActive) {
		detail, detailErr := s.monsterBattleDetailResponse(ctx, battle)
		if detailErr != nil {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "battle is waiting for party invite responses",
			})
			return
		}
		ctx.JSON(http.StatusConflict, gin.H{
			"error":  "battle is waiting for party invite responses",
			"battle": detail,
		})
		return
	}

	monster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusBonuses, err := s.dbClient.MonsterStatus().GetActiveStatBonuses(ctx, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	attackerBonuses, err := s.getCharacterTotalBonuses(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	damageWithBonus, _, _ := applyAffinityDamageBonus(
		requestBody.Damage,
		damageAffinity,
		attackerBonuses,
	)
	appliedDamage, normalizedAffinity, affinityModifier := applyMonsterAffinityDamage(
		monster,
		damageWithBonus,
		damageAffinity,
		statusBonuses,
	)

	if err := s.dbClient.MonsterBattle().AdjustMonsterHealthDeficit(ctx, battle.ID, appliedDamage); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	if err := s.advanceUserCooldownsForCombatTurn(ctx, user.ID, nil, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.advanceMonsterCooldownsForCombatTurn(ctx, battle, nil, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	userDotDamage, monsterDotDamage, err := s.applyBattleTurnDamageOverTime(ctx, user.ID, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.advanceBattleStatusDurations(ctx, user.ID, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battle, err = s.dbClient.MonsterBattle().FindByID(ctx, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if monsterDotDamage > 0 {
		battle.MonsterHealthDeficit += monsterDotDamage
	}
	if battle, err = s.finalizeMonsterBattleIfDefeated(ctx, battle); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle, err = s.advanceMonsterBattleTurnState(ctx, battle, &user.ID, nil); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.recordMonsterBattleLastAction(
		ctx,
		battle,
		monsterBattleLastActionFromRequest(user, monster, actionRequest, abilityID, appliedDamage),
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battleDetail, err := s.monsterBattleDetailResponse(ctx, battle)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"battle":                     monsterBattleResponseFrom(battle),
		"battleDetail":               battleDetail,
		"baseDamage":                 requestBody.Damage,
		"appliedDamage":              appliedDamage,
		"damageAffinity":             normalizedAffinity,
		"affinityModifier":           affinityModifier,
		"battleTurnUserDotDamage":    userDotDamage,
		"battleTurnMonsterDotDamage": monsterDotDamage,
	})
	log.Printf(
		"[party-combat][damage-by-monster][result] user=%s battle=%s turnIndex=%d deficit=%d userDot=%d monsterDot=%d",
		user.ID,
		battle.ID,
		battle.TurnIndex,
		battle.MonsterHealthDeficit,
		userDotDamage,
		monsterDotDamage,
	)
}

func (s *server) applyMonsterBattleDamageByID(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	battleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid battle ID"})
		return
	}

	var requestBody struct {
		Damage         int                         `json:"damage"`
		DamageAffinity *string                     `json:"damageAffinity"`
		Action         *monsterBattleActionRequest `json:"action"`
	}
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.Damage <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "damage must be positive"})
		return
	}
	var damageAffinity *string
	if requestBody.DamageAffinity != nil {
		damageAffinity, err = parseOptionalDamageAffinity(*requestBody.DamageAffinity, "damageAffinity")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	actionRequest := monsterBattleActionRequest{ActionType: "attack"}
	if requestBody.Action != nil {
		actionRequest = *requestBody.Action
	}
	abilityID, err := parseMonsterBattleActionAbilityID(actionRequest.AbilityID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	battle, err := s.dbClient.MonsterBattle().FindByID(ctx, battleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle == nil || battle.EndedAt != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "battle not found"})
		return
	}
	canAccess, err := s.userCanAccessMonsterBattle(ctx, user.ID, battle)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canAccess {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "not a battle participant"})
		return
	}
	log.Printf(
		"[party-combat][damage-by-id] user=%s battle=%s damage=%d",
		user.ID,
		battle.ID,
		requestBody.Damage,
	)

	if battle, err = s.refreshMonsterBattleInviteState(ctx, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle.State != string(models.MonsterBattleStateActive) {
		detail, detailErr := s.monsterBattleDetailResponse(ctx, battle)
		if detailErr != nil {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "battle is waiting for party invite responses",
			})
			return
		}
		ctx.JSON(http.StatusConflict, gin.H{
			"error":  "battle is waiting for party invite responses",
			"battle": detail,
		})
		return
	}

	monster, err := s.dbClient.Monster().FindByID(ctx, battle.MonsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusBonuses, err := s.dbClient.MonsterStatus().GetActiveStatBonuses(ctx, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	attackerBonuses, err := s.getCharacterTotalBonuses(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	damageWithBonus, _, _ := applyAffinityDamageBonus(
		requestBody.Damage,
		damageAffinity,
		attackerBonuses,
	)
	appliedDamage, normalizedAffinity, affinityModifier := applyMonsterAffinityDamage(
		monster,
		damageWithBonus,
		damageAffinity,
		statusBonuses,
	)

	if err := s.dbClient.MonsterBattle().AdjustMonsterHealthDeficit(ctx, battle.ID, appliedDamage); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	if err := s.advanceUserCooldownsForCombatTurn(ctx, user.ID, nil, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.advanceMonsterCooldownsForCombatTurn(ctx, battle, nil, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	userDotDamage, monsterDotDamage, err := s.applyBattleTurnDamageOverTime(ctx, user.ID, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.advanceBattleStatusDurations(ctx, user.ID, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battle, err = s.dbClient.MonsterBattle().FindByID(ctx, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if monsterDotDamage > 0 {
		battle.MonsterHealthDeficit += monsterDotDamage
	}
	if battle, err = s.finalizeMonsterBattleIfDefeated(ctx, battle); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle, err = s.advanceMonsterBattleTurnState(ctx, battle, &user.ID, nil); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.recordMonsterBattleLastAction(
		ctx,
		battle,
		monsterBattleLastActionFromRequest(user, monster, actionRequest, abilityID, appliedDamage),
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battleDetail, err := s.monsterBattleDetailResponse(ctx, battle)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"battle":                     monsterBattleResponseFrom(battle),
		"battleDetail":               battleDetail,
		"baseDamage":                 requestBody.Damage,
		"appliedDamage":              appliedDamage,
		"damageAffinity":             normalizedAffinity,
		"affinityModifier":           affinityModifier,
		"battleTurnUserDotDamage":    userDotDamage,
		"battleTurnMonsterDotDamage": monsterDotDamage,
	})
	log.Printf(
		"[party-combat][damage-by-id][result] user=%s battle=%s turnIndex=%d deficit=%d userDot=%d monsterDot=%d",
		user.ID,
		battle.ID,
		battle.TurnIndex,
		battle.MonsterHealthDeficit,
		userDotDamage,
		monsterDotDamage,
	)
}

func (s *server) advanceMonsterBattleTurn(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}
	if _, err := s.dbClient.Monster().FindByID(ctx, monsterID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	battle, err := s.findActiveMonsterBattleForUser(ctx, user.ID, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no active battle for this monster"})
		return
	}
	if battle, err = s.refreshMonsterBattleInviteState(ctx, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle.State != string(models.MonsterBattleStateActive) {
		detail, detailErr := s.monsterBattleDetailResponse(ctx, battle)
		if detailErr != nil {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "battle is waiting for party invite responses",
			})
			return
		}
		ctx.JSON(http.StatusConflict, gin.H{
			"error":  "battle is waiting for party invite responses",
			"battle": detail,
		})
		return
	}

	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	turnOrder, err := s.buildMonsterBattleTurnOrder(ctx, battle, participants)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(turnOrder) > 0 {
		currentIndex := battle.TurnIndex
		if currentIndex < 0 || currentIndex >= len(turnOrder) {
			currentIndex = 0
		}
		currentTurn := turnOrder[currentIndex]
		if strings.ToLower(strings.TrimSpace(currentTurn.EntityType)) != "monster" {
			ctx.JSON(http.StatusConflict, gin.H{"error": "it is not the monster's turn"})
			return
		}
	}

	now := time.Now()
	if err := s.dbClient.MonsterBattle().Touch(ctx, battle.ID, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battle.LastActivityAt = now
	if err := s.advanceUserCooldownsForCombatTurn(ctx, user.ID, nil, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.advanceMonsterCooldownsForCombatTurn(ctx, battle, nil, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userDotDamage, monsterDotDamage, err := s.applyBattleTurnDamageOverTime(ctx, user.ID, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if monsterDotDamage > 0 {
		battle.MonsterHealthDeficit += monsterDotDamage
	}
	if err := s.advanceBattleStatusDurations(ctx, user.ID, battle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedMonster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battleMonster, err := s.monsterScaledForBattle(ctx, battle, updatedMonster)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	monsterAction, participantResources, err := s.executeMonsterBattleAction(ctx, battle, battleMonster)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battle, err = s.dbClient.MonsterBattle().FindByID(ctx, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if battle, err = s.finalizeMonsterBattleIfDefeated(ctx, battle); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle, err = s.advanceMonsterBattleTurnState(ctx, battle, nil, &battle.MonsterID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if monsterAction != nil {
		if err := s.recordMonsterBattleLastAction(
			ctx,
			battle,
			monsterBattleLastActionFromMonsterAction(monsterAction),
		); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	updatedMonster, err = s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	monsterResponse, err := s.buildMonsterResponse(ctx, user.ID, updatedMonster)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, userMaxHealth, _, userHealth, _, err := s.getScenarioResourceState(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battleDetail, err := s.monsterBattleDetailResponse(ctx, battle)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"battle":               monsterBattleResponseFrom(battle),
		"battleDetail":         battleDetail,
		"monster":              monsterResponse,
		"userDotDamage":        userDotDamage,
		"monsterDotDamage":     monsterDotDamage,
		"userHealth":           userHealth,
		"userMaxHealth":        userMaxHealth,
		"monsterAction":        monsterAction,
		"participantResources": participantResources,
	})
	log.Printf(
		"[party-combat][turn][result] user=%s battle=%s turnIndex=%d deficit=%d userDot=%d monsterDot=%d userHealth=%d",
		user.ID,
		battle.ID,
		battle.TurnIndex,
		battle.MonsterHealthDeficit,
		userDotDamage,
		monsterDotDamage,
		userHealth,
	)
}

func (s *server) createMonster(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody monsterUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	monster, itemRewards, err := s.parseMonsterUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if monster.ImageURL != "" {
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusComplete
		emptyError := ""
		monster.ImageGenerationError = &emptyError
	}

	if err := s.dbClient.Monster().Create(ctx, monster); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Monster().ReplaceItemRewards(ctx, monster.ID, itemRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.dbClient.Monster().FindByID(ctx, monster.ID)
	if err != nil {
		response, responseErr := s.buildMonsterResponse(ctx, user.ID, monster)
		if responseErr != nil {
			ctx.JSON(http.StatusCreated, monster)
			return
		}
		ctx.JSON(http.StatusCreated, response)
		return
	}
	response, err := s.buildMonsterResponse(ctx, user.ID, created)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

func (s *server) updateMonster(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	existingMonster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody monsterUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	monster, itemRewards, err := s.parseMonsterUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if monster.ImageURL != "" {
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusComplete
		emptyError := ""
		monster.ImageGenerationError = &emptyError
	} else if existingMonster.ImageGenerationStatus == models.MonsterImageGenerationStatusQueued ||
		existingMonster.ImageGenerationStatus == models.MonsterImageGenerationStatusInProgress {
		monster.ImageGenerationStatus = existingMonster.ImageGenerationStatus
		monster.ImageGenerationError = existingMonster.ImageGenerationError
	} else {
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusNone
		emptyError := ""
		monster.ImageGenerationError = &emptyError
	}

	if err := s.dbClient.Monster().Update(ctx, monsterID, monster); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Monster().ReplaceItemRewards(ctx, monsterID, itemRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"id": monsterID})
		return
	}
	response, err := s.buildMonsterResponse(ctx, user.ID, updated)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) deleteMonster(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	if _, err := s.dbClient.Monster().FindByID(ctx, monsterID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Monster().Delete(ctx, monsterID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "monster deleted successfully"})
}

func (s *server) generateMonsterImage(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	monster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	monster.ImageGenerationStatus = models.MonsterImageGenerationStatusQueued
	emptyError := ""
	monster.ImageGenerationError = &emptyError
	if err := s.dbClient.Monster().Update(ctx, monsterID, monster); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue monster image generation: " + err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.GenerateMonsterImageTaskPayload{MonsterID: monsterID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build monster image generation payload"})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateMonsterImageTaskType, payloadBytes)); err != nil {
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusFailed
		errMessage := err.Error()
		monster.ImageGenerationError = &errMessage
		_ = s.dbClient.Monster().Update(ctx, monsterID, monster)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue monster image generation: " + err.Error()})
		return
	}

	updatedMonster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		response, responseErr := s.buildMonsterResponse(ctx, user.ID, monster)
		if responseErr != nil {
			ctx.JSON(http.StatusOK, monster)
			return
		}
		ctx.JSON(http.StatusOK, response)
		return
	}
	response, err := s.buildMonsterResponse(ctx, user.ID, updatedMonster)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}
