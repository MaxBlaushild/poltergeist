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

func inferGeneratedDamageAffinity(text string, abilityType models.SpellAbilityType) string {
	if containsAnyKeyword(text, []string{"fire", "flame", "ember", "inferno", "burn", "pyro"}) {
		return string(models.DamageAffinityFire)
	}
	if containsAnyKeyword(text, []string{"ice", "frost", "glacier", "chill", "cryo"}) {
		return string(models.DamageAffinityIce)
	}
	if containsAnyKeyword(text, []string{"lightning", "storm", "thunder", "volt", "spark"}) {
		return string(models.DamageAffinityLightning)
	}
	if containsAnyKeyword(text, []string{"venom", "poison", "toxin", "blight"}) {
		return string(models.DamageAffinityPoison)
	}
	if containsAnyKeyword(text, []string{"holy", "radiant", "sun", "dawn", "sanct"}) {
		return string(models.DamageAffinityHoly)
	}
	if containsAnyKeyword(text, []string{"shadow", "void", "night", "umbral", "hex"}) {
		return string(models.DamageAffinityShadow)
	}
	if abilityType == models.SpellAbilityTypeTechnique {
		return string(models.DamageAffinityPhysical)
	}
	if containsAnyKeyword(text, []string{"arcane", "rune", "spell", "mana", "astral"}) {
		return string(models.DamageAffinityArcane)
	}
	return string(models.DamageAffinityPhysical)
}

func affinityKeywordsForName(affinity string) []string {
	switch models.NormalizeDamageAffinity(affinity) {
	case models.DamageAffinityFire:
		return []string{"fire", "flame", "ember", "inferno", "burn"}
	case models.DamageAffinityIce:
		return []string{"ice", "frost", "glacier", "chill", "rime"}
	case models.DamageAffinityLightning:
		return []string{"lightning", "storm", "thunder", "volt", "spark"}
	case models.DamageAffinityPoison:
		return []string{"poison", "venom", "toxin", "blight"}
	case models.DamageAffinityArcane:
		return []string{"arcane", "rune", "mana", "astral"}
	case models.DamageAffinityHoly:
		return []string{"holy", "radiant", "sanct", "dawn", "sun"}
	case models.DamageAffinityShadow:
		return []string{"shadow", "void", "umbral", "night", "hex"}
	default:
		return []string{"physical", "blade", "strike", "slam", "crush", "rend"}
	}
}

func effectKeywordsForName(effectType models.SpellEffectType) []string {
	switch effectType {
	case models.SpellEffectTypeRestoreLifePartyMember:
		return []string{"heal", "mend", "renew", "restore", "recovery", "vital"}
	case models.SpellEffectTypeRestoreLifeAllParty:
		return []string{"heal", "renew", "restore", "chorus", "rally", "aura", "all"}
	case models.SpellEffectTypeApplyBeneficialStatus:
		return []string{"aegis", "ward", "guard", "boon", "fortify", "stance", "focus"}
	case models.SpellEffectTypeRemoveDetrimental:
		return []string{"cleanse", "purge", "dispel", "purify", "clear"}
	case models.SpellEffectTypeDealDamage:
		return []string{"strike", "lance", "bolt", "blast", "slash", "rend", "burst", "volley", "spear", "javelin", "assault", "pounce", "riposte"}
	default:
		return nil
	}
}

func effectKeywordsForDescription(effectType models.SpellEffectType) []string {
	switch effectType {
	case models.SpellEffectTypeRestoreLifePartyMember:
		return []string{"heal", "restore", "mend", "renew", "recover", "vitalize"}
	case models.SpellEffectTypeRestoreLifeAllParty:
		return []string{"heal", "restore", "mend", "renew", "allies", "party", "group", "team"}
	case models.SpellEffectTypeApplyBeneficialStatus:
		return []string{"boon", "ward", "fortify", "empower", "enhance", "status", "stance", "focus"}
	case models.SpellEffectTypeRemoveDetrimental:
		return []string{"cleanse", "purge", "remove", "dispel", "clear", "ailment", "debuff", "detrimental"}
	case models.SpellEffectTypeDealDamage:
		return []string{"damage", "harm", "attack", "strike", "blast", "hit", "wound", "burst"}
	default:
		return nil
	}
}

