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
	Strength                      int `json:"strength"`
	Dexterity                     int `json:"dexterity"`
	Constitution                  int `json:"constitution"`
	Intelligence                  int `json:"intelligence"`
	Wisdom                        int `json:"wisdom"`
	Charisma                      int `json:"charisma"`
	PhysicalDamageBonusPercent    int `json:"physicalDamageBonusPercent"`
	PiercingDamageBonusPercent    int `json:"piercingDamageBonusPercent"`
	SlashingDamageBonusPercent    int `json:"slashingDamageBonusPercent"`
	BludgeoningDamageBonusPercent int `json:"bludgeoningDamageBonusPercent"`
	FireDamageBonusPercent        int `json:"fireDamageBonusPercent"`
	IceDamageBonusPercent         int `json:"iceDamageBonusPercent"`
	LightningDamageBonusPercent   int `json:"lightningDamageBonusPercent"`
	PoisonDamageBonusPercent      int `json:"poisonDamageBonusPercent"`
	ArcaneDamageBonusPercent      int `json:"arcaneDamageBonusPercent"`
	HolyDamageBonusPercent        int `json:"holyDamageBonusPercent"`
	ShadowDamageBonusPercent      int `json:"shadowDamageBonusPercent"`
	PhysicalResistancePercent     int `json:"physicalResistancePercent"`
	PiercingResistancePercent     int `json:"piercingResistancePercent"`
	SlashingResistancePercent     int `json:"slashingResistancePercent"`
	BludgeoningResistancePercent  int `json:"bludgeoningResistancePercent"`
	FireResistancePercent         int `json:"fireResistancePercent"`
	IceResistancePercent          int `json:"iceResistancePercent"`
	LightningResistancePercent    int `json:"lightningResistancePercent"`
	PoisonResistancePercent       int `json:"poisonResistancePercent"`
	ArcaneResistancePercent       int `json:"arcaneResistancePercent"`
	HolyResistancePercent         int `json:"holyResistancePercent"`
	ShadowResistancePercent       int `json:"shadowResistancePercent"`
}

func (b CharacterStatBonuses) Add(other CharacterStatBonuses) CharacterStatBonuses {
	return CharacterStatBonuses{
		Strength:                      b.Strength + other.Strength,
		Dexterity:                     b.Dexterity + other.Dexterity,
		Constitution:                  b.Constitution + other.Constitution,
		Intelligence:                  b.Intelligence + other.Intelligence,
		Wisdom:                        b.Wisdom + other.Wisdom,
		Charisma:                      b.Charisma + other.Charisma,
		PhysicalDamageBonusPercent:    b.PhysicalDamageBonusPercent + other.PhysicalDamageBonusPercent,
		PiercingDamageBonusPercent:    b.PiercingDamageBonusPercent + other.PiercingDamageBonusPercent,
		SlashingDamageBonusPercent:    b.SlashingDamageBonusPercent + other.SlashingDamageBonusPercent,
		BludgeoningDamageBonusPercent: b.BludgeoningDamageBonusPercent + other.BludgeoningDamageBonusPercent,
		FireDamageBonusPercent:        b.FireDamageBonusPercent + other.FireDamageBonusPercent,
		IceDamageBonusPercent:         b.IceDamageBonusPercent + other.IceDamageBonusPercent,
		LightningDamageBonusPercent:   b.LightningDamageBonusPercent + other.LightningDamageBonusPercent,
		PoisonDamageBonusPercent:      b.PoisonDamageBonusPercent + other.PoisonDamageBonusPercent,
		ArcaneDamageBonusPercent:      b.ArcaneDamageBonusPercent + other.ArcaneDamageBonusPercent,
		HolyDamageBonusPercent:        b.HolyDamageBonusPercent + other.HolyDamageBonusPercent,
		ShadowDamageBonusPercent:      b.ShadowDamageBonusPercent + other.ShadowDamageBonusPercent,
		PhysicalResistancePercent:     b.PhysicalResistancePercent + other.PhysicalResistancePercent,
		PiercingResistancePercent:     b.PiercingResistancePercent + other.PiercingResistancePercent,
		SlashingResistancePercent:     b.SlashingResistancePercent + other.SlashingResistancePercent,
		BludgeoningResistancePercent:  b.BludgeoningResistancePercent + other.BludgeoningResistancePercent,
		FireResistancePercent:         b.FireResistancePercent + other.FireResistancePercent,
		IceResistancePercent:          b.IceResistancePercent + other.IceResistancePercent,
		LightningResistancePercent:    b.LightningResistancePercent + other.LightningResistancePercent,
		PoisonResistancePercent:       b.PoisonResistancePercent + other.PoisonResistancePercent,
		ArcaneResistancePercent:       b.ArcaneResistancePercent + other.ArcaneResistancePercent,
		HolyResistancePercent:         b.HolyResistancePercent + other.HolyResistancePercent,
		ShadowResistancePercent:       b.ShadowResistancePercent + other.ShadowResistancePercent,
	}
}

