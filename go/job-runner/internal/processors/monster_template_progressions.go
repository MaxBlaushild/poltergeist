package processors

import (
	"context"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type monsterTemplateProgressionCandidate struct {
	ProgressionID     uuid.UUID
	Name              string
	AbilityType       models.SpellAbilityType
	MemberCount       int
	DirectDamage      bool
	AreaDamage        bool
	Healing           bool
	BeneficialStatus  bool
	DetrimentalStatus bool
	AffinityCounts    map[models.DamageAffinity]int
}

func loadMonsterTemplateProgressionCandidates(
	ctx context.Context,
	dbClient db.DbClient,
) ([]monsterTemplateProgressionCandidate, error) {
	spells, err := dbClient.Spell().FindAll(ctx)
	if err != nil {
		return nil, err
	}

	candidatesByID := make(map[uuid.UUID]*monsterTemplateProgressionCandidate)
	for _, spell := range spells {
		for _, link := range spell.ProgressionLinks {
			if link.ProgressionID == uuid.Nil {
				continue
			}
			candidate, exists := candidatesByID[link.ProgressionID]
			if !exists {
				name := strings.TrimSpace(link.Progression.Name)
				if name == "" {
					name = strings.TrimSpace(spell.Name)
				}
				abilityType := link.Progression.AbilityType
				if abilityType == "" {
					abilityType = spell.AbilityType
				}
				if abilityType == "" {
					abilityType = models.SpellAbilityTypeSpell
				}
				candidate = &monsterTemplateProgressionCandidate{
					ProgressionID:  link.ProgressionID,
					Name:           name,
					AbilityType:    abilityType,
					AffinityCounts: make(map[models.DamageAffinity]int),
				}
				candidatesByID[link.ProgressionID] = candidate
			}
			mergeSpellIntoProgressionCandidate(candidate, &spell)
		}
	}

	candidates := make([]monsterTemplateProgressionCandidate, 0, len(candidatesByID))
	for _, candidate := range candidatesByID {
		candidates = append(candidates, *candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].AbilityType != candidates[j].AbilityType {
			return candidates[i].AbilityType < candidates[j].AbilityType
		}
		if candidates[i].MemberCount != candidates[j].MemberCount {
			return candidates[i].MemberCount > candidates[j].MemberCount
		}
		return strings.ToLower(candidates[i].Name) < strings.ToLower(candidates[j].Name)
	})
	return candidates, nil
}

func mergeSpellIntoProgressionCandidate(candidate *monsterTemplateProgressionCandidate, spell *models.Spell) {
	if candidate == nil || spell == nil {
		return
	}
	candidate.MemberCount++
	for _, effect := range spell.Effects {
		switch effect.Type {
		case models.SpellEffectTypeDealDamage:
			candidate.DirectDamage = true
		case models.SpellEffectTypeDealDamageAllEnemies:
			candidate.DirectDamage = true
			candidate.AreaDamage = true
		case models.SpellEffectTypeRestoreLifePartyMember, models.SpellEffectTypeRestoreLifeAllParty,
			models.SpellEffectTypeRevivePartyMember, models.SpellEffectTypeReviveAllDownedParty:
			candidate.Healing = true
		case models.SpellEffectTypeApplyBeneficialStatus:
			candidate.BeneficialStatus = true
		case models.SpellEffectTypeApplyDetrimentalStatus, models.SpellEffectTypeApplyDetrimentalAll:
			candidate.DetrimentalStatus = true
		}

		if effect.DamageAffinity != nil && strings.TrimSpace(*effect.DamageAffinity) != "" {
			affinity := models.NormalizeDamageAffinity(strings.TrimSpace(*effect.DamageAffinity))
			candidate.AffinityCounts[affinity]++
		} else if candidate.AbilityType == models.SpellAbilityTypeTechnique && candidate.DirectDamage {
			candidate.AffinityCounts[models.DamageAffinityPhysical]++
		}
	}
}

func chooseProgressionsForMonsterTemplate(
	template *models.MonsterTemplate,
	candidates []monsterTemplateProgressionCandidate,
) []models.MonsterTemplateProgression {
	if template == nil || len(candidates) == 0 {
		return nil
	}

	targetCount := preferredAbilityCountForTemplate(template)
	if targetCount < 1 {
		targetCount = 1
	}

	selected := make([]models.MonsterTemplateProgression, 0, targetCount)
	used := make(map[uuid.UUID]struct{}, targetCount)
	for slot := 0; slot < targetCount; slot++ {
		desiredType := preferredAbilityTypeForTemplate(template, slot)
		bestIndex := -1
		bestScore := -1 << 30
		for index, candidate := range candidates {
			if _, exists := used[candidate.ProgressionID]; exists {
				continue
			}
			score := scoreMonsterTemplateProgressionCandidate(template, candidate, desiredType, slot)
			if score > bestScore {
				bestScore = score
				bestIndex = index
				continue
			}
			if score == bestScore && bestIndex >= 0 {
				left := strings.ToLower(strings.TrimSpace(candidate.Name))
				right := strings.ToLower(strings.TrimSpace(candidates[bestIndex].Name))
				if left < right {
					bestIndex = index
				}
			}
		}
		if bestIndex < 0 {
			break
		}
		chosen := candidates[bestIndex]
		used[chosen.ProgressionID] = struct{}{}
		selected = append(selected, models.MonsterTemplateProgression{
			ProgressionID: chosen.ProgressionID,
		})
	}
	return selected
}

