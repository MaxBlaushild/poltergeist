package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type seededAbilityPackItem struct {
	Name        string `json:"name"`
	AbilityType string `json:"abilityType"`
	Action      string `json:"action"`
}

type seededAbilityPackResponse struct {
	AbilityType    string                  `json:"abilityType"`
	ProcessedCount int                     `json:"processedCount"`
	CreatedCount   int                     `json:"createdCount"`
	UpdatedCount   int                     `json:"updatedCount"`
	Items          []seededAbilityPackItem `json:"items"`
}

func (s *server) seedSpellCombatPack(ctx *gin.Context) {
	s.seedCombatAbilityPack(ctx, models.SpellAbilityTypeSpell)
}

func (s *server) seedTechniqueCombatPack(ctx *gin.Context) {
	s.seedCombatAbilityPack(ctx, models.SpellAbilityTypeTechnique)
}

func (s *server) seedCombatAbilityPack(ctx *gin.Context, abilityType models.SpellAbilityType) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	requests := combatAbilitySeedPackRequests(abilityType)
	existingSpells, err := s.dbClient.Spell().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	existingByKey := make(map[string]models.Spell, len(existingSpells))
	for _, existing := range existingSpells {
		if normalizeSpellAbilityType(string(existing.AbilityType)) != abilityType {
			continue
		}
		key := seededAbilityKey(existing.Name, abilityType)
		if key == "" {
			continue
		}
		existingByKey[key] = existing
	}

	response := seededAbilityPackResponse{
		AbilityType: string(abilityType),
		Items:       make([]seededAbilityPackItem, 0, len(requests)),
	}

	for _, request := range requests {
		spell, err := s.parseSpellUpsertRequest(request, 1)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		imageStatus := models.SpellImageGenerationStatusNone
		var imageError interface{} = nil
		if strings.TrimSpace(spell.IconURL) != "" {
			imageStatus = models.SpellImageGenerationStatusComplete
			clearErr := ""
			imageError = &clearErr
		}

		key := seededAbilityKey(spell.Name, abilityType)
		if existing, ok := existingByKey[key]; ok {
			if err := s.dbClient.Spell().Update(ctx, existing.ID, map[string]interface{}{
				"name":                    spell.Name,
				"description":             spell.Description,
				"icon_url":                spell.IconURL,
				"ability_type":            spell.AbilityType,
				"ability_level":           spell.AbilityLevel,
				"cooldown_turns":          spell.CooldownTurns,
				"effect_text":             spell.EffectText,
				"school_of_magic":         spell.SchoolOfMagic,
				"mana_cost":               spell.ManaCost,
				"effects":                 spell.Effects,
				"image_generation_status": imageStatus,
				"image_generation_error":  imageError,
			}); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			response.UpdatedCount++
			response.Items = append(response.Items, seededAbilityPackItem{
				Name:        spell.Name,
				AbilityType: string(abilityType),
				Action:      "updated",
			})
			response.ProcessedCount++
			continue
		}

		if strings.TrimSpace(spell.IconURL) != "" {
			spell.ImageGenerationStatus = models.SpellImageGenerationStatusComplete
			clearErr := ""
			spell.ImageGenerationError = &clearErr
		} else {
			spell.ImageGenerationStatus = models.SpellImageGenerationStatusNone
			spell.ImageGenerationError = nil
		}
		if err := s.dbClient.Spell().Create(ctx, spell); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		existingByKey[key] = *spell
		response.CreatedCount++
		response.Items = append(response.Items, seededAbilityPackItem{
			Name:        spell.Name,
			AbilityType: string(abilityType),
			Action:      "created",
		})
		response.ProcessedCount++
	}

	ctx.JSON(http.StatusOK, response)
}

func seededAbilityKey(name string, abilityType models.SpellAbilityType) string {
	trimmed := strings.TrimSpace(strings.ToLower(name))
	if trimmed == "" {
		return ""
	}
	return string(abilityType) + "|" + trimmed
}

func combatAbilitySeedPackRequests(abilityType models.SpellAbilityType) []spellUpsertRequest {
	if abilityType == models.SpellAbilityTypeTechnique {
		return combatTechniqueSeedPackRequests()
	}
	return combatSpellSeedPackRequests()
}

