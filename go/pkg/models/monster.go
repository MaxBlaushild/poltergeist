package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Monster struct {
	ID                          uuid.UUID           `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt                   time.Time           `json:"createdAt"`
	UpdatedAt                   time.Time           `json:"updatedAt"`
	Name                        string              `json:"name"`
	Description                 string              `json:"description"`
	ImageURL                    string              `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL                string              `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	OwnerUserID                 *uuid.UUID          `json:"ownerUserId,omitempty" gorm:"column:owner_user_id;type:uuid"`
	OwnerUser                   *User               `json:"ownerUser,omitempty" gorm:"foreignKey:OwnerUserID"`
	Ephemeral                   bool                `json:"ephemeral" gorm:"column:ephemeral"`
	ZoneID                      uuid.UUID           `json:"zoneId" gorm:"column:zone_id"`
	Zone                        Zone                `json:"zone"`
	Latitude                    float64             `json:"latitude"`
	Longitude                   float64             `json:"longitude"`
	Geometry                    string              `json:"geometry" gorm:"type:geometry(Point,4326)"`
	TemplateID                  *uuid.UUID          `json:"templateId,omitempty" gorm:"column:template_id"`
	Template                    *MonsterTemplate    `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	DominantHandInventoryItemID *int                `json:"dominantHandInventoryItemId,omitempty" gorm:"column:dominant_hand_inventory_item_id"`
	DominantHandInventoryItem   *InventoryItem      `json:"dominantHandInventoryItem,omitempty" gorm:"foreignKey:DominantHandInventoryItemID"`
	OffHandInventoryItemID      *int                `json:"offHandInventoryItemId,omitempty" gorm:"column:off_hand_inventory_item_id"`
	OffHandInventoryItem        *InventoryItem      `json:"offHandInventoryItem,omitempty" gorm:"foreignKey:OffHandInventoryItemID"`
	WeaponInventoryItemID       *int                `json:"weaponInventoryItemId,omitempty" gorm:"column:weapon_inventory_item_id"`
	WeaponInventoryItem         *InventoryItem      `json:"weaponInventoryItem,omitempty" gorm:"foreignKey:WeaponInventoryItemID"`
	Level                       int                 `json:"level"`
	RewardMode                  RewardMode          `json:"rewardMode" gorm:"column:reward_mode"`
	RandomRewardSize            RandomRewardSize    `json:"randomRewardSize" gorm:"column:random_reward_size"`
	RewardExperience            int                 `json:"rewardExperience" gorm:"column:reward_experience"`
	RewardGold                  int                 `json:"rewardGold" gorm:"column:reward_gold"`
	ImageGenerationStatus       string              `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError        *string             `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
	ItemRewards                 []MonsterItemReward `json:"itemRewards" gorm:"foreignKey:MonsterID"`
}

type MonsterTemplate struct {
	ID                    uuid.UUID              `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt             time.Time              `json:"createdAt"`
	UpdatedAt             time.Time              `json:"updatedAt"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	ImageURL              string                 `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL          string                 `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	BaseStrength          int                    `json:"baseStrength" gorm:"column:base_strength"`
	BaseDexterity         int                    `json:"baseDexterity" gorm:"column:base_dexterity"`
	BaseConstitution      int                    `json:"baseConstitution" gorm:"column:base_constitution"`
	BaseIntelligence      int                    `json:"baseIntelligence" gorm:"column:base_intelligence"`
	BaseWisdom            int                    `json:"baseWisdom" gorm:"column:base_wisdom"`
	BaseCharisma          int                    `json:"baseCharisma" gorm:"column:base_charisma"`
	ImageGenerationStatus string                 `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError  *string                `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
	Spells                []MonsterTemplateSpell `json:"spells" gorm:"foreignKey:MonsterTemplateID"`
}

type MonsterTemplateSpell struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	MonsterTemplateID uuid.UUID `json:"monsterTemplateId" gorm:"column:monster_template_id"`
	SpellID           uuid.UUID `json:"spellId" gorm:"column:spell_id"`
	Spell             Spell     `json:"spell"`
}

type MonsterItemReward struct {
	ID              uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	MonsterID       uuid.UUID     `json:"monsterId" gorm:"column:monster_id"`
	InventoryItemID int           `json:"inventoryItemId" gorm:"column:inventory_item_id"`
	InventoryItem   InventoryItem `json:"inventoryItem"`
	Quantity        int           `json:"quantity"`
}

func (m *Monster) TableName() string {
	return "monsters"
}

func (m *MonsterTemplate) TableName() string {
	return "monster_templates"
}

func (m *MonsterTemplateSpell) TableName() string {
	return "monster_template_spells"
}

func (m *MonsterItemReward) TableName() string {
	return "monster_item_rewards"
}

func (m *Monster) BeforeSave(tx *gorm.DB) error {
	if err := m.SetGeometry(m.Latitude, m.Longitude); err != nil {
		return err
	}
	return nil
}

func (m *Monster) SetGeometry(latitude, longitude float64) error {
	m.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}

func (m *Monster) EffectiveLevel() int {
	if m.Level < 1 {
		return 1
	}
	return m.Level
}

