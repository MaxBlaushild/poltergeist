package server

import (
	"fmt"
	"math"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

type monsterAffinityModifier string

const (
	monsterAffinityModifierNone          monsterAffinityModifier = ""
	monsterAffinityModifierStrongAgainst monsterAffinityModifier = "strong_against"
	monsterAffinityModifierWeakAgainst   monsterAffinityModifier = "weak_against"
)

func parseOptionalDamageAffinity(raw string, fieldName string) (*string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	if !models.IsValidDamageAffinity(trimmed) {
		options := make([]string, 0, len(models.DamageAffinities))
		for _, affinity := range models.DamageAffinities {
			options = append(options, string(affinity))
		}
		return nil, fmt.Errorf("%s must be one of %s", fieldName, strings.Join(options, ", "))
	}
	normalized := string(models.NormalizeDamageAffinity(trimmed))
	return &normalized, nil
}

func applyMonsterAffinityDamage(
	monster *models.Monster,
	damage int,
	rawAffinity *string,
	statusBonuses models.CharacterStatBonuses,
) (adjustedDamage int, normalizedAffinity *string, modifier monsterAffinityModifier) {
	if damage <= 0 {
		return 0, models.NormalizeOptionalDamageAffinity(rawAffinity), monsterAffinityModifierNone
	}

	normalizedAffinity = models.NormalizeOptionalDamageAffinity(rawAffinity)
	if normalizedAffinity == nil {
		return damage, normalizedAffinity, monsterAffinityModifierNone
	}
	totalBonuses := statusBonuses
	if monster != nil && monster.Template != nil {
		totalBonuses = monster.Template.AffinityBonuses().Add(totalBonuses)
	}
	resistancePercent := totalBonuses.ResistancePercentForAffinity(*normalizedAffinity)
	if resistancePercent == 0 {
		return damage, normalizedAffinity, monsterAffinityModifierNone
	}
	reduction := int(math.Round(float64(damage) * float64(resistancePercent) / 100.0))
	adjustedDamage = damage - reduction
	if adjustedDamage < 0 {
		adjustedDamage = 0
	}
	if resistancePercent > 0 {
		return adjustedDamage, normalizedAffinity, monsterAffinityModifierStrongAgainst
	}
	return adjustedDamage, normalizedAffinity, monsterAffinityModifierWeakAgainst
}

func applyCharacterAffinityResistance(
	damage int,
	rawAffinity *string,
	bonuses models.CharacterStatBonuses,
) (adjustedDamage int, normalizedAffinity *string, resistancePercent int) {
	if damage <= 0 {
		return 0, models.NormalizeOptionalDamageAffinity(rawAffinity), 0
	}
	normalizedAffinity = models.NormalizeOptionalDamageAffinity(rawAffinity)
	if normalizedAffinity == nil {
		return damage, nil, 0
	}
	resistancePercent = bonuses.ResistancePercentForAffinity(*normalizedAffinity)
	if resistancePercent == 0 {
		return damage, normalizedAffinity, 0
	}
	reduction := int(math.Round(float64(damage) * float64(resistancePercent) / 100.0))
	adjustedDamage = damage - reduction
	if adjustedDamage < 0 {
		adjustedDamage = 0
	}
	return adjustedDamage, normalizedAffinity, resistancePercent
}

func applyAffinityDamageBonus(
	damage int,
	rawAffinity *string,
	bonuses models.CharacterStatBonuses,
) (adjustedDamage int, normalizedAffinity *string, bonusPercent int) {
	if damage <= 0 {
		return 0, models.NormalizeOptionalDamageAffinity(rawAffinity), 0
	}
	normalizedAffinity = models.NormalizeOptionalDamageAffinity(rawAffinity)
	if normalizedAffinity == nil {
		return damage, nil, 0
	}
	bonusPercent = bonuses.DamageBonusPercentForAffinity(*normalizedAffinity)
	if bonusPercent == 0 {
		return damage, normalizedAffinity, 0
	}
	bonus := int(math.Round(float64(damage) * float64(bonusPercent) / 100.0))
	adjustedDamage = damage + bonus
	if adjustedDamage < 0 {
		adjustedDamage = 0
	}
	return adjustedDamage, normalizedAffinity, bonusPercent
}
