package processors

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

var supportedBulkEffectTypes = []models.SpellEffectType{
	models.SpellEffectTypeDealDamage,
	models.SpellEffectTypeRestoreLifePartyMember,
	models.SpellEffectTypeRestoreLifeAllParty,
	models.SpellEffectTypeApplyBeneficialStatus,
	models.SpellEffectTypeRemoveDetrimental,
}

type bulkEffectDistributionEntry struct {
	effectType models.SpellEffectType
	count      int
}

func containsAnyKeyword(haystack string, keywords []string) bool {
	normalizedHaystack := strings.ToLower(haystack)
	tokenSet := map[string]struct{}{}
	for _, token := range strings.FieldsFunc(normalizedHaystack, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	}) {
		if token == "" {
			continue
		}
		tokenSet[token] = struct{}{}
	}

	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
		if normalizedKeyword == "" {
			continue
		}
		if strings.Contains(normalizedKeyword, " ") {
			if strings.Contains(normalizedHaystack, normalizedKeyword) {
				return true
			}
			continue
		}
		if _, exists := tokenSet[normalizedKeyword]; exists {
			return true
		}
	}
	return false
}

func inferGeneratedAbilityEffects(
	spec jobs.SpellCreationSpec,
	abilityType models.SpellAbilityType,
	manaCost int,
) models.SpellEffects {
	return inferGeneratedAbilityEffectsWithPreference(spec, abilityType, manaCost, "", nil)
}

func inferGeneratedAbilityEffectsWithPreference(
	spec jobs.SpellCreationSpec,
	abilityType models.SpellAbilityType,
	manaCost int,
	preferred models.SpellEffectType,
	targetLevel *int,
) models.SpellEffects {
	effectType := preferred
	if effectType == "" {
		effectType = inferGeneratedAbilityPrimaryEffectType(spec, abilityType)
	}
	text := strings.ToLower(strings.TrimSpace(strings.Join([]string{
		spec.Name,
		spec.Description,
		spec.EffectText,
		spec.SchoolOfMagic,
	}, " ")))

	switch effectType {
	case models.SpellEffectTypeRestoreLifePartyMember:
		amount := 12 + (manaCost / 2)
		if abilityType == models.SpellAbilityTypeTechnique {
			amount = 9
		}
		if amount < 6 {
			amount = 6
		}
		amount = scaleAbilityAmountForLevel(amount, targetLevel)
		return models.SpellEffects{{
			Type:   models.SpellEffectTypeRestoreLifePartyMember,
			Amount: amount,
		}}
	case models.SpellEffectTypeRestoreLifeAllParty:
		amount := 8 + (manaCost / 3)
		if abilityType == models.SpellAbilityTypeTechnique {
			amount = 6
		}
		if amount < 4 {
			amount = 4
		}
		amount = scaleAbilityAmountForLevel(amount, targetLevel)
		return models.SpellEffects{{
			Type:   models.SpellEffectTypeRestoreLifeAllParty,
			Amount: amount,
		}}
	case models.SpellEffectTypeApplyBeneficialStatus:
		status := models.ScenarioFailureStatusTemplate{
			Name:            "Fortified",
			Description:     "A reinforced stance improves resilience.",
			Effect:          "Gain bonus constitution while active.",
			Positive:        true,
			DurationSeconds: 45,
			ConstitutionMod: 2,
		}
		if containsAnyKeyword(text, []string{"focus", "clarity", "mind", "arcane", "rune"}) {
			status.Name = "Focused"
			status.Description = "Heightened concentration improves execution."
			status.Effect = "Gain bonus intelligence and wisdom while active."
			status.ConstitutionMod = 0
			status.IntelligenceMod = 2
			status.WisdomMod = 1
		}
		if containsAnyKeyword(text, []string{"quick", "swift", "haste", "rush", "step"}) {
			status.Name = "Quickened"
			status.Description = "Fast movement improves reaction speed."
			status.Effect = "Gain bonus dexterity while active."
			status.ConstitutionMod = 0
			status.IntelligenceMod = 0
			status.WisdomMod = 0
			status.DexterityMod = 2
		}
		status.DurationSeconds = scaleAbilityDurationForLevel(status.DurationSeconds, targetLevel)
		status.StrengthMod = scaleAbilityStatModForLevel(status.StrengthMod, targetLevel)
		status.DexterityMod = scaleAbilityStatModForLevel(status.DexterityMod, targetLevel)
		status.ConstitutionMod = scaleAbilityStatModForLevel(status.ConstitutionMod, targetLevel)
		status.IntelligenceMod = scaleAbilityStatModForLevel(status.IntelligenceMod, targetLevel)
		status.WisdomMod = scaleAbilityStatModForLevel(status.WisdomMod, targetLevel)
		status.CharismaMod = scaleAbilityStatModForLevel(status.CharismaMod, targetLevel)
		return models.SpellEffects{{
			Type: models.SpellEffectTypeApplyBeneficialStatus,
			StatusesToApply: models.ScenarioFailureStatusTemplates{
				status,
			},
		}}
	case models.SpellEffectTypeRemoveDetrimental:
		return models.SpellEffects{{
			Type: models.SpellEffectTypeRemoveDetrimental,
			StatusesToRemove: models.StringArray{
				"poisoned",
				"burning",
				"bleeding",
			},
		}}
	case models.SpellEffectTypeDealDamage:
		fallthrough
	default:
		damage := 14 + (manaCost / 3)
		if abilityType == models.SpellAbilityTypeTechnique {
			damage = 10
			if containsAnyKeyword(text, []string{"heavy", "crush", "breaker", "slam", "assault"}) {
				damage = 14
			}
		}
		if damage < 10 {
			damage = 10
		}
		damage = scaleAbilityAmountForLevel(damage, targetLevel)
		return models.SpellEffects{{
			Type:   models.SpellEffectTypeDealDamage,
			Amount: damage,
		}}
	}
}