func scoreMonsterTemplateProgressionCandidate(
	template *models.MonsterTemplate,
	candidate monsterTemplateProgressionCandidate,
	desiredType models.SpellAbilityType,
	slot int,
) int {
	score := 0
	if candidate.AbilityType == desiredType {
		score += 90
	} else {
		score -= 20
	}

	physical, mental := monsterTemplateCombatStyle(template)
	if candidate.AbilityType == models.SpellAbilityTypeTechnique {
		score += physical*2 - mental/2
		if candidate.DirectDamage {
			score += 18
		}
	} else {
		score += mental*2 - physical/2
		if candidate.DetrimentalStatus {
			score += 12
		}
		if candidate.Healing || candidate.BeneficialStatus {
			score += 6
		}
	}

	if candidate.AreaDamage {
		switch models.NormalizeMonsterTemplateType(string(template.MonsterType)) {
		case models.MonsterTemplateTypeRaid:
			score += 18
		case models.MonsterTemplateTypeBoss:
			score += 10
		default:
			score += 3
		}
	}
	if candidate.MemberCount > 0 {
		score += minInt(candidate.MemberCount, 6)
	}

	preferredAffinities := preferredAffinitiesForMonsterTemplate(template)
	for index, affinity := range preferredAffinities {
		if count := candidate.AffinityCounts[affinity]; count > 0 {
			score += 18 - (index * 4)
			score += minInt(count, 3) * 2
		}
	}

	if candidate.Healing && slot == 0 {
		score -= 8
	}
	return score
}

func preferredAffinitiesForMonsterTemplate(template *models.MonsterTemplate) []models.DamageAffinity {
	type affinityPreference struct {
		affinity models.DamageAffinity
		score    int
	}
	preferences := make([]affinityPreference, 0, len(models.DamageAffinities))
	for _, affinity := range models.DamageAffinities {
		score := affinityPreferenceScore(template, affinity)
		preferences = append(preferences, affinityPreference{affinity: affinity, score: score})
	}
	sort.Slice(preferences, func(i, j int) bool {
		if preferences[i].score != preferences[j].score {
			return preferences[i].score > preferences[j].score
		}
		return preferences[i].affinity < preferences[j].affinity
	})

	result := make([]models.DamageAffinity, 0, len(preferences))
	for _, preference := range preferences {
		result = append(result, preference.affinity)
	}
	return result
}

func affinityPreferenceScore(template *models.MonsterTemplate, affinity models.DamageAffinity) int {
	if template == nil {
		return 0
	}
	score := 0
	switch affinity {
	case models.DamageAffinityPhysical:
		score = absInt(template.PhysicalDamageBonusPercent) + absInt(template.PhysicalResistancePercent)
	case models.DamageAffinityPiercing:
		score = absInt(template.PiercingDamageBonusPercent) + absInt(template.PiercingResistancePercent)
	case models.DamageAffinitySlashing:
		score = absInt(template.SlashingDamageBonusPercent) + absInt(template.SlashingResistancePercent)
	case models.DamageAffinityBludgeoning:
		score = absInt(template.BludgeoningDamageBonusPercent) + absInt(template.BludgeoningResistancePercent)
	case models.DamageAffinityFire:
		score = absInt(template.FireDamageBonusPercent) + absInt(template.FireResistancePercent)
	case models.DamageAffinityIce:
		score = absInt(template.IceDamageBonusPercent) + absInt(template.IceResistancePercent)
	case models.DamageAffinityLightning:
		score = absInt(template.LightningDamageBonusPercent) + absInt(template.LightningResistancePercent)
	case models.DamageAffinityPoison:
		score = absInt(template.PoisonDamageBonusPercent) + absInt(template.PoisonResistancePercent)
	case models.DamageAffinityArcane:
		score = absInt(template.ArcaneDamageBonusPercent) + absInt(template.ArcaneResistancePercent)
	case models.DamageAffinityHoly:
		score = absInt(template.HolyDamageBonusPercent) + absInt(template.HolyResistancePercent)
	case models.DamageAffinityShadow:
		score = absInt(template.ShadowDamageBonusPercent) + absInt(template.ShadowResistancePercent)
	}

	physical, mental := monsterTemplateCombatStyle(template)
	if score == 0 {
		if mental > physical {
			switch affinity {
			case models.DamageAffinityArcane, models.DamageAffinityShadow, models.DamageAffinityFire, models.DamageAffinityIce:
				score = 6
			case models.DamageAffinityLightning, models.DamageAffinityHoly, models.DamageAffinityPoison:
				score = 4
			}
		} else {
			switch affinity {
			case models.DamageAffinityPhysical, models.DamageAffinitySlashing, models.DamageAffinityPiercing, models.DamageAffinityBludgeoning:
				score = 6
			case models.DamageAffinityPoison:
				score = 3
			}
		}
	}
	return score
}

func monsterTemplateCombatStyle(template *models.MonsterTemplate) (physical int, mental int) {
	if template == nil {
		return 0, 0
	}
	physical = template.BaseStrength + template.BaseDexterity + (template.BaseConstitution / 2)
	mental = template.BaseIntelligence + template.BaseWisdom + (template.BaseCharisma / 2)
	return physical, mental
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
