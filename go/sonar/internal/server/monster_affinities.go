package server

import (
	"fmt"
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
) (adjustedDamage int, normalizedAffinity *string, modifier monsterAffinityModifier) {
	if damage <= 0 {
		return 0, models.NormalizeOptionalDamageAffinity(rawAffinity), monsterAffinityModifierNone
	}

	normalizedAffinity = models.NormalizeOptionalDamageAffinity(rawAffinity)
	if monster == nil || monster.Template == nil || normalizedAffinity == nil {
		return damage, normalizedAffinity, monsterAffinityModifierNone
	}

	strongAgainst := models.NormalizeOptionalDamageAffinity(monster.Template.StrongAgainstAffinity)
	if strongAgainst != nil && *strongAgainst == *normalizedAffinity {
		return max(1, damage/2), normalizedAffinity, monsterAffinityModifierStrongAgainst
	}

	weakAgainst := models.NormalizeOptionalDamageAffinity(monster.Template.WeakAgainstAffinity)
	if weakAgainst != nil && *weakAgainst == *normalizedAffinity {
		return damage * 2, normalizedAffinity, monsterAffinityModifierWeakAgainst
	}

	return damage, normalizedAffinity, monsterAffinityModifierNone
}