func generatedAbilityNameMatchesEffect(
	name string,
	effectType models.SpellEffectType,
	damageAffinity string,
) bool {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return false
	}
	if !containsAnyKeyword(trimmed, effectKeywordsForName(effectType)) {
		return false
	}
	if effectType != models.SpellEffectTypeDealDamage {
		return true
	}
	normalizedAffinity := models.NormalizeDamageAffinity(damageAffinity)
	if normalizedAffinity == models.DamageAffinityPhysical {
		return true
	}
	return containsAnyKeyword(trimmed, affinityKeywordsForName(string(normalizedAffinity)))
}

func generatedAbilityNameForEffect(
	effectType models.SpellEffectType,
	damageAffinity string,
	abilityType models.SpellAbilityType,
) string {
	switch effectType {
	case models.SpellEffectTypeRestoreLifePartyMember:
		if abilityType == models.SpellAbilityTypeTechnique {
			return "Second Wind"
		}
		return "Mending Touch"
	case models.SpellEffectTypeRestoreLifeAllParty:
		if abilityType == models.SpellAbilityTypeTechnique {
			return "Warband Rally"
		}
		return "Renewal Chorus"
	case models.SpellEffectTypeApplyBeneficialStatus:
		if abilityType == models.SpellAbilityTypeTechnique {
			return "Guarded Stance"
		}
		return "Aegis Boon"
	case models.SpellEffectTypeRemoveDetrimental:
		if abilityType == models.SpellAbilityTypeTechnique {
			return "Cleansing Form"
		}
		return "Purging Rite"
	case models.SpellEffectTypeDealDamage:
		normalizedAffinity := models.NormalizeDamageAffinity(damageAffinity)
		prefix := "Iron"
		suffix := "Strike"
		if abilityType == models.SpellAbilityTypeSpell {
			suffix = "Burst"
		}
		switch normalizedAffinity {
		case models.DamageAffinityFire:
			prefix = "Ember"
			if abilityType == models.SpellAbilityTypeSpell {
				suffix = "Lance"
			} else {
				suffix = "Slash"
			}
		case models.DamageAffinityIce:
			prefix = "Frost"
			if abilityType == models.SpellAbilityTypeSpell {
				suffix = "Shards"
			} else {
				suffix = "Cut"
			}
		case models.DamageAffinityLightning:
			prefix = "Storm"
			if abilityType == models.SpellAbilityTypeSpell {
				suffix = "Bolt"
			} else {
				suffix = "Thrust"
			}
		case models.DamageAffinityPoison:
			prefix = "Venom"
			if abilityType == models.SpellAbilityTypeSpell {
				suffix = "Hex"
			} else {
				suffix = "Sting"
			}
		case models.DamageAffinityArcane:
			prefix = "Arcane"
			if abilityType == models.SpellAbilityTypeSpell {
				suffix = "Burst"
			} else {
				suffix = "Kata"
			}
		case models.DamageAffinityHoly:
			prefix = "Radiant"
			if abilityType == models.SpellAbilityTypeSpell {
				suffix = "Ray"
			} else {
				suffix = "Smite"
			}
		case models.DamageAffinityShadow:
			prefix = "Umbral"
			if abilityType == models.SpellAbilityTypeSpell {
				suffix = "Volley"
			} else {
				suffix = "Rend"
			}
		default:
			if abilityType == models.SpellAbilityTypeSpell {
				prefix = "Force"
				suffix = "Barrage"
			}
		}
		return fmt.Sprintf("%s %s", prefix, suffix)
	default:
		return ""
	}
}

func harmonizeGeneratedAbilityNameWithEffects(
	currentName string,
	abilityType models.SpellAbilityType,
	effects models.SpellEffects,
) string {
	trimmed := strings.TrimSpace(currentName)
	if len(effects) == 0 {
		return trimmed
	}
	primary := effects[0]
	damageAffinity := ""
	if primary.DamageAffinity != nil {
		damageAffinity = strings.TrimSpace(*primary.DamageAffinity)
	}
	if generatedAbilityNameMatchesEffect(trimmed, primary.Type, damageAffinity) {
		return trimmed
	}
	return generatedAbilityNameForEffect(primary.Type, damageAffinity, abilityType)
}

func generatedAbilityDescriptionMatchesEffect(
	description string,
	effectType models.SpellEffectType,
	damageAffinity string,
) bool {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return false
	}
	if !containsAnyKeyword(trimmed, effectKeywordsForDescription(effectType)) {
		return false
	}
	if effectType == models.SpellEffectTypeDealDamage {
		normalizedAffinity := models.NormalizeDamageAffinity(damageAffinity)
		if normalizedAffinity == models.DamageAffinityPhysical {
			return true
		}
		return containsAnyKeyword(trimmed, affinityKeywordsForName(string(normalizedAffinity)))
	}
	return true
}

