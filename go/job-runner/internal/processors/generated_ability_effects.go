package processors

import (
	"fmt"
	"hash/fnv"
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

type generatedMonsterCurveProfile struct {
	EstimatedHealth        int
	EstimatedDamagePerTurn int
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
		return []string{"heal", "mend", "renew", "restore", "restoration", "recovery", "vital"}
	case models.SpellEffectTypeRestoreLifeAllParty:
		return []string{"heal", "renew", "restore", "restoration", "chorus", "rally", "aura", "all"}
	case models.SpellEffectTypeApplyBeneficialStatus:
		return []string{"aegis", "ward", "guard", "boon", "fortify", "stance", "focus"}
	case models.SpellEffectTypeRemoveDetrimental:
		return []string{"cleanse", "purge", "dispel", "purify", "clear"}
	case models.SpellEffectTypeDealDamage:
		return []string{
			"strike", "lance", "bolt", "blast", "slash", "rend", "burst", "volley",
			"spear", "javelin", "assault", "pounce", "riposte", "ray", "thrust", "smite", "cut", "barrage",
		}
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

func stableNameVariantIndex(seed string, size int) int {
	if size <= 1 {
		return 0
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(strings.TrimSpace(seed)))
	return int(hasher.Sum32() % uint32(size))
}

func pickGeneratedAbilityNameVariant(seed string, options []string) string {
	if len(options) == 0 {
		return ""
	}
	idx := stableNameVariantIndex(seed, len(options))
	return options[idx]
}

func generatedAbilityNameForEffect(
	effect models.SpellEffect,
	abilityType models.SpellAbilityType,
	seed string,
) string {
	seedBase := strings.TrimSpace(seed)
	if seedBase == "" {
		seedBase = "generated ability"
	}

	switch effect.Type {
	case models.SpellEffectTypeRestoreLifePartyMember:
		if abilityType == models.SpellAbilityTypeTechnique {
			return pickGeneratedAbilityNameVariant(seedBase, []string{
				"Mending Riposte",
				"Recovery Stance",
				"Renewing Step",
				"Healing Cadence",
			})
		}
		return pickGeneratedAbilityNameVariant(seedBase, []string{
			"Mending Touch",
			"Renewal Weave",
			"Healing Pulse",
			"Restoration Sigil",
		})
	case models.SpellEffectTypeRestoreLifeAllParty:
		if abilityType == models.SpellAbilityTypeTechnique {
			return pickGeneratedAbilityNameVariant(seedBase, []string{
				"Healing Rally",
				"Renewal Formation",
				"Recovery Rally",
				"Mending Phalanx",
			})
		}
		return pickGeneratedAbilityNameVariant(seedBase, []string{
			"Renewal Chorus",
			"Healing Aura",
			"Restoring Chorus",
			"Group Mend",
		})
	case models.SpellEffectTypeApplyBeneficialStatus:
		if abilityType == models.SpellAbilityTypeTechnique {
			return pickGeneratedAbilityNameVariant(seedBase, []string{
				"Guarded Stance",
				"Fortify Form",
				"Warden Focus",
				"Steadfast Guard",
			})
		}
		return pickGeneratedAbilityNameVariant(seedBase, []string{
			"Aegis Boon",
			"Fortifying Ward",
			"Guarding Sigil",
			"Warden Blessing",
		})
	case models.SpellEffectTypeRemoveDetrimental:
		if abilityType == models.SpellAbilityTypeTechnique {
			return pickGeneratedAbilityNameVariant(seedBase, []string{
				"Cleansing Form",
				"Purging Step",
				"Dispel Stance",
				"Purify Rhythm",
			})
		}
		return pickGeneratedAbilityNameVariant(seedBase, []string{
			"Purging Rite",
			"Cleansing Weave",
			"Dispel Invocation",
			"Purify Glyph",
		})
	case models.SpellEffectTypeDealDamage:
		normalizedAffinity := models.NormalizeDamageAffinity(stringValue(effect.DamageAffinity))
		damageSeed := fmt.Sprintf("%s|%s|%s|%d", seedBase, effect.Type, normalizedAffinity, effect.Amount)
		switch normalizedAffinity {
		case models.DamageAffinityFire:
			if abilityType == models.SpellAbilityTypeSpell {
				return pickGeneratedAbilityNameVariant(damageSeed, []string{
					"Ember Lance",
					"Cinder Bolt",
					"Inferno Burst",
					"Flame Spear",
				})
			}
			return pickGeneratedAbilityNameVariant(damageSeed, []string{
				"Ember Slash",
				"Cinder Strike",
				"Flame Rend",
				"Pyre Cut",
			})
		case models.DamageAffinityIce:
			if abilityType == models.SpellAbilityTypeSpell {
				return pickGeneratedAbilityNameVariant(damageSeed, []string{
					"Frost Shards",
					"Glacial Lance",
					"Rime Burst",
					"Chill Bolt",
				})
			}
			return pickGeneratedAbilityNameVariant(damageSeed, []string{
				"Frost Cut",
				"Glacial Strike",
				"Rime Rend",
				"Chill Slash",
			})
		case models.DamageAffinityLightning:
			if abilityType == models.SpellAbilityTypeSpell {
				return pickGeneratedAbilityNameVariant(damageSeed, []string{
					"Storm Bolt",
					"Thunder Lance",
					"Volt Burst",
					"Spark Javelin",
				})
			}
			return pickGeneratedAbilityNameVariant(damageSeed, []string{
				"Storm Thrust",
				"Thunder Strike",
				"Volt Rend",
				"Spark Slash",
			})
		case models.DamageAffinityPoison:
			if abilityType == models.SpellAbilityTypeSpell {
				return pickGeneratedAbilityNameVariant(damageSeed, []string{
					"Venom Bolt",
					"Toxin Burst",
					"Blight Lance",
					"Virulent Spear",
				})
			}
			return pickGeneratedAbilityNameVariant(damageSeed, []string{
				"Venom Sting",
				"Toxin Strike",
				"Blight Rend",
				"Virulent Cut",
			})
		case models.DamageAffinityArcane:
			if abilityType == models.SpellAbilityTypeSpell {
				return pickGeneratedAbilityNameVariant(damageSeed, []string{
					"Arcane Burst",
					"Rune Volley",
					"Astral Lance",
					"Mana Barrage",
				})
			}
			return pickGeneratedAbilityNameVariant(damageSeed, []string{
				"Arcane Strike",
				"Rune Thrust",
				"Astral Rend",
				"Mana Cut",
			})
		case models.DamageAffinityHoly:
			if abilityType == models.SpellAbilityTypeSpell {
				return pickGeneratedAbilityNameVariant(damageSeed, []string{
					"Radiant Ray",
					"Dawn Lance",
					"Sanctified Burst",
					"Sunfire Bolt",
				})
			}
			return pickGeneratedAbilityNameVariant(damageSeed, []string{
				"Radiant Smite",
				"Dawn Strike",
				"Sanctified Rend",
				"Sunfire Cut",
			})
		case models.DamageAffinityShadow:
			if abilityType == models.SpellAbilityTypeSpell {
				return pickGeneratedAbilityNameVariant(damageSeed, []string{
					"Umbral Volley",
					"Night Bolt",
					"Void Lance",
					"Hexfire Burst",
				})
			}
			return pickGeneratedAbilityNameVariant(damageSeed, []string{
				"Umbral Rend",
				"Night Strike",
				"Void Slash",
				"Hexfire Cut",
			})
		default:
			if abilityType == models.SpellAbilityTypeSpell {
				return pickGeneratedAbilityNameVariant(damageSeed, []string{
					"Force Barrage",
					"Iron Bolt",
					"Battle Lance",
					"Impact Burst",
				})
			}
			return pickGeneratedAbilityNameVariant(damageSeed, []string{
				"Iron Strike",
				"Battle Rend",
				"Impact Slash",
				"Steel Thrust",
			})
		}
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
	return generatedAbilityNameForEffect(primary, abilityType, trimmed)
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
	curve := estimateMonsterCurveForTargetLevel(targetLevel)

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
		if curve != nil {
			curveRatio := 0.08
			if abilityType == models.SpellAbilityTypeTechnique {
				curveRatio = 0.07
			}
			manaFactor := math.Min(1, float64(maxInt(0, manaCost))/60.0)
			curveRatio += (0.06 * manaFactor)
			curveAmount := int(math.Round(float64(curve.EstimatedHealth) * curveRatio))
			if curve.EstimatedDamagePerTurn > 0 {
				// Single-target healing should recover most of one expected incoming hit.
				curveAmount = maxInt(curveAmount, int(math.Round(float64(curve.EstimatedDamagePerTurn)*0.75)))
			}
			amount = blendGeneratedAbilityAmount(amount, curveAmount, 0.6)
		}
		amount = varyGeneratedAbilityAmount(amount, text+"|restore_life_party_member|"+string(abilityType), effectType)
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
		if curve != nil {
			curveRatio := 0.05
			if abilityType == models.SpellAbilityTypeTechnique {
				curveRatio = 0.045
			}
			manaFactor := math.Min(1, float64(maxInt(0, manaCost))/60.0)
			curveRatio += (0.05 * manaFactor)
			curveAmount := int(math.Round(float64(curve.EstimatedHealth) * curveRatio))
			if curve.EstimatedDamagePerTurn > 0 {
				// Group healing is intentionally smaller per target than single-target healing.
				curveAmount = maxInt(curveAmount, int(math.Round(float64(curve.EstimatedDamagePerTurn)*0.45)))
			}
			amount = blendGeneratedAbilityAmount(amount, curveAmount, 0.6)
		}
		amount = varyGeneratedAbilityAmount(amount, text+"|restore_life_all_party_members|"+string(abilityType), effectType)
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
		if curve != nil {
			curveRatio := 0.10
			if abilityType == models.SpellAbilityTypeTechnique {
				curveRatio = 0.085
				if containsAnyKeyword(text, []string{"heavy", "crush", "breaker", "slam", "assault"}) {
					curveRatio += 0.02
				}
			}
			manaFactor := math.Min(1, float64(maxInt(0, manaCost))/60.0)
			curveRatio += (0.08 * manaFactor)
			curveAmount := int(math.Round(float64(curve.EstimatedHealth) * curveRatio))
			damage = blendGeneratedAbilityAmount(damage, curveAmount, 0.65)
		}
		damage = varyGeneratedAbilityAmount(damage, text+"|deal_damage|"+string(abilityType), effectType)
		damageAffinity := inferGeneratedDamageAffinity(text, abilityType)
		return models.SpellEffects{{
			Type:           models.SpellEffectTypeDealDamage,
			Amount:         damage,
			DamageAffinity: &damageAffinity,
		}}
	}
}

func estimateMonsterCurveForTargetLevel(targetLevel *int) *generatedMonsterCurveProfile {
	level, ok := normalizeAbilityTargetLevel(targetLevel)
	if !ok {
		return nil
	}
	// Mirror monster progression formulas in pkg/models/monster.go with midline template baselines.
	// Effective constitution drives max health: max_health = constitution * 10.
	baseConstitution := 12
	baseStrength := 11
	baseDexterity := 11
	effectiveConstitution := maxInt(1, baseConstitution+level-1)
	effectiveStrength := maxInt(1, baseStrength+level-1)
	effectiveDexterity := maxInt(1, baseDexterity+level-1)
	estimatedHealth := effectiveConstitution * 10

	// Fallback monster attack profile:
	// damage_min = strength/3 + level/2
	// damage_max = damage_min + 2 + dexterity/5
	damageMin := maxInt(1, (effectiveStrength/3)+(level/2))
	damageMax := maxInt(damageMin, damageMin+2+(effectiveDexterity/5))
	estimatedDamagePerTurn := int(math.Round(float64(damageMin+damageMax) / 2.0))
	if estimatedDamagePerTurn < 1 {
		estimatedDamagePerTurn = 1
	}
	return &generatedMonsterCurveProfile{
		EstimatedHealth:        estimatedHealth,
		EstimatedDamagePerTurn: estimatedDamagePerTurn,
	}
}

func blendGeneratedAbilityAmount(legacyAmount int, curveAmount int, curveWeight float64) int {
	if curveAmount <= 0 {
		if legacyAmount < 1 {
			return 1
		}
		return legacyAmount
	}
	if legacyAmount <= 0 {
		return maxInt(1, curveAmount)
	}
	if curveWeight < 0 {
		curveWeight = 0
	}
	if curveWeight > 1 {
		curveWeight = 1
	}
	legacyWeight := 1.0 - curveWeight
	blended := int(math.Round((float64(legacyAmount) * legacyWeight) + (float64(curveAmount) * curveWeight)))
	if blended < 1 {
		return 1
	}
	return blended
}

func generatedAmountVarianceRatio(seed string) float64 {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(strings.TrimSpace(seed)))
	// Map to [-1.0, 1.0] deterministically.
	bucket := float64(hasher.Sum32()%1001) / 1000.0
	return (bucket * 2.0) - 1.0
}

func varyGeneratedAbilityAmount(base int, seed string, effectType models.SpellEffectType) int {
	if base <= 1 {
		return base
	}
	spread := 0.08
	switch effectType {
	case models.SpellEffectTypeDealDamage:
		spread = 0.14
	case models.SpellEffectTypeRestoreLifePartyMember, models.SpellEffectTypeRestoreLifeAllParty:
		spread = 0.10
	}
	delta := int(math.Round(float64(base) * spread * generatedAmountVarianceRatio(seed)))
	value := base + delta
	if value < 1 {
		return 1
	}
	return value
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
