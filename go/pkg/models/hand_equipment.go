package models

import (
	"fmt"
	"strings"
)

type HandItemCategory string

const (
	HandItemCategoryWeapon HandItemCategory = "weapon"
	HandItemCategoryShield HandItemCategory = "shield"
	HandItemCategoryOrb    HandItemCategory = "orb"
	HandItemCategoryStaff  HandItemCategory = "staff"
)

var HandItemCategories = []HandItemCategory{
	HandItemCategoryWeapon,
	HandItemCategoryShield,
	HandItemCategoryOrb,
	HandItemCategoryStaff,
}

type Handedness string

const (
	HandednessOneHanded Handedness = "one_handed"
	HandednessTwoHanded Handedness = "two_handed"
)

var HandednessOptions = []Handedness{
	HandednessOneHanded,
	HandednessTwoHanded,
}

type HandEquipmentAttributes struct {
	HandItemCategory        *string
	Handedness              *string
	DamageMin               *int
	DamageMax               *int
	DamageAffinity          *string
	SwipesPerAttack         *int
	BlockPercentage         *int
	DamageBlocked           *int
	SpellDamageBonusPercent *int
}

func IsValidHandItemCategory(value string) bool {
	for _, category := range HandItemCategories {
		if string(category) == value {
			return true
		}
	}
	return false
}

func IsValidHandedness(value string) bool {
	for _, handedness := range HandednessOptions {
		if string(handedness) == value {
			return true
		}
	}
	return false
}

func IsHandEquipSlot(slot string) bool {
	switch slot {
	case string(EquipmentSlotDominantHand), string(EquipmentSlotOffHand):
		return true
	default:
		return false
	}
}

func NormalizeAndValidateHandEquipment(
	equipSlot *string,
	input HandEquipmentAttributes,
) (HandEquipmentAttributes, error) {
	normalized := HandEquipmentAttributes{
		HandItemCategory:        normalizeOptionalLowerString(input.HandItemCategory),
		Handedness:              normalizeOptionalLowerString(input.Handedness),
		DamageMin:               input.DamageMin,
		DamageMax:               input.DamageMax,
		DamageAffinity:          normalizeOptionalLowerString(input.DamageAffinity),
		SwipesPerAttack:         input.SwipesPerAttack,
		BlockPercentage:         input.BlockPercentage,
		DamageBlocked:           input.DamageBlocked,
		SpellDamageBonusPercent: input.SpellDamageBonusPercent,
	}

	slot := ""
	if equipSlot != nil {
		slot = strings.TrimSpace(*equipSlot)
	}
	if !IsHandEquipSlot(slot) {
		if hasAnyHandEquipmentField(normalized) {
			return HandEquipmentAttributes{}, fmt.Errorf("hand equipment fields are only valid for dominant_hand or off_hand")
		}
		return HandEquipmentAttributes{}, nil
	}

	if normalized.HandItemCategory == nil || !IsValidHandItemCategory(*normalized.HandItemCategory) {
		return HandEquipmentAttributes{}, fmt.Errorf("valid handItemCategory is required for hand equipment")
	}
	if normalized.Handedness == nil || !IsValidHandedness(*normalized.Handedness) {
		return HandEquipmentAttributes{}, fmt.Errorf("valid handedness is required for hand equipment")
	}

	switch slot {
	case string(EquipmentSlotDominantHand):
		return validateDominantHandAttributes(normalized)
	case string(EquipmentSlotOffHand):
		return validateOffHandAttributes(normalized)
	default:
		return HandEquipmentAttributes{}, fmt.Errorf("invalid hand equip slot")
	}
}