func generatedAbilityDescriptionForEffect(effect models.SpellEffect, abilityType models.SpellAbilityType) string {
	switch effect.Type {
	case models.SpellEffectTypeRestoreLifePartyMember:
		if abilityType == models.SpellAbilityTypeTechnique {
			return "A practiced combat rhythm that restores health to one ally."
		}
		return "Channels restorative magic to heal one ally."
	case models.SpellEffectTypeRestoreLifeAllParty:
		if abilityType == models.SpellAbilityTypeTechnique {
			return "A commanding rally that restores health across your whole party."
		}
		return "Radiates restorative magic that heals all allies."
	case models.SpellEffectTypeApplyBeneficialStatus:
		statusName := "a beneficial status"
		if len(effect.StatusesToApply) > 0 && strings.TrimSpace(effect.StatusesToApply[0].Name) != "" {
			statusName = effect.StatusesToApply[0].Name
		}
		if abilityType == models.SpellAbilityTypeTechnique {
			return fmt.Sprintf("Adopts a disciplined form that grants %s.", statusName)
		}
		return fmt.Sprintf("Weaves a protective enchantment that grants %s.", statusName)
	case models.SpellEffectTypeRemoveDetrimental:
		statusTarget := "detrimental conditions"
		if len(effect.StatusesToRemove) > 0 {
			statusTarget = strings.Join(effect.StatusesToRemove, ", ")
		}
		if abilityType == models.SpellAbilityTypeTechnique {
			return fmt.Sprintf("Resets footing and clears %s from allies.", statusTarget)
		}
		return fmt.Sprintf("Purges %s from allies.", statusTarget)
	case models.SpellEffectTypeDealDamage:
		affinity := "force"
		switch models.NormalizeDamageAffinity(stringValue(effect.DamageAffinity)) {
		case models.DamageAffinityFire:
			affinity = "fire"
		case models.DamageAffinityIce:
			affinity = "ice"
		case models.DamageAffinityLightning:
			affinity = "lightning"
		case models.DamageAffinityPoison:
			affinity = "poison"
		case models.DamageAffinityArcane:
			affinity = "arcane"
		case models.DamageAffinityHoly:
			affinity = "holy"
		case models.DamageAffinityShadow:
			affinity = "shadow"
		case models.DamageAffinityPhysical:
			affinity = "physical"
		}
		if abilityType == models.SpellAbilityTypeTechnique {
			return fmt.Sprintf("A focused %s strike that damages a single enemy.", affinity)
		}
		return fmt.Sprintf("Unleashes %s energy to damage a single enemy.", affinity)
	default:
		return ""
	}
}

func harmonizeGeneratedAbilityDescriptionWithEffects(
	currentDescription string,
	abilityType models.SpellAbilityType,
	effects models.SpellEffects,
) string {
	trimmed := strings.TrimSpace(currentDescription)
	if len(effects) == 0 {
		return trimmed
	}
	primary := effects[0]
	damageAffinity := ""
	if primary.DamageAffinity != nil {
		damageAffinity = strings.TrimSpace(*primary.DamageAffinity)
	}
	if generatedAbilityDescriptionMatchesEffect(trimmed, primary.Type, damageAffinity) {
		return trimmed
	}
	replacement := strings.TrimSpace(generatedAbilityDescriptionForEffect(primary, abilityType))
	if replacement != "" {
		return replacement
	}
	if trimmed != "" {
		return trimmed
	}
	return buildGeneratedAbilityEffectText(effects, abilityType)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
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
		damageAffinity := inferGeneratedDamageAffinity(text, abilityType)
		return models.SpellEffects{{
			Type:           models.SpellEffectTypeDealDamage,
			Amount:         damage,
			DamageAffinity: &damageAffinity,
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
			damageLabel := "damage"
			if effect.DamageAffinity != nil && strings.TrimSpace(*effect.DamageAffinity) != "" {
				damageLabel = fmt.Sprintf("%s damage", strings.TrimSpace(*effect.DamageAffinity))
			}
			if effect.Amount > 0 {
				parts = append(parts, fmt.Sprintf("Deals %d %s.", effect.Amount, damageLabel))
			} else {
				parts = append(parts, fmt.Sprintf("Deals direct %s.", damageLabel))
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
