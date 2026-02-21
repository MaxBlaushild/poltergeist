package models

type EquipmentSlot string

const (
	EquipmentSlotHat          EquipmentSlot = "hat"
	EquipmentSlotNecklace     EquipmentSlot = "necklace"
	EquipmentSlotChest        EquipmentSlot = "chest"
	EquipmentSlotLegs         EquipmentSlot = "legs"
	EquipmentSlotShoes        EquipmentSlot = "shoes"
	EquipmentSlotGloves       EquipmentSlot = "gloves"
	EquipmentSlotDominantHand EquipmentSlot = "dominant_hand"
	EquipmentSlotOffHand      EquipmentSlot = "off_hand"
	EquipmentSlotRing         EquipmentSlot = "ring"
	EquipmentSlotRingLeft     EquipmentSlot = "ring_left"
	EquipmentSlotRingRight    EquipmentSlot = "ring_right"
)

var EquipmentSlots = []EquipmentSlot{
	EquipmentSlotHat,
	EquipmentSlotNecklace,
	EquipmentSlotChest,
	EquipmentSlotLegs,
	EquipmentSlotShoes,
	EquipmentSlotGloves,
	EquipmentSlotDominantHand,
	EquipmentSlotOffHand,
	EquipmentSlotRingLeft,
	EquipmentSlotRingRight,
}

var InventoryEquipSlots = []EquipmentSlot{
	EquipmentSlotHat,
	EquipmentSlotNecklace,
	EquipmentSlotChest,
	EquipmentSlotLegs,
	EquipmentSlotShoes,
	EquipmentSlotGloves,
	EquipmentSlotDominantHand,
	EquipmentSlotOffHand,
	EquipmentSlotRing,
	EquipmentSlotRingLeft,
	EquipmentSlotRingRight,
}

func IsValidEquipmentSlot(slot string) bool {
	for _, s := range EquipmentSlots {
		if string(s) == slot {
			return true
		}
	}
	return false
}

func IsValidInventoryEquipSlot(slot string) bool {
	for _, s := range InventoryEquipSlots {
		if string(s) == slot {
			return true
		}
	}
	return false
}

func IsRingSlot(slot string) bool {
	return slot == string(EquipmentSlotRingLeft) || slot == string(EquipmentSlotRingRight)
}

type CharacterStatBonuses struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

func (b CharacterStatBonuses) ToMap() map[string]int {
	return map[string]int{
		"strength":     b.Strength,
		"dexterity":    b.Dexterity,
		"constitution": b.Constitution,
		"intelligence": b.Intelligence,
		"wisdom":       b.Wisdom,
		"charisma":     b.Charisma,
	}
}