func combatSpellSeedPackRequests() []spellUpsertRequest {
	return []spellUpsertRequest{
		seededSpell("Ember Bolt", 5, 0, 16, "Pyromancy", "Fire a compact lance of flame into one foe.", "Deal 50 fire damage to one enemy.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 50, 1, "fire"),
		}),
		seededSpell("Frost Needle", 7, 0, 18, "Cryomancy", "Drive a shard of winter into an enemy and slow their reaction speed.", "Deal 70 ice damage and apply Chilled.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 70, 1, "ice"),
			singleDetrimentalStatusEffect(statusChilled(2, 45)),
		}),
		seededSpell("Venom Dart", 8, 0, 20, "Toxicology", "Sink a poisoned bolt into a target and let the toxin work.", "Deal 80 poison damage and apply Poisoned.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 80, 1, "poison"),
			singleDetrimentalStatusEffect(statusPoisoned(6, 60)),
		}),
		seededSpell("Arcane Lance", 10, 0, 22, "Arcane", "Loose a stable spear of raw arcane force.", "Deal 100 arcane damage to one enemy.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 100, 1, "arcane"),
		}),
		seededSpell("Radiant Spear", 12, 0, 24, "Radiance", "A focused shaft of light strikes with clean holy force.", "Deal 120 holy damage to one enemy.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 120, 1, "holy"),
		}),
		seededSpell("Shadow Brand", 14, 0, 24, "Umbramancy", "Brand a foe with enervating shadow.", "Deal 140 shadow damage and apply Weakened.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 140, 1, "shadow"),
			singleDetrimentalStatusEffect(statusWeakened(3, 45)),
		}),
		seededSpell("Cinder Rain", 16, 1, 36, "Pyromancy", "A shower of embers scorches every enemy on the field.", "Deal 96 fire damage to all enemies.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamageAllEnemies, 96, 1, "fire"),
		}),
		seededSpell("Static Bloom", 18, 1, 40, "Stormcalling", "A burst of living static lashes out across the enemy line.", "Deal 108 lightning damage to all enemies.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamageAllEnemies, 108, 1, "lightning"),
		}),
		seededSpell("Winter Veil", 20, 2, 36, "Cryomancy", "A freezing curtain settles over the battlefield and steals momentum.", "Deal 120 ice damage to all enemies and apply Chilled.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamageAllEnemies, 120, 1, "ice"),
			allDetrimentalStatusEffect(statusChilled(2, 45)),
		}),
		seededSpell("Mend Wounds", 6, 0, 16, "Restoration", "Restore a meaningful chunk of vitality to one ally.", "Restore 28 health to one ally.", []spellEffectPayload{
			healEffect(models.SpellEffectTypeRestoreLifePartyMember, 28),
		}),
		seededSpell("Renewing Surge", 15, 2, 32, "Restoration", "A wash of curative energy restores the whole party.", "Restore 18 health to all allies.", []spellEffectPayload{
			healEffect(models.SpellEffectTypeRestoreLifeAllParty, 18),
		}),
		seededSpell("Second Breath", 18, 2, 40, "Restoration", "Call a fallen ally back into the fight.", "Revive one ally with 35 health.", []spellEffectPayload{
			healEffect(models.SpellEffectTypeRevivePartyMember, 35),
		}),
		seededSpell("Phoenix Oath", 35, 5, 72, "Radiance", "A rare vow of fire and light rekindles every fallen companion.", "Revive all downed allies with 25 health.", []spellEffectPayload{
			healEffect(models.SpellEffectTypeReviveAllDownedParty, 25),
		}),
		seededSpell("Bastion Ward", 10, 1, 20, "Abjuration", "Wrap one ally in a stout defensive ward.", "Apply Guarded to one ally.", []spellEffectPayload{
			beneficialStatusEffect(statusGuarded(4, 60)),
		}),
		seededSpell("Fleet Invocation", 12, 1, 20, "Windcraft", "Sharpen an ally's reflexes with a gust of quickening force.", "Apply Quickened to one ally.", []spellEffectPayload{
			beneficialStatusEffect(statusQuickened(4, 60)),
		}),
		seededSpell("Meditative Current", 14, 1, 24, "Mysticism", "A calm current of focus restores magical rhythm.", "Apply Clarity to one ally.", []spellEffectPayload{
			beneficialStatusEffect(statusClarity(6, 60)),
		}),
		seededSpell("Regrowth Prayer", 16, 1, 28, "Verdancy", "Invite living magic to slowly mend an ally's wounds.", "Apply Regenerating to one ally.", []spellEffectPayload{
			beneficialStatusEffect(statusRegenerating(8, 60)),
		}),
		seededSpell("Hex of Frailty", 11, 1, 20, "Hexcraft", "Sap an enemy's sturdiness with a brittle curse.", "Apply Sundered to one enemy.", []spellEffectPayload{
			singleDetrimentalStatusEffect(statusSundered(4, 45)),
		}),
		seededSpell("Shock Sigil", 13, 1, 24, "Stormcalling", "A crackling sigil scrambles the target's concentration.", "Deal 130 lightning damage and apply Shocked.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 130, 1, "lightning"),
			singleDetrimentalStatusEffect(statusShocked(3, 45)),
		}),
		seededSpell("Purge", 12, 1, 18, "Purification", "Strip away hostile lingering effects from one ally.", "Remove common detrimental statuses from one ally.", []spellEffectPayload{
			removeStatusesEffect("Burning", "Poisoned", "Chilled", "Shocked", "Weakened", "Sundered", "Off-Balance"),
		}),
	}
}