func normalizeAbilityTargetLevel(raw *int) (int, bool) {
	if raw == nil {
		return 0, false
	}
	level := *raw
	if level < 1 {
		level = 1
	}
	if level > 100 {
		level = 100
	}
	return level, true
}

func abilityLevelScaleMultiplier(level int) float64 {
	// Scale from ~0.75x at level 1 to ~1.75x at level 100.
	return 0.75 + (float64(level-1) * (1.0 / 99.0))
}

func scaleAbilityAmountForLevel(base int, targetLevel *int) int {
	if base <= 0 {
		return base
	}
	level, ok := normalizeAbilityTargetLevel(targetLevel)
	if !ok {
		return base
	}
	scaled := int(math.Round(float64(base) * abilityLevelScaleMultiplier(level)))
	if scaled < 1 {
		return 1
	}
	return scaled
}

func scaleAbilityDurationForLevel(baseSeconds int, targetLevel *int) int {
	if baseSeconds <= 0 {
		return baseSeconds
	}
	level, ok := normalizeAbilityTargetLevel(targetLevel)
	if !ok {
		return baseSeconds
	}
	durationScale := 0.8 + (float64(level-1) * (0.8 / 99.0))
	scaled := int(math.Round(float64(baseSeconds) * durationScale))
	if scaled < 10 {
		return 10
	}
	if scaled > 180 {
		return 180
	}
	return scaled
}

func scaleAbilityStatModForLevel(base int, targetLevel *int) int {
	if base == 0 {
		return 0
	}
	level, ok := normalizeAbilityTargetLevel(targetLevel)
	if !ok {
		return base
	}
	bonus := (level - 1) / 25
	if base > 0 {
		return base + bonus
	}
	return base - bonus
}

func inferGeneratedAbilityPrimaryEffectType(
	spec jobs.SpellCreationSpec,
	abilityType models.SpellAbilityType,
) models.SpellEffectType {
	text := strings.ToLower(strings.TrimSpace(strings.Join([]string{
		spec.Name,
		spec.Description,
		spec.EffectText,
		spec.SchoolOfMagic,
	}, " ")))

	if abilityType == models.SpellAbilityTypeTechnique {
		if containsAnyKeyword(text, []string{"cleanse", "purge", "dispel"}) {
			return models.SpellEffectTypeRemoveDetrimental
		}
		if containsAnyKeyword(text, []string{"ward", "barrier", "guard", "shield", "stance", "fortify"}) {
			return models.SpellEffectTypeApplyBeneficialStatus
		}
		return models.SpellEffectTypeDealDamage
	}

	if containsAnyKeyword(text, []string{"heal", "renew", "restore", "revive", "recovery", "vital"}) {
		if containsAnyKeyword(text, []string{"all", "party", "group", "aura"}) {
			return models.SpellEffectTypeRestoreLifeAllParty
		}
		return models.SpellEffectTypeRestoreLifePartyMember
	}
	if containsAnyKeyword(text, []string{"ward", "barrier", "guard", "shield", "stance", "fortify"}) {
		return models.SpellEffectTypeApplyBeneficialStatus
	}
	if containsAnyKeyword(text, []string{"cleanse", "purge", "dispel"}) {
		return models.SpellEffectTypeRemoveDetrimental
	}
	return models.SpellEffectTypeDealDamage
}

