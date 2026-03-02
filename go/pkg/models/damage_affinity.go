package models

import "strings"

type DamageAffinity string

const (
	DamageAffinityPhysical  DamageAffinity = "physical"
	DamageAffinityFire      DamageAffinity = "fire"
	DamageAffinityIce       DamageAffinity = "ice"
	DamageAffinityLightning DamageAffinity = "lightning"
	DamageAffinityPoison    DamageAffinity = "poison"
	DamageAffinityArcane    DamageAffinity = "arcane"
	DamageAffinityHoly      DamageAffinity = "holy"
	DamageAffinityShadow    DamageAffinity = "shadow"
)

var DamageAffinities = []DamageAffinity{
	DamageAffinityPhysical,
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

func IsValidDamageAffinity(raw string) bool {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	for _, affinity := range DamageAffinities {
		if normalized == string(affinity) {
			return true
		}
	}
	return false
}