func combatTechniqueSeedPackRequests() []spellUpsertRequest {
	return []spellUpsertRequest{
		seededTechnique("Precise Thrust", 3, 0, "Martial", "Drive a narrow opening strike through a target's guard.", "Deal 24 piercing damage to one enemy.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 24, 1, "piercing"),
		}),
		seededTechnique("Hamstring Cut", 7, 1, "Martial", "Slice low and rob an enemy of mobility.", "Deal 56 slashing damage and apply Off-Balance.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 56, 1, "slashing"),
			singleDetrimentalStatusEffect(statusOffBalance(2, 45)),
		}),
		seededTechnique("Shield Bash", 9, 1, "Martial", "Crash a shield into an enemy and knock the strength out of them.", "Deal 72 bludgeoning damage and apply Weakened.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 72, 1, "bludgeoning"),
			singleDetrimentalStatusEffect(statusWeakened(2, 45)),
		}),
		seededTechnique("Whirlwind Sweep", 14, 2, "Martial", "A spinning cut catches every nearby foe.", "Deal 70 slashing damage to all enemies.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamageAllEnemies, 70, 1, "slashing"),
		}),
		seededTechnique("Rally Stance", 8, 1, "Martial", "Adopt a commanding stance that bolsters a comrade.", "Apply Blessed to one ally.", []spellEffectPayload{
			beneficialStatusEffect(statusBlessed(2, 2, 60)),
		}),
		seededTechnique("Iron Guard", 10, 1, "Martial", "Set a defensive line and harden an ally's posture.", "Apply Guarded to one ally.", []spellEffectPayload{
			beneficialStatusEffect(statusGuarded(3, 60)),
		}),
		seededTechnique("First Aid", 6, 1, "Martial", "Use a practiced battlefield treatment to patch an ally up.", "Restore 18 health to one ally.", []spellEffectPayload{
			healEffect(models.SpellEffectTypeRestoreLifePartyMember, 18),
		}),
		seededTechnique("Adrenal Surge", 12, 1, "Martial", "Drive an ally into a sharper, faster combat rhythm.", "Apply Quickened to one ally.", []spellEffectPayload{
			beneficialStatusEffect(statusQuickened(3, 60)),
		}),
		seededTechnique("Expose Weakness", 13, 1, "Martial", "Read the target's defense and tear it open.", "Apply Sundered to one enemy.", []spellEffectPayload{
			singleDetrimentalStatusEffect(statusSundered(3, 45)),
		}),
		seededTechnique("Shake It Off", 10, 1, "Martial", "Force the body back under control and shrug off hindrances.", "Remove common detrimental statuses from one ally.", []spellEffectPayload{
			removeStatusesEffect("Burning", "Poisoned", "Chilled", "Shocked", "Weakened", "Sundered", "Off-Balance"),
		}),
		seededTechnique("Execution Flurry", 18, 2, "Martial", "A disciplined burst of chained strikes punishes a single foe.", "Deal 48 slashing damage to one enemy 3 times.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamage, 48, 3, "slashing"),
		}),
		seededTechnique("Overrun", 25, 3, "Martial", "Drive through the enemy line with a crushing advance.", "Deal 63 bludgeoning damage to all enemies 2 times.", []spellEffectPayload{
			damageEffect(models.SpellEffectTypeDealDamageAllEnemies, 63, 2, "bludgeoning"),
		}),
	}
}