func (b CharacterStatBonuses) ToMap() map[string]int {
	return map[string]int{
		"strength":                      b.Strength,
		"dexterity":                     b.Dexterity,
		"constitution":                  b.Constitution,
		"intelligence":                  b.Intelligence,
		"wisdom":                        b.Wisdom,
		"charisma":                      b.Charisma,
		"physicalDamageBonusPercent":    b.PhysicalDamageBonusPercent,
		"piercingDamageBonusPercent":    b.PiercingDamageBonusPercent,
		"slashingDamageBonusPercent":    b.SlashingDamageBonusPercent,
		"bludgeoningDamageBonusPercent": b.BludgeoningDamageBonusPercent,
		"fireDamageBonusPercent":        b.FireDamageBonusPercent,
		"iceDamageBonusPercent":         b.IceDamageBonusPercent,
		"lightningDamageBonusPercent":   b.LightningDamageBonusPercent,
		"poisonDamageBonusPercent":      b.PoisonDamageBonusPercent,
		"arcaneDamageBonusPercent":      b.ArcaneDamageBonusPercent,
		"holyDamageBonusPercent":        b.HolyDamageBonusPercent,
		"shadowDamageBonusPercent":      b.ShadowDamageBonusPercent,
		"physicalResistancePercent":     b.PhysicalResistancePercent,
		"piercingResistancePercent":     b.PiercingResistancePercent,
		"slashingResistancePercent":     b.SlashingResistancePercent,
		"bludgeoningResistancePercent":  b.BludgeoningResistancePercent,
		"fireResistancePercent":         b.FireResistancePercent,
		"iceResistancePercent":          b.IceResistancePercent,
		"lightningResistancePercent":    b.LightningResistancePercent,
		"poisonResistancePercent":       b.PoisonResistancePercent,
		"arcaneResistancePercent":       b.ArcaneResistancePercent,
		"holyResistancePercent":         b.HolyResistancePercent,
		"shadowResistancePercent":       b.ShadowResistancePercent,
	}
}

func (b CharacterStatBonuses) AffinityDamageBonusMap() map[string]int {
	return map[string]int{
		"physical":    b.PhysicalDamageBonusPercent,
		"piercing":    b.PiercingDamageBonusPercent,
		"slashing":    b.SlashingDamageBonusPercent,
		"bludgeoning": b.BludgeoningDamageBonusPercent,
		"fire":        b.FireDamageBonusPercent,
		"ice":         b.IceDamageBonusPercent,
		"lightning":   b.LightningDamageBonusPercent,
		"poison":      b.PoisonDamageBonusPercent,
		"arcane":      b.ArcaneDamageBonusPercent,
		"holy":        b.HolyDamageBonusPercent,
		"shadow":      b.ShadowDamageBonusPercent,
	}
}

func (b CharacterStatBonuses) DamageBonusPercentForAffinity(raw string) int {
	switch NormalizeDamageAffinity(raw) {
	case DamageAffinityPiercing:
		return b.PiercingDamageBonusPercent
	case DamageAffinitySlashing:
		return b.SlashingDamageBonusPercent
	case DamageAffinityBludgeoning:
		return b.BludgeoningDamageBonusPercent
	case DamageAffinityFire:
		return b.FireDamageBonusPercent
	case DamageAffinityIce:
		return b.IceDamageBonusPercent
	case DamageAffinityLightning:
		return b.LightningDamageBonusPercent
	case DamageAffinityPoison:
		return b.PoisonDamageBonusPercent
	case DamageAffinityArcane:
		return b.ArcaneDamageBonusPercent
	case DamageAffinityHoly:
		return b.HolyDamageBonusPercent
	case DamageAffinityShadow:
		return b.ShadowDamageBonusPercent
	default:
		return b.PhysicalDamageBonusPercent
	}
}

func (b CharacterStatBonuses) AffinityResistanceMap() map[string]int {
	return map[string]int{
		"physical":    b.PhysicalResistancePercent,
		"piercing":    b.PiercingResistancePercent,
		"slashing":    b.SlashingResistancePercent,
		"bludgeoning": b.BludgeoningResistancePercent,
		"fire":        b.FireResistancePercent,
		"ice":         b.IceResistancePercent,
		"lightning":   b.LightningResistancePercent,
		"poison":      b.PoisonResistancePercent,
		"arcane":      b.ArcaneResistancePercent,
		"holy":        b.HolyResistancePercent,
		"shadow":      b.ShadowResistancePercent,
	}
}

func (b CharacterStatBonuses) ResistancePercentForAffinity(raw string) int {
	switch NormalizeDamageAffinity(raw) {
	case DamageAffinityPiercing:
		return b.PiercingResistancePercent
	case DamageAffinitySlashing:
		return b.SlashingResistancePercent
	case DamageAffinityBludgeoning:
		return b.BludgeoningResistancePercent
	case DamageAffinityFire:
		return b.FireResistancePercent
	case DamageAffinityIce:
		return b.IceResistancePercent
	case DamageAffinityLightning:
		return b.LightningResistancePercent
	case DamageAffinityPoison:
		return b.PoisonResistancePercent
	case DamageAffinityArcane:
		return b.ArcaneResistancePercent
	case DamageAffinityHoly:
		return b.HolyResistancePercent
	case DamageAffinityShadow:
		return b.ShadowResistancePercent
	default:
		return b.PhysicalResistancePercent
	}
}
