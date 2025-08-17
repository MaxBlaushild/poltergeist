package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type InventoryItem struct {
	ID               uuid.UUID       `gorm:"primary_key" json:"id"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
	Name             string          `json:"name"`
	ImageURL         string          `gorm:"column:image_url" json:"imageUrl"`
	FlavorText       string          `gorm:"column:flavor_text" json:"flavorText"`
	EffectText       string          `gorm:"column:effect_text" json:"effectText"`
	RarityTier       string          `gorm:"column:rarity_tier;type:rarity_tier_type" json:"rarityTier"`
	EquipmentSlot    *string         `gorm:"column:equipment_slot;type:equipment_slot_type" json:"equipmentSlot,omitempty"`
	MinDamage        int             `gorm:"column:min_damage" json:"minDamage"`
	MaxDamage        int             `gorm:"column:max_damage" json:"maxDamage"`
	Defense          int             `gorm:"column:defense" json:"defense"`
	Health           int             `gorm:"column:health" json:"health"`
	Speed            int             `gorm:"column:speed" json:"speed"`
	CritChance       int             `gorm:"column:crit_chance" json:"critChance"`
	CritDamage       int             `gorm:"column:crit_damage" json:"critDamage"`
	AttackRange      int             `gorm:"column:attack_range" json:"attackRange"`
	DamageType       string          `gorm:"column:damage_type" json:"damageType"`
	PlusStrength     int             `gorm:"column:plus_strength" json:"plusStrength"`
	PlusAgility      int             `gorm:"column:plus_agility" json:"plusAgility"`
	PlusIntelligence int             `gorm:"column:plus_intelligence" json:"plusIntelligence"`
	PlusWisdom       int             `gorm:"column:plus_wisdom" json:"plusWisdom"`
	PlusConstitution int             `gorm:"column:plus_constitution" json:"plusConstitution"`
	PlusCharisma     int             `gorm:"column:plus_charisma" json:"plusCharisma"`
	PermanentID      string          `gorm:"column:permanant_identifier" json:"permanentId"`
	Weight           float64         `gorm:"column:weight" json:"weight"`
	Value            int             `gorm:"column:value" json:"value"`
	Durability       int             `gorm:"column:durability" json:"durability"`
	MaxDurability    int             `gorm:"column:max_durability" json:"maxDurability"`
	LevelRequirement int             `gorm:"column:level_requirement" json:"levelRequirement"`
	Stackable        bool            `gorm:"column:stackable" json:"stackable"`
	MaxStackSize     int             `gorm:"column:max_stack_size" json:"maxStackSize"`
	Tradeable        bool            `gorm:"column:tradeable" json:"tradeable"`
	Cooldown         int             `gorm:"column:cooldown" json:"cooldown"`
	Charges          int             `gorm:"column:charges" json:"charges"`
	MaxCharges       int             `gorm:"column:max_charges" json:"maxCharges"`
	QuestRelated     bool            `gorm:"column:quest_related" json:"questRelated"`
	CraftIngredients json.RawMessage `gorm:"column:crafting_ingredients" json:"craftingIngredients"`
	SpecialAbilities json.RawMessage `gorm:"column:special_abilities" json:"specialAbilities"`
	ItemColor        string          `gorm:"column:item_color" json:"itemColor"`
	AnimationEffects string          `gorm:"column:animation_effects" json:"animationEffects"`
	SoundEffects     string          `gorm:"column:sound_effects" json:"soundEffects"`
	BonusStats       json.RawMessage `gorm:"column:bonus_stats" json:"bonusStats"`
}

func (InventoryItem) TableName() string {
	return "inventory_items"
}
