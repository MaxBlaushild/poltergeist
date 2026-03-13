package models

import "strings"

type DamageAffinity string

const (
	DamageAffinityPhysical    DamageAffinity = "physical"
	DamageAffinityPiercing    DamageAffinity = "piercing"
	DamageAffinitySlashing    DamageAffinity = "slashing"
	DamageAffinityBludgeoning DamageAffinity = "bludgeoning"
	DamageAffinityFire        DamageAffinity = "fire"
	DamageAffinityIce         DamageAffinity = "ice"
	DamageAffinityLightning   DamageAffinity = "lightning"
	DamageAffinityPoison      DamageAffinity = "poison"
	DamageAffinityArcane      DamageAffinity = "arcane"
	DamageAffinityHoly        DamageAffinity = "holy"
	DamageAffinityShadow      DamageAffinity = "shadow"
)

var DamageAffinities = []DamageAffinity{
	DamageAffinityPhysical,
	DamageAffinityPiercing,
	DamageAffinitySlashing,
	DamageAffinityBludgeoning,
	DamageAffinityFire,
	DamageAffinityIce,
	DamageAffinityLightning,
	DamageAffinityPoison,
	DamageAffinityArcane,
	DamageAffinityHoly,
	DamageAffinityShadow,
}

func NormalizeDamageAffinity(raw string) DamageAffinity {
	switch DamageAffinity(strings.ToLower(strings.TrimSpace(raw))) {
	case DamageAffinityPiercing:
		return DamageAffinityPiercing
	case DamageAffinitySlashing:
		return DamageAffinitySlashing
	case DamageAffinityBludgeoning:
		return DamageAffinityBludgeoning
	case DamageAffinityFire:
		return DamageAffinityFire
	case DamageAffinityIce:
		return DamageAffinityIce
	case DamageAffinityLightning:
		return DamageAffinityLightning
	case DamageAffinityPoison:
		return DamageAffinityPoison
	case DamageAffinityArcane:
		return DamageAffinityArcane
	case DamageAffinityHoly:
		return DamageAffinityHoly
	case DamageAffinityShadow:
		return DamageAffinityShadow
	default:
		return DamageAffinityPhysical
	}
}

func IsPhysicalLikeDamageAffinity(raw string) bool {
	switch NormalizeDamageAffinity(raw) {
	case DamageAffinityPhysical, DamageAffinityPiercing, DamageAffinitySlashing, DamageAffinityBludgeoning:
		return true
	default:
		return false
	}
}

func NormalizeOptionalDamageAffinity(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}
	normalized := string(NormalizeDamageAffinity(trimmed))
	return &normalized
}

func IsValidDamageAffinity(raw string) bool {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	for _, affinity := range DamageAffinities {
		if normalized == string(affinity) {
			return true
		}
	}
	return false
}
