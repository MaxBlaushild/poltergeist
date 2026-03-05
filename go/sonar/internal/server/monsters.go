package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

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
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	ImageURL         string   `json:"imageUrl"`
	ThumbnailURL     string   `json:"thumbnailUrl"`
	BaseStrength     int      `json:"baseStrength"`
	BaseDexterity    int      `json:"baseDexterity"`
	BaseConstitution int      `json:"baseConstitution"`
	BaseIntelligence int      `json:"baseIntelligence"`
	BaseWisdom       int      `json:"baseWisdom"`
	BaseCharisma     int      `json:"baseCharisma"`
	SpellIDs         []string `json:"spellIds"`
}

type bulkGenerateMonsterTemplatesRequest struct {
	Count int `json:"count"`
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
You are designing %d original monster templates for a fantasy action RPG.

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

type monsterRewardItemPayload struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type monsterUpsertRequest struct {
	Name                        string                     `json:"name"`
	Description                 string                     `json:"description"`
	ImageURL                    string                     `json:"imageUrl"`
	ThumbnailURL                string                     `json:"thumbnailUrl"`
	ZoneID                      string                     `json:"zoneId"`
	Latitude                    float64                    `json:"latitude"`
	Longitude                   float64                    `json:"longitude"`
	TemplateID                  string                     `json:"templateId"`
	DominantHandInventoryItemID *int                       `json:"dominantHandInventoryItemId"`
	OffHandInventoryItemID      *int                       `json:"offHandInventoryItemId"`
	WeaponInventoryItemID       *int                       `json:"weaponInventoryItemId"`
	Level                       int                        `json:"level"`
	RewardMode                  string                     `json:"rewardMode"`
	RandomRewardSize            string                     `json:"randomRewardSize"`
	RewardExperience            int                        `json:"rewardExperience"`
	RewardGold                  int                        `json:"rewardGold"`
	ItemRewards                 []monsterRewardItemPayload `json:"itemRewards"`
}

type monsterEncounterUpsertRequest struct {
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	ImageURL            string   `json:"imageUrl"`
	ThumbnailURL        string   `json:"thumbnailUrl"`
	ScaleWithUserLevel  bool     `json:"scaleWithUserLevel"`
	RecurrenceFrequency *string  `json:"recurrenceFrequency"`
	ZoneID              string   `json:"zoneId"`
	Latitude            float64  `json:"latitude"`
	Longitude           float64  `json:"longitude"`
	MonsterIDs          []string `json:"monsterIds"`
}

type monsterTemplateResponse struct {
	ID                    uuid.UUID      `json:"id"`
	CreatedAt             time.Time      `json:"createdAt"`
	UpdatedAt             time.Time      `json:"updatedAt"`
	Name                  string         `json:"name"`
	Description           string         `json:"description"`
	ImageURL              string         `json:"imageUrl"`
	ThumbnailURL          string         `json:"thumbnailUrl"`
	BaseStrength          int            `json:"baseStrength"`
	BaseDexterity         int            `json:"baseDexterity"`
	BaseConstitution      int            `json:"baseConstitution"`
	BaseIntelligence      int            `json:"baseIntelligence"`
	BaseWisdom            int            `json:"baseWisdom"`
	BaseCharisma          int            `json:"baseCharisma"`
	Spells                []models.Spell `json:"spells"`
	ImageGenerationStatus string         `json:"imageGenerationStatus"`
	ImageGenerationError  *string        `json:"imageGenerationError,omitempty"`
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
	Spells                      []models.Spell             `json:"spells"`
	Statuses                    []models.MonsterStatus     `json:"statuses"`
	ActiveBattleID              *uuid.UUID                 `json:"activeBattleId,omitempty"`
	RewardMode                  models.RewardMode          `json:"rewardMode"`
	RandomRewardSize            models.RandomRewardSize    `json:"randomRewardSize"`
	RewardExperience            int                        `json:"rewardExperience"`
	RewardGold                  int                        `json:"rewardGold"`
	ItemRewards                 []models.MonsterItemReward `json:"itemRewards"`
	ImageGenerationStatus       string                     `json:"imageGenerationStatus"`
	ImageGenerationError        *string                    `json:"imageGenerationError,omitempty"`
}

type monsterBattleResponse struct {
	ID                   uuid.UUID  `json:"id"`
	UserID               uuid.UUID  `json:"userId"`
	MonsterID            uuid.UUID  `json:"monsterId"`
	StartedAt            time.Time  `json:"startedAt"`
	LastActivityAt       time.Time  `json:"lastActivityAt"`
	MonsterHealthDeficit int        `json:"monsterHealthDeficit"`
	EndedAt              *time.Time `json:"endedAt,omitempty"`
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
	ScaleWithUserLevel          bool                             `json:"scaleWithUserLevel"`
	RecurringMonsterEncounterID *uuid.UUID                       `json:"recurringMonsterEncounterId,omitempty"`
	RecurrenceFrequency         *string                          `json:"recurrenceFrequency,omitempty"`
	NextRecurrenceAt            *time.Time                       `json:"nextRecurrenceAt,omitempty"`
	ZoneID                      uuid.UUID                        `json:"zoneId"`
	Zone                        models.Zone                      `json:"zone"`
	Latitude                    float64                          `json:"latitude"`
	Longitude                   float64                          `json:"longitude"`
	MonsterCount                int                              `json:"monsterCount"`
	Members                     []monsterEncounterMemberResponse `json:"members"`
	Monsters                    []monsterResponse                `json:"monsters"`
}

func monsterBattleResponseFrom(battle *models.MonsterBattle) *monsterBattleResponse {
	if battle == nil {
		return nil
	}
	return &monsterBattleResponse{
		ID:                   battle.ID,
		UserID:               battle.UserID,
		MonsterID:            battle.MonsterID,
		StartedAt:            battle.StartedAt,
		LastActivityAt:       battle.LastActivityAt,
		MonsterHealthDeficit: battle.MonsterHealthDeficit,
		EndedAt:              battle.EndedAt,
	}
}

func (s *server) monsterEncounterResponseFrom(
	ctx context.Context,
	userID uuid.UUID,
	encounter *models.MonsterEncounter,
	userLevel int,
	applyLevelScaling bool,
) (monsterEncounterResponse, error) {
	members := make([]monsterEncounterMemberResponse, 0, len(encounter.Members))
	monsters := make([]monsterResponse, 0, len(encounter.Members))
	for i := range encounter.Members {
		member := encounter.Members[i]
		monster := member.Monster
		if applyLevelScaling && encounter.ScaleWithUserLevel {
			monster.Level = normalizeScaledLevel(userLevel)
		}
		entry, err := s.buildMonsterResponse(ctx, userID, &monster)
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
		ScaleWithUserLevel:          encounter.ScaleWithUserLevel,
		RecurringMonsterEncounterID: encounter.RecurringMonsterEncounterID,
		RecurrenceFrequency:         encounter.RecurrenceFrequency,
		NextRecurrenceAt:            encounter.NextRecurrenceAt,
		ZoneID:                      encounter.ZoneID,
		Zone:                        encounter.Zone,
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
	spells := make([]models.Spell, 0, len(template.Spells))
	for _, templateSpell := range template.Spells {
		if templateSpell.Spell.ID == uuid.Nil {
			continue
		}
		spells = append(spells, templateSpell.Spell)
	}
	return &monsterTemplateResponse{
		ID:                    template.ID,
		CreatedAt:             template.CreatedAt,
		UpdatedAt:             template.UpdatedAt,
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
		Spells:                spells,
		ImageGenerationStatus: template.ImageGenerationStatus,
		ImageGenerationError:  template.ImageGenerationError,
	}
}

func monsterResponseFrom(
	monster *models.Monster,
	statusBonuses models.CharacterStatBonuses,
	activeStatuses []models.MonsterStatus,
	activeBattle *models.MonsterBattle,
) monsterResponse {
	stats := monster.EffectiveStatsWithBonuses(statusBonuses)
	maxHealth := monster.DerivedMaxHealthWithBonuses(statusBonuses)
	maxMana := monster.DerivedMaxManaWithBonuses(statusBonuses)
	damageMin, damageMax, swipes := monster.DerivedAttackProfileWithBonuses(statusBonuses)
	currentHealth := maxHealth
	if activeBattle != nil {
		currentHealth = maxHealth - activeBattle.MonsterHealthDeficit
		if currentHealth < 0 {
			currentHealth = 0
		}
	}
	spells := []models.Spell{}
	if monster.Template != nil {
		for _, templateSpell := range monster.Template.Spells {
			if templateSpell.Spell.ID == uuid.Nil {
				continue
			}
			spells = append(spells, templateSpell.Spell)
		}
	}
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
		Mana:                        maxMana,
		MaxMana:                     maxMana,
		AttackDamageMin:             damageMin,
		AttackDamageMax:             damageMax,
		AttackSwipesPerAttack:       swipes,
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
		return activeBattle, nil
	}

	now := time.Now()
	battle := &models.MonsterBattle{
		UserID:         userID,
		MonsterID:      monsterID,
		StartedAt:      now,
		LastActivityAt: now,
	}
	if err := s.dbClient.MonsterBattle().Create(ctx, battle); err != nil {
		return nil, err
	}
	return battle, nil
}

func (s *server) buildMonsterResponse(
	ctx context.Context,
	userID uuid.UUID,
	monster *models.Monster,
) (monsterResponse, error) {
	activeBattle, err := s.dbClient.MonsterBattle().FindActiveByUserAndMonster(ctx, userID, monster.ID)
	if err != nil {
		return monsterResponse{}, err
	}
	if activeBattle == nil {
		response := monsterResponseFrom(monster, models.CharacterStatBonuses{}, []models.MonsterStatus{}, nil)
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
	response := monsterResponseFrom(monster, totalStatusBonuses, activeStatuses, activeBattle)
	if err := s.applyMonsterRewardsForUser(ctx, userID, monster, &response); err != nil {
		return monsterResponse{}, err
	}
	return response, nil
}

func (s *server) parseMonsterTemplateUpsertRequest(
	ctx context.Context,
	body monsterTemplateUpsertRequest,
) (*models.MonsterTemplate, []models.MonsterTemplateSpell, error) {
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if body.BaseStrength < 1 ||
		body.BaseDexterity < 1 ||
		body.BaseConstitution < 1 ||
		body.BaseIntelligence < 1 ||
		body.BaseWisdom < 1 ||
		body.BaseCharisma < 1 {
		return nil, nil, fmt.Errorf("all base stats must be positive")
	}

	spells := []models.MonsterTemplateSpell{}
	seenSpellIDs := map[uuid.UUID]bool{}
	for index, spellIDString := range body.SpellIDs {
		spellID, err := uuid.Parse(strings.TrimSpace(spellIDString))
		if err != nil {
			return nil, nil, fmt.Errorf("spellIds[%d] must be a valid UUID", index)
		}
		if seenSpellIDs[spellID] {
			continue
		}
		seenSpellIDs[spellID] = true
		if _, err := s.dbClient.Spell().FindByID(ctx, spellID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fmt.Errorf("spellIds[%d] not found", index)
			}
			return nil, nil, err
		}
		spells = append(spells, models.MonsterTemplateSpell{SpellID: spellID})
	}

	template := &models.MonsterTemplate{
		Name:             name,
		Description:      strings.TrimSpace(body.Description),
		ImageURL:         strings.TrimSpace(body.ImageURL),
		ThumbnailURL:     strings.TrimSpace(body.ThumbnailURL),
		BaseStrength:     body.BaseStrength,
		BaseDexterity:    body.BaseDexterity,
		BaseConstitution: body.BaseConstitution,
		BaseIntelligence: body.BaseIntelligence,
		BaseWisdom:       body.BaseWisdom,
		BaseCharisma:     body.BaseCharisma,
	}
	if template.ThumbnailURL == "" && template.ImageURL != "" {
		template.ThumbnailURL = template.ImageURL
	}
	return template, spells, nil
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
	rewardMode := models.NormalizeRewardMode(body.RewardMode)
	if strings.TrimSpace(body.RewardMode) == "" {
		if body.RewardExperience > 0 || body.RewardGold > 0 || len(body.ItemRewards) > 0 {
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

	if len(body.MonsterIDs) < 1 || len(body.MonsterIDs) > 9 {
		return nil, nil, fmt.Errorf("monsterIds must include between 1 and 9 monsters")
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

	name := strings.TrimSpace(body.Name)
	if name == "" {
		if len(resolvedMonsters) == 1 {
			name = fmt.Sprintf("%s Encounter", strings.TrimSpace(resolvedMonsters[0].Name))
		} else {
			name = fmt.Sprintf("%d-Monster Encounter", len(resolvedMonsters))
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
		ScaleWithUserLevel: body.ScaleWithUserLevel,
		ZoneID:             zoneID,
		Latitude:           body.Latitude,
		Longitude:          body.Longitude,
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

	template, spells, err := s.parseMonsterTemplateUpsertRequest(ctx, requestBody)
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

func buildBulkMonsterTemplateSpecsFromSeeds(count int, usedNames map[string]struct{}) []jobs.MonsterTemplateCreationSpec {
	specs := make([]jobs.MonsterTemplateCreationSpec, 0, count)
	if count <= 0 || len(dndMonsterTemplateSeeds) == 0 {
		return specs
	}
	for i := 0; i < count; i++ {
		seed := dndMonsterTemplateSeeds[i%len(dndMonsterTemplateSeeds)]
		specs = append(specs, jobs.MonsterTemplateCreationSpec{
			Name:             nextUniqueMonsterTemplateName(seed.Name, usedNames),
			Description:      strings.TrimSpace(seed.Description),
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
) ([]jobs.MonsterTemplateCreationSpec, string, error) {
	if count <= 0 {
		return []jobs.MonsterTemplateCreationSpec{}, "none", nil
	}

	specs := make([]jobs.MonsterTemplateCreationSpec, 0, count)
	source := "seed_generated"

	if s.deepPriest != nil {
		aiSpecs, err := s.generateMonsterTemplateSpecsWithLLM(count, usedNames, existingNames)
		if err == nil && len(aiSpecs) > 0 {
			specs = append(specs, aiSpecs...)
			source = "ai_generated"
		}
	}

	if remaining := count - len(specs); remaining > 0 {
		fallback := buildBulkMonsterTemplateSpecsFromSeeds(remaining, usedNames)
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
		prompt := fmt.Sprintf(
			generateMonsterTemplatesPromptTemplate,
			remaining,
			formatMonsterTemplateNamesForPrompt(denyList),
			remaining,
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

	templateSpecs, source, err := s.buildBulkMonsterTemplateSpecs(requestBody.Count, usedNames, existingNames)
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
		JobID:      jobID,
		Source:     source,
		TotalCount: len(templateSpecs),
		Templates:  templateSpecs,
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

	template, spells, err := s.parseMonsterTemplateUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
		entry, err := s.buildMonsterResponse(ctx, user.ID, &monsters[i])
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response = append(response, entry)
	}
	ctx.JSON(http.StatusOK, response)
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
	response := make([]monsterEncounterResponse, 0, len(encounters))
	for i := range encounters {
		entry, err := s.monsterEncounterResponseFrom(ctx, user.ID, &encounters[i], 1, false)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response = append(response, entry)
	}
	ctx.JSON(http.StatusOK, response)
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
	userLevel, err := s.currentUserLevel(ctx, user.ID)
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
	userLevel, err := s.currentUserLevel(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]monsterEncounterResponse, 0, len(encounters))
	for i := range encounters {
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
	if _, err := s.dbClient.Monster().FindByID(ctx, monsterID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	battle, err := s.getOrCreateActiveMonsterBattle(ctx, user.ID, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	if err := s.dbClient.MonsterBattle().Touch(ctx, battle.ID, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battle.LastActivityAt = now

	ctx.JSON(http.StatusOK, monsterBattleResponseFrom(battle))
}

func (s *server) endMonsterBattle(ctx *gin.Context) {
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

	battle, err := s.dbClient.MonsterBattle().FindActiveByUserAndMonster(ctx, user.ID, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no active battle for this monster"})
		return
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

	battle, err := s.dbClient.MonsterBattle().FindActiveByUserAndMonster(ctx, user.ID, monsterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if battle == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no active battle for this monster"})
		return
	}

	now := time.Now()
	if err := s.dbClient.MonsterBattle().Touch(ctx, battle.ID, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	battle.LastActivityAt = now

	userDotDamage, monsterDotDamage, err := s.applyBattleTurnDamageOverTime(ctx, user.ID, battle.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if monsterDotDamage > 0 {
		battle.MonsterHealthDeficit += monsterDotDamage
	}

	updatedMonster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
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

	ctx.JSON(http.StatusOK, gin.H{
		"battle":           monsterBattleResponseFrom(battle),
		"monster":          monsterResponse,
		"userDotDamage":    userDotDamage,
		"monsterDotDamage": monsterDotDamage,
		"userHealth":       userHealth,
		"userMaxHealth":    userMaxHealth,
	})
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