func buildGeneratedAbilityEffectText(effects models.SpellEffects, abilityType models.SpellAbilityType) string {
	if len(effects) == 0 {
		if abilityType == models.SpellAbilityTypeTechnique {
			return "Executes a disciplined combat maneuver."
		}
		return "Channels focused magical energy."
	}
	parts := make([]string, 0, len(effects))
	for _, effect := range effects {
		switch effect.Type {
		case models.SpellEffectTypeDealDamage:
			if effect.Amount > 0 {
				parts = append(parts, fmt.Sprintf("Deals %d damage.", effect.Amount))
			} else {
				parts = append(parts, "Deals direct damage.")
			}
		case models.SpellEffectTypeRestoreLifePartyMember:
			if effect.Amount > 0 {
				parts = append(parts, fmt.Sprintf("Restores %d health to one ally.", effect.Amount))
			} else {
				parts = append(parts, "Restores health to one ally.")
			}
		case models.SpellEffectTypeRestoreLifeAllParty:
			if effect.Amount > 0 {
				parts = append(parts, fmt.Sprintf("Restores %d health to all allies.", effect.Amount))
			} else {
				parts = append(parts, "Restores health to all allies.")
			}
		case models.SpellEffectTypeApplyBeneficialStatus:
			if len(effect.StatusesToApply) == 0 {
				parts = append(parts, "Applies a beneficial status.")
				continue
			}
			statusParts := make([]string, 0, len(effect.StatusesToApply))
			for _, status := range effect.StatusesToApply {
				name := strings.TrimSpace(status.Name)
				if name == "" {
					name = "beneficial status"
				}
				if status.DurationSeconds > 0 {
					statusParts = append(statusParts, fmt.Sprintf("%s (%ds)", name, status.DurationSeconds))
				} else {
					statusParts = append(statusParts, name)
				}
			}
			parts = append(parts, fmt.Sprintf("Applies %s.", strings.Join(statusParts, ", ")))
		case models.SpellEffectTypeRemoveDetrimental:
			if len(effect.StatusesToRemove) == 0 {
				parts = append(parts, "Removes detrimental statuses.")
				continue
			}
			parts = append(parts, fmt.Sprintf("Removes %s.", strings.Join(effect.StatusesToRemove, ", ")))
		default:
			parts = append(parts, "Triggers a special effect.")
		}
	}
	return strings.Join(parts, " ")
}

func buildConfiguredAbilityEffectPlan(
	total int,
	counts *jobs.SpellBulkEffectCounts,
) []models.SpellEffectType {
	if total <= 0 || counts == nil {
		return nil
	}

	configuredCounts := map[models.SpellEffectType]int{
		models.SpellEffectTypeDealDamage:             counts.DealDamage,
		models.SpellEffectTypeRestoreLifePartyMember: counts.RestoreLifePartyMember,
		models.SpellEffectTypeRestoreLifeAllParty:    counts.RestoreLifeAllParty,
		models.SpellEffectTypeApplyBeneficialStatus:  counts.ApplyBeneficialStatuses,
		models.SpellEffectTypeRemoveDetrimental:      counts.RemoveDetrimentalEffects,
	}

	entries := make([]bulkEffectDistributionEntry, 0, len(supportedBulkEffectTypes))
	configuredTotal := 0
	for _, effectType := range supportedBulkEffectTypes {
		count := configuredCounts[effectType]
		if count <= 0 {
			continue
		}
		entries = append(entries, bulkEffectDistributionEntry{effectType: effectType, count: count})
		configuredTotal += count
	}
	if len(entries) == 0 || configuredTotal != total {
		return nil
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].count > entries[j].count
	})

	plan := make([]models.SpellEffectType, 0, total)
	for len(plan) < total {
		progressed := false
		for index := range entries {
			if entries[index].count <= 0 {
				continue
			}
			plan = append(plan, entries[index].effectType)
			entries[index].count--
			progressed = true
			if len(plan) >= total {
				break
			}
		}
		if !progressed {
			break
		}
	}
	return plan
}