func (m *Monster) EffectiveStats() CharacterStatBonuses {
	return m.EffectiveStatsWithBonuses(CharacterStatBonuses{})
}

func (m *Monster) EffectiveStatsWithBonuses(bonuses CharacterStatBonuses) CharacterStatBonuses {
	levelBonus := m.EffectiveLevel() - 1
	baseStrength := 10
	baseDexterity := 10
	baseConstitution := 10
	baseIntelligence := 10
	baseWisdom := 10
	baseCharisma := 10
	if m.Template != nil {
		baseStrength = maxInt(1, m.Template.BaseStrength)
		baseDexterity = maxInt(1, m.Template.BaseDexterity)
		baseConstitution = maxInt(1, m.Template.BaseConstitution)
		baseIntelligence = maxInt(1, m.Template.BaseIntelligence)
		baseWisdom = maxInt(1, m.Template.BaseWisdom)
		baseCharisma = maxInt(1, m.Template.BaseCharisma)
	}
	return CharacterStatBonuses{
		Strength:     baseStrength + levelBonus,
		Dexterity:    baseDexterity + levelBonus,
		Constitution: baseConstitution + levelBonus,
		Intelligence: baseIntelligence + levelBonus,
		Wisdom:       baseWisdom + levelBonus,
		Charisma:     baseCharisma + levelBonus,
	}.Add(bonuses)
}

func (m *Monster) DerivedMaxHealth() int {
	return m.DerivedMaxHealthWithBonuses(CharacterStatBonuses{})
}

func (m *Monster) DerivedMaxHealthWithBonuses(bonuses CharacterStatBonuses) int {
	stats := m.EffectiveStatsWithBonuses(bonuses)
	return maxInt(1, stats.Constitution) * 10
}

func (m *Monster) DerivedMaxMana() int {
	return m.DerivedMaxManaWithBonuses(CharacterStatBonuses{})
}

func (m *Monster) DerivedMaxManaWithBonuses(bonuses CharacterStatBonuses) int {
	stats := m.EffectiveStatsWithBonuses(bonuses)
	mental := maxInt(1, stats.Intelligence+stats.Wisdom)
	return mental * 5
}

func (m *Monster) DerivedAttackProfile() (damageMin int, damageMax int, swipesPerAttack int) {
	return m.DerivedAttackProfileWithBonuses(CharacterStatBonuses{})
}

func (m *Monster) DerivedAttackProfileWithBonuses(
	bonuses CharacterStatBonuses,
) (damageMin int, damageMax int, swipesPerAttack int) {
	stats := m.EffectiveStatsWithBonuses(bonuses)
	strengthBonus := maxInt(0, (stats.Strength-10)/2)
	dexterityBonus := maxInt(0, (stats.Dexterity-10)/4)
	totalBonus := strengthBonus + dexterityBonus

	dominantWeapon := m.DominantHandInventoryItem
	if dominantWeapon == nil {
		dominantWeapon = m.WeaponInventoryItem
	}

	if dominantWeapon != nil &&
		dominantWeapon.DamageMin != nil &&
		dominantWeapon.DamageMax != nil {
		damageMin = maxInt(1, *dominantWeapon.DamageMin+totalBonus)
		damageMax = maxInt(damageMin, *dominantWeapon.DamageMax+totalBonus)
		swipesPerAttack = 1
		if dominantWeapon.SwipesPerAttack != nil && *dominantWeapon.SwipesPerAttack > 0 {
			swipesPerAttack = *dominantWeapon.SwipesPerAttack
		}
		if m.OffHandInventoryItem != nil &&
			m.OffHandInventoryItem.DamageMin != nil &&
			m.OffHandInventoryItem.DamageMax != nil {
			damageMin += maxInt(1, *m.OffHandInventoryItem.DamageMin)
			damageMax += maxInt(1, *m.OffHandInventoryItem.DamageMax)
			if m.OffHandInventoryItem.SwipesPerAttack != nil && *m.OffHandInventoryItem.SwipesPerAttack > 0 {
				swipesPerAttack += *m.OffHandInventoryItem.SwipesPerAttack
			} else {
				swipesPerAttack += 1
			}
		}
		return damageMin, damageMax, maxInt(1, swipesPerAttack)
	}

	damageMin = maxInt(1, stats.Strength/3+m.EffectiveLevel()/2)
	damageMax = maxInt(damageMin, damageMin+2+stats.Dexterity/5)
	swipesPerAttack = 1
	return damageMin, damageMax, swipesPerAttack
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

const (
	MonsterImageGenerationStatusNone       = "none"
	MonsterImageGenerationStatusQueued     = "queued"
	MonsterImageGenerationStatusInProgress = "in_progress"
	MonsterImageGenerationStatusComplete   = "complete"
	MonsterImageGenerationStatusFailed     = "failed"
)

const (
	MonsterTemplateImageGenerationStatusNone       = "none"
	MonsterTemplateImageGenerationStatusQueued     = "queued"
	MonsterTemplateImageGenerationStatusInProgress = "in_progress"
	MonsterTemplateImageGenerationStatusComplete   = "complete"
	MonsterTemplateImageGenerationStatusFailed     = "failed"
)
