package models

import "time"

type InventoryItem struct {
	ID                                       int                            `json:"id" gorm:"primaryKey"`
	CreatedAt                                time.Time                      `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt                                time.Time                      `json:"updatedAt" gorm:"column:updated_at"`
	Archived                                 bool                           `json:"archived" gorm:"column:archived;default:false"`
	Name                                     string                         `json:"name" gorm:"column:name"`
	ImageURL                                 string                         `json:"imageUrl" gorm:"column:image_url"`
	FlavorText                               string                         `json:"flavorText" gorm:"column:flavor_text"`
	EffectText                               string                         `json:"effectText" gorm:"column:effect_text"`
	RarityTier                               string                         `json:"rarityTier" gorm:"column:rarity_tier"`
	IsCaptureType                            bool                           `json:"isCaptureType" gorm:"column:is_capture_type"`
	BuyPrice                                 *int                           `json:"buyPrice" gorm:"column:buy_price"`
	UnlockTier                               *int                           `json:"unlockTier" gorm:"column:unlock_tier"`
	UnlockLocksStrength                      *int                           `json:"unlockLocksStrength" gorm:"column:unlock_locks_strength"`
	ItemLevel                                int                            `json:"itemLevel" gorm:"column:item_level"`
	EquipSlot                                *string                        `json:"equipSlot" gorm:"column:equip_slot"`
	StrengthMod                              int                            `json:"strengthMod" gorm:"column:strength_mod"`
	DexterityMod                             int                            `json:"dexterityMod" gorm:"column:dexterity_mod"`
	ConstitutionMod                          int                            `json:"constitutionMod" gorm:"column:constitution_mod"`
	IntelligenceMod                          int                            `json:"intelligenceMod" gorm:"column:intelligence_mod"`
	WisdomMod                                int                            `json:"wisdomMod" gorm:"column:wisdom_mod"`
	CharismaMod                              int                            `json:"charismaMod" gorm:"column:charisma_mod"`
	HandItemCategory                         *string                        `json:"handItemCategory" gorm:"column:hand_item_category"`
	Handedness                               *string                        `json:"handedness" gorm:"column:handedness"`
	DamageMin                                *int                           `json:"damageMin" gorm:"column:damage_min"`
	DamageMax                                *int                           `json:"damageMax" gorm:"column:damage_max"`
	DamageAffinity                           *string                        `json:"damageAffinity" gorm:"column:damage_affinity"`
	SwipesPerAttack                          *int                           `json:"swipesPerAttack" gorm:"column:swipes_per_attack"`
	BlockPercentage                          *int                           `json:"blockPercentage" gorm:"column:block_percentage"`
	DamageBlocked                            *int                           `json:"damageBlocked" gorm:"column:damage_blocked"`
	SpellDamageBonusPercent                  *int                           `json:"spellDamageBonusPercent" gorm:"column:spell_damage_bonus_percent"`
	ConsumeHealthDelta                       int                            `json:"consumeHealthDelta" gorm:"column:consume_health_delta"`
	ConsumeManaDelta                         int                            `json:"consumeManaDelta" gorm:"column:consume_mana_delta"`
	ConsumeRevivePartyMemberHealth           int                            `json:"consumeRevivePartyMemberHealth" gorm:"column:consume_revive_party_member_health"`
	ConsumeReviveAllDownedPartyMembersHealth int                            `json:"consumeReviveAllDownedPartyMembersHealth" gorm:"column:consume_revive_all_downed_party_members_health"`
	ConsumeDealDamage                        int                            `json:"consumeDealDamage" gorm:"column:consume_deal_damage"`
	ConsumeDealDamageHits                    int                            `json:"consumeDealDamageHits" gorm:"column:consume_deal_damage_hits"`
	ConsumeDealDamageAllEnemies              int                            `json:"consumeDealDamageAllEnemies" gorm:"column:consume_deal_damage_all_enemies"`
	ConsumeDealDamageAllEnemiesHits          int                            `json:"consumeDealDamageAllEnemiesHits" gorm:"column:consume_deal_damage_all_enemies_hits"`
	ConsumeStatusesToAdd                     ScenarioFailureStatusTemplates `json:"consumeStatusesToAdd" gorm:"column:consume_statuses_to_add;type:jsonb"`
	ConsumeStatusesToRemove                  StringArray                    `json:"consumeStatusesToRemove" gorm:"column:consume_statuses_to_remove;type:jsonb"`
	ConsumeSpellIDs                          StringArray                    `json:"consumeSpellIds" gorm:"column:consume_spell_ids;type:jsonb"`
	InternalTags                             StringArray                    `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	ImageGenerationStatus                    string                         `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError                     *string                        `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
}

func (InventoryItem) TableName() string {
	return "inventory_items"
}

const (
	InventoryImageGenerationStatusNone       = "none"
	InventoryImageGenerationStatusQueued     = "queued"
	InventoryImageGenerationStatusInProgress = "in_progress"
	InventoryImageGenerationStatusComplete   = "complete"
	InventoryImageGenerationStatusFailed     = "failed"
)