func seededSpell(
	name string,
	level int,
	cooldownTurns int,
	manaCost int,
	school string,
	description string,
	effectText string,
	effects []spellEffectPayload,
) spellUpsertRequest {
	levelCopy := level
	return spellUpsertRequest{
		Name:          name,
		Description:   description,
		AbilityType:   string(models.SpellAbilityTypeSpell),
		AbilityLevel:  &levelCopy,
		CooldownTurns: cooldownTurns,
		EffectText:    effectText,
		SchoolOfMagic: school,
		ManaCost:      manaCost,
		Effects:       effects,
	}
}

func seededTechnique(
	name string,
	level int,
	cooldownTurns int,
	school string,
	description string,
	effectText string,
	effects []spellEffectPayload,
) spellUpsertRequest {
	levelCopy := level
	return spellUpsertRequest{
		Name:          name,
		Description:   description,
		AbilityType:   string(models.SpellAbilityTypeTechnique),
		AbilityLevel:  &levelCopy,
		CooldownTurns: cooldownTurns,
		EffectText:    effectText,
		SchoolOfMagic: school,
		ManaCost:      0,
		Effects:       effects,
	}
}

func damageEffect(effectType models.SpellEffectType, amount int, hits int, affinity string) spellEffectPayload {
	payload := spellEffectPayload{
		Type:   string(effectType),
		Amount: intPtr(amount),
		Hits:   intPtr(hits),
	}
	if trimmed := strings.TrimSpace(affinity); trimmed != "" {
		payload.DamageAffinity = stringPtr(trimmed)
	}
	return payload
}

func healEffect(effectType models.SpellEffectType, amount int) spellEffectPayload {
	return spellEffectPayload{
		Type:   string(effectType),
		Amount: intPtr(amount),
	}
}

func beneficialStatusEffect(status scenarioFailureStatusPayload) spellEffectPayload {
	return spellEffectPayload{
		Type:            string(models.SpellEffectTypeApplyBeneficialStatus),
		StatusesToApply: []scenarioFailureStatusPayload{status},
	}
}

func singleDetrimentalStatusEffect(status scenarioFailureStatusPayload) spellEffectPayload {
	return spellEffectPayload{
		Type:            string(models.SpellEffectTypeApplyDetrimentalStatus),
		StatusesToApply: []scenarioFailureStatusPayload{status},
	}
}

func allDetrimentalStatusEffect(status scenarioFailureStatusPayload) spellEffectPayload {
	return spellEffectPayload{
		Type:            string(models.SpellEffectTypeApplyDetrimentalAll),
		StatusesToApply: []scenarioFailureStatusPayload{status},
	}
}

func removeStatusesEffect(names ...string) spellEffectPayload {
	return spellEffectPayload{
		Type:             string(models.SpellEffectTypeRemoveDetrimental),
		StatusesToRemove: names,
	}
}

func statusBurning(damagePerTick int, durationSeconds int) scenarioFailureStatusPayload {
	return detrimentalStatus("Burning", "Flames cling to the target.", "Suffers fire damage over time.", "damage_over_time", durationSeconds, 0, 0, 0, damagePerTick)
}

func statusPoisoned(damagePerTick int, durationSeconds int) scenarioFailureStatusPayload {
	return detrimentalStatus("Poisoned", "Venom works through the target's system.", "Suffers poison damage over time.", "damage_over_time", durationSeconds, 0, 0, 0, damagePerTick)
}

func statusChilled(dexPenalty int, durationSeconds int) scenarioFailureStatusPayload {
	return detrimentalStatus("Chilled", "Cold slows every motion.", "Dexterity is reduced.", "stat_modifier", durationSeconds, 0, -dexPenalty, 0, 0)
}

func statusShocked(intPenalty int, durationSeconds int) scenarioFailureStatusPayload {
	return detrimentalStatus("Shocked", "Sparking nerves disrupt concentration.", "Intelligence and wisdom are reduced.", "stat_modifier", durationSeconds, 0, 0, 0, 0,
		func(status *scenarioFailureStatusPayload) {
			status.IntelligenceMod = -intPenalty
			status.WisdomMod = -intPenalty
		},
	)
}