func validateDominantHandAttributes(attrs HandEquipmentAttributes) (HandEquipmentAttributes, error) {
	category := *attrs.HandItemCategory
	switch category {
	case string(HandItemCategoryWeapon), string(HandItemCategoryStaff):
	default:
		return HandEquipmentAttributes{}, fmt.Errorf("dominant hand items must be weapon or staff")
	}

	if attrs.DamageMin == nil || attrs.DamageMax == nil || attrs.SwipesPerAttack == nil {
		return HandEquipmentAttributes{}, fmt.Errorf("dominant hand items require damageMin, damageMax, and swipesPerAttack")
	}
	if attrs.DamageAffinity == nil {
		defaultAffinity := string(DamageAffinityPhysical)
		if category == string(HandItemCategoryStaff) {
			defaultAffinity = string(DamageAffinityArcane)
		}
		attrs.DamageAffinity = &defaultAffinity
	}
	if !IsValidDamageAffinity(*attrs.DamageAffinity) {
		return HandEquipmentAttributes{}, fmt.Errorf("damageAffinity must be one of: physical, fire, ice, lightning, poison, arcane, holy, shadow")
	}
	if *attrs.DamageMin <= 0 || *attrs.DamageMax <= 0 || *attrs.SwipesPerAttack <= 0 {
		return HandEquipmentAttributes{}, fmt.Errorf("damageMin, damageMax, and swipesPerAttack must be positive")
	}
	if *attrs.DamageMax < *attrs.DamageMin {
		return HandEquipmentAttributes{}, fmt.Errorf("damageMax must be greater than or equal to damageMin")
	}
	if attrs.BlockPercentage != nil || attrs.DamageBlocked != nil {
		return HandEquipmentAttributes{}, fmt.Errorf("dominant hand items cannot set shield block fields")
	}

	handedness := *attrs.Handedness
	if category == string(HandItemCategoryStaff) {
		if handedness != string(HandednessTwoHanded) {
			return HandEquipmentAttributes{}, fmt.Errorf("staff items must be two_handed")
		}
		if attrs.SpellDamageBonusPercent == nil || *attrs.SpellDamageBonusPercent <= 0 {
			return HandEquipmentAttributes{}, fmt.Errorf("staff items require spellDamageBonusPercent > 0")
		}
		return attrs, nil
	}

	if attrs.SpellDamageBonusPercent != nil {
		if *attrs.SpellDamageBonusPercent <= 0 {
			return HandEquipmentAttributes{}, fmt.Errorf("spellDamageBonusPercent must be positive when set")
		}
		return HandEquipmentAttributes{}, fmt.Errorf("weapon items cannot set spellDamageBonusPercent")
	}

	return attrs, nil
}

func validateOffHandAttributes(attrs HandEquipmentAttributes) (HandEquipmentAttributes, error) {
	if *attrs.Handedness != string(HandednessOneHanded) {
		return HandEquipmentAttributes{}, fmt.Errorf("off hand items must be one_handed")
	}
	if attrs.DamageMin != nil || attrs.DamageMax != nil || attrs.SwipesPerAttack != nil {
		return HandEquipmentAttributes{}, fmt.Errorf("off hand items cannot set damage fields")
	}
	if attrs.DamageAffinity != nil {
		return HandEquipmentAttributes{}, fmt.Errorf("off hand items cannot set damageAffinity")
	}

	category := *attrs.HandItemCategory
	switch category {
	case string(HandItemCategoryShield):
		if attrs.BlockPercentage == nil || attrs.DamageBlocked == nil {
			return HandEquipmentAttributes{}, fmt.Errorf("shield items require blockPercentage and damageBlocked")
		}
		if *attrs.BlockPercentage <= 0 || *attrs.BlockPercentage > 100 {
			return HandEquipmentAttributes{}, fmt.Errorf("blockPercentage must be between 1 and 100")
		}
		if *attrs.DamageBlocked <= 0 {
			return HandEquipmentAttributes{}, fmt.Errorf("damageBlocked must be positive")
		}
		if attrs.SpellDamageBonusPercent != nil {
			return HandEquipmentAttributes{}, fmt.Errorf("shield items cannot set spellDamageBonusPercent")
		}
		return attrs, nil
	case string(HandItemCategoryOrb):
		if attrs.BlockPercentage != nil || attrs.DamageBlocked != nil {
			return HandEquipmentAttributes{}, fmt.Errorf("orb items cannot set shield block fields")
		}
		if attrs.SpellDamageBonusPercent == nil || *attrs.SpellDamageBonusPercent <= 0 {
			return HandEquipmentAttributes{}, fmt.Errorf("orb items require spellDamageBonusPercent > 0")
		}
		return attrs, nil
	default:
		return HandEquipmentAttributes{}, fmt.Errorf("off hand items must be shield or orb")
	}
}

func hasAnyHandEquipmentField(attrs HandEquipmentAttributes) bool {
	return attrs.HandItemCategory != nil ||
		attrs.Handedness != nil ||
		attrs.DamageMin != nil ||
		attrs.DamageMax != nil ||
		attrs.DamageAffinity != nil ||
		attrs.SwipesPerAttack != nil ||
		attrs.BlockPercentage != nil ||
		attrs.DamageBlocked != nil ||
		attrs.SpellDamageBonusPercent != nil
}

func normalizeOptionalLowerString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.ToLower(strings.TrimSpace(*value))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