func statusWeakened(strPenalty int, durationSeconds int) scenarioFailureStatusPayload {
	return detrimentalStatus("Weakened", "The target's striking power falters.", "Strength is reduced.", "stat_modifier", durationSeconds, -strPenalty, 0, 0, 0)
}

func statusSundered(conPenalty int, durationSeconds int) scenarioFailureStatusPayload {
	return detrimentalStatus("Sundered", "The target's defenses are compromised.", "Constitution is reduced.", "stat_modifier", durationSeconds, 0, 0, -conPenalty, 0)
}

func statusOffBalance(dexPenalty int, durationSeconds int) scenarioFailureStatusPayload {
	return detrimentalStatus("Off-Balance", "The target struggles to reset their footing.", "Dexterity is reduced.", "stat_modifier", durationSeconds, 0, -dexPenalty, 0, 0)
}

func statusGuarded(conBonus int, durationSeconds int) scenarioFailureStatusPayload {
	return beneficialStatus("Guarded", "Protective discipline hardens the body.", "Constitution is increased.", "stat_modifier", durationSeconds, 0, 0, conBonus, 0)
}

func statusQuickened(dexBonus int, durationSeconds int) scenarioFailureStatusPayload {
	return beneficialStatus("Quickened", "A burst of speed sharpens reactions.", "Dexterity is increased.", "stat_modifier", durationSeconds, 0, dexBonus, 0, 0)
}

func statusClarity(manaPerTick int, durationSeconds int) scenarioFailureStatusPayload {
	return beneficialStatus("Clarity", "A steady current restores magical composure.", "Mana is restored over time.", "mana_over_time", durationSeconds, 0, 0, 0, 0,
		func(status *scenarioFailureStatusPayload) {
			status.ManaPerTick = manaPerTick
		},
	)
}

func statusRegenerating(healthPerTick int, durationSeconds int) scenarioFailureStatusPayload {
	return beneficialStatus("Regenerating", "Living magic steadily knits wounds closed.", "Health is restored over time.", "health_over_time", durationSeconds, 0, 0, 0, 0,
		func(status *scenarioFailureStatusPayload) {
			status.HealthPerTick = healthPerTick
		},
	)
}

func statusBlessed(strBonus int, wisBonus int, durationSeconds int) scenarioFailureStatusPayload {
	return beneficialStatus("Blessed", "Resolve and instinct rise together.", "Strength and wisdom are increased.", "stat_modifier", durationSeconds, strBonus, 0, 0, 0,
		func(status *scenarioFailureStatusPayload) {
			status.WisdomMod = wisBonus
		},
	)
}

func beneficialStatus(
	name string,
	description string,
	effect string,
	effectType string,
	durationSeconds int,
	strengthMod int,
	dexterityMod int,
	constitutionMod int,
	damagePerTick int,
	mutators ...func(*scenarioFailureStatusPayload),
) scenarioFailureStatusPayload {
	positive := true
	status := scenarioFailureStatusPayload{
		Name:            name,
		Description:     description,
		Effect:          effect,
		EffectType:      effectType,
		Positive:        &positive,
		DamagePerTick:   damagePerTick,
		DurationSeconds: durationSeconds,
		StrengthMod:     strengthMod,
		DexterityMod:    dexterityMod,
		ConstitutionMod: constitutionMod,
	}
	for _, mutate := range mutators {
		mutate(&status)
	}
	return status
}

func detrimentalStatus(
	name string,
	description string,
	effect string,
	effectType string,
	durationSeconds int,
	strengthMod int,
	dexterityMod int,
	constitutionMod int,
	damagePerTick int,
	mutators ...func(*scenarioFailureStatusPayload),
) scenarioFailureStatusPayload {
	positive := false
	status := scenarioFailureStatusPayload{
		Name:            name,
		Description:     description,
		Effect:          effect,
		EffectType:      effectType,
		Positive:        &positive,
		DamagePerTick:   damagePerTick,
		DurationSeconds: durationSeconds,
		StrengthMod:     strengthMod,
		DexterityMod:    dexterityMod,
		ConstitutionMod: constitutionMod,
	}
	for _, mutate := range mutators {
		mutate(&status)
	}
	return status
}
